package services

import (
	"context"
	"fmt"

	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-pg/pg/v10"
	uuid "github.com/satori/go.uuid"
)

// UserService is a struct that contains references to the db and the StatelessAuthenticationProvider
type UserService struct {
	db                 *pg.DB
	stateless          *auth.StatelessAuthenticationProvider
	userProductService *UserProductService
	productService     *ProductService
}

var userServiceDefaultInstance *UserService

// GetUserServiceDefaultInstance returns the default instance of UserService
func GetUserServiceDefaultInstance() *UserService {
	if userServiceDefaultInstance == nil {
		userServiceDefaultInstance = &UserService{
			db:                 db.GetDefaultInstance().GetDB(),
			stateless:          auth.GetStatelessAuthenticationProviderDefaultInstance(),
			userProductService: GetUserProductServiceDefaultInstance(),
			productService:     GetProductServiceDefaultInstance(),
		}
	}

	return userServiceDefaultInstance
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers() (*payloads.UserList, error) {
	return s.getAllUsers()
}
func (s *UserService) getAllUsers() (*payloads.UserList, error) {
	users := make([]*models.User, 0)

	err := s.db.Model(&users).Select()
	if err != nil {
		return nil, err
	}

	userList := &payloads.UserList{}
	userList.Users = make([]*payloads.UserDetails, len(users))

	for i, user := range users {
		userList.Users[i] = payloads.MapUserToUserDetails(user)
	}

	return userList, nil
}

// GetUserByID returns the requested user by id
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	return s.getUserByID(userID)
}
func (s *UserService) getUserByID(userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	switch err := s.db.Model(user).Where("id = ?", userID).Select(); err {
	case pg.ErrNoRows:
		return user, db.ErrNoMatch
	default:
		return user, err
	}
}

// GetUserByUsername returns the requested user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.getUserByUsername(username)
}
func (s *UserService) getUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	switch err := s.db.Model(user).Where("username = ?", username).Select(); err {
	case pg.ErrNoRows:
		return user, db.ErrNoMatch
	default:
		return user, err
	}
}

// CreateUser creates a user using the provided payload
func (s *UserService) CreateUser(ctx context.Context, createUser *payloads.CreateUserPayload) (*models.User, error) {
	user := &models.User{}
	if err := createUser.Validate(); err != nil {
		return user, err
	}

	var err error
	err = s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		user, err = s.createUser(tx, createUser)
		return err
	})
	if err != nil {
		return user, err
	}
	return user, err
}
func (s *UserService) createUser(dbSession *pg.Tx, createUser *payloads.CreateUserPayload) (*models.User, error) {
	user := createUser.ToUserModel()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return &models.User{}, fmt.Errorf("error while hashing password")
	}
	user.Password = string(hashedPassword)
	user.ID = uuid.NewV4()
	_, err = dbSession.Model(user).Insert()
	if err != nil {
		return user, err
	}

	// We need the user to be created (for their id) before we can create their auth token
	user.Token, err = s.stateless.CreateUserAuthToken(user)
	if err != nil {
		return user, err
	}

	if _, err := dbSession.Model(user).Where("id = ?", user.ID).Update(); err != nil {
		return user, err
	}

	return user, nil
}

// LoginUser creates a user using the provided payload
func (s *UserService) LoginUser(ctx context.Context, loginUser *payloads.LoginUserPayload) (*models.User, error) {
	var updatedUser *models.User
	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		updatedUser, err = s.loginUser(tx, loginUser)
		return err
	})

	return updatedUser, err
}
func (s *UserService) loginUser(dbSession *pg.Tx, loginUser *payloads.LoginUserPayload) (*models.User, error) {
	user, err := s.getUserByUsername(loginUser.Username)
	if err != nil {
		return &models.User{}, fmt.Errorf("incorrect username or password")
	}
	hashPasswordErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if user.Username != loginUser.Username || hashPasswordErr != nil {
		return &models.User{}, fmt.Errorf("incorrect username or password")
	}
	return user, nil
}

// UpdateUser updates the user by id using the provided payload
func (s *UserService) UpdateUser(ctx context.Context, updateUser *payloads.UpdateUserPayload) (*models.User, error) {
	var updatedUser *models.User
	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		updatedUser, err = s.updateUser(tx, updateUser)
		return err
	})

	return updatedUser, err
}
func (s *UserService) updateUser(dbSession *pg.Tx, updateUser *payloads.UpdateUserPayload) (*models.User, error) {
	user := updateUser.ToUserModel()
	existingUser, err := s.GetUserByID(user.ID)
	if err != nil {
		return &models.User{}, db.ErrNoMatch
	}

	user.Merge(*existingUser)

	if _, err := dbSession.Model(user).Where("id = ?", user.ID).Update(); err != nil {
		if err == pg.ErrNoRows {
			return user, db.ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

// DepositMoney updates the user deposit by adding the specified amount
func (s *UserService) DepositMoney(ctx context.Context, depositMoney *payloads.DepositMoneyPayload, userID uuid.UUID) (*models.User, error) {
	var updatedUser *models.User
	if err := depositMoney.Validate(); err != nil {
		return &models.User{}, err
	}
	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		updatedUser, err = s.depositMoney(tx, depositMoney, userID)
		return err
	})

	return updatedUser, err
}
func (s *UserService) depositMoney(dbSession *pg.Tx, depositMoney *payloads.DepositMoneyPayload, userID uuid.UUID) (*models.User, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return &models.User{}, db.ErrNoMatch
	}
	if user.Role != models.UserRoleBuyer {
		return &models.User{}, db.ErrUserForbidden
	}
	user.Deposit += depositMoney.DepositAmount
	if _, err := dbSession.Model(user).Where("id = ?", user.ID).Update(); err != nil {
		if err == pg.ErrNoRows {
			return user, db.ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

// ResetDeposit resets the user deposit
func (s *UserService) ResetDeposit(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var updatedUser *models.User

	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		updatedUser, err = s.resetDeposit(tx, userID)
		return err
	})

	return updatedUser, err
}
func (s *UserService) resetDeposit(dbSession *pg.Tx, userID uuid.UUID) (*models.User, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return &models.User{}, db.ErrNoMatch
	}
	user.Deposit = 0
	if _, err := dbSession.Model(user).Where("id = ?", user.ID).Update(); err != nil {
		if err == pg.ErrNoRows {
			return user, db.ErrNoMatch
		}
		return user, err
	}
	return user, nil
}

// DeleteUser deletes the user by id
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return s.deleteUser(tx, userID)
	})
}
func (s *UserService) deleteUser(dbSession *pg.Tx, userID uuid.UUID) error {
	user := &models.User{ID: userID}

	result, err := dbSession.Model(user).WherePK().Delete()

	if err != nil {
		switch err {
		case pg.ErrNoRows:
			return db.ErrNoMatch
		default:
			return err
		}
	}

	if result.RowsAffected() == 0 {
		err = db.ErrNoMatch
	}

	return err
}

// BuyProduct links a product to the given user
func (s *UserService) BuyProduct(ctx context.Context, createUserProduct *payloads.UserProductPurchase) (*payloads.UserBuysReport, error) {
	var userReport *payloads.UserBuysReport

	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		userReport, err = s.buyProduct(ctx, tx, createUserProduct)
		return err
	})

	return userReport, err
}
func (s *UserService) buyProduct(ctx context.Context, dbSession *pg.Tx, createUserProduct *payloads.UserProductPurchase) (*payloads.UserBuysReport, error) {
	userReport := &payloads.UserBuysReport{}

	user := &models.User{ID: createUserProduct.UserID}
	user, err := s.GetUserByID(user.ID)
	if err != nil {
		return userReport, db.ErrNoMatch
	}

	product, err := s.productService.GetProductByID(createUserProduct.ProductID)
	if err != nil {
		return userReport, db.ErrNoMatch
	}

	amountToBeSpent := product.Cost * createUserProduct.Amount
	if user.Deposit < amountToBeSpent {
		return userReport, fmt.Errorf("unable to buy product amount, deposit too low")
	}
	if _, err = s.userProductService.CreateUserProduct(ctx, createUserProduct); err != nil {
		return userReport, err
	}
	user.Deposit -= amountToBeSpent
	if _, err := dbSession.Model(user).Where("id = ?", user.ID).Update(); err != nil {
		if err == pg.ErrNoRows {
			return userReport, db.ErrNoMatch
		}
		return userReport, err
	}
	userReport, err = s.userProductService.GetUserBuysReport(user.ID)
	if err != nil {
		return userReport, err
	}
	return userReport, err
}
