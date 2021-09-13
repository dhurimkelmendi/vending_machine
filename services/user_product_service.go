package services

import (
	"context"

	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/go-pg/pg/v10"
	uuid "github.com/satori/go.uuid"
)

// UserProductService is a struct that contains references to the db and the StatelessAuthenticationProvider
type UserProductService struct {
	db        *pg.DB
	stateless *auth.StatelessAuthenticationProvider
}

var userProductServiceDefaultInstance *UserProductService

// GetUserProductServiceDefaultInstance returns the default instance of UserProductService
func GetUserProductServiceDefaultInstance() *UserProductService {
	if userProductServiceDefaultInstance == nil {

		userProductServiceDefaultInstance = &UserProductService{
			db:        db.GetDefaultInstance().GetDB(),
			stateless: auth.GetStatelessAuthenticationProviderDefaultInstance(),
		}
	}
	return userProductServiceDefaultInstance
}

// GetUserPurchasesReport returns all userProducts related to a given user
func (s *UserProductService) GetUserPurchasesReport(userID uuid.UUID) (*payloads.UserPurchasesReport, error) {
	return s.getUserPurchasesReport(userID)
}
func (s *UserProductService) getUserPurchasesReport(userID uuid.UUID) (*payloads.UserPurchasesReport, error) {
	userReport := &payloads.UserPurchasesReport{UserID: userID}
	user := &models.User{ID: userID}
	if err := s.db.Model(user).
		WherePK().
		Relation("Products").
		Select(); err != nil {
		switch err {
		case pg.ErrNoRows:
			return userReport, db.ErrNoMatch
		default:
			return userReport, err
		}
	}
	productAmounts := make(map[string]int32, len(userReport.Products))
	userProducts, err := s.getAllUserProductsForUser(userID)
	if err != nil {
		return &payloads.UserPurchasesReport{}, err
	}

	for _, userProduct := range userProducts {
		productAmounts[userProduct.ProductID.String()] = userProduct.Amount
	}

	for _, product := range user.Products {
		userReport.AmountSpent += product.Cost * productAmounts[product.ID.String()]
	}
	userReport.Products = user.Products
	userChange := user.Deposit - userReport.AmountSpent
	userReport.Change = *payloads.CreateChangeRepresentation(userChange)

	return userReport, nil
}

func (s *UserProductService) getAllUserProductsForUser(UserID uuid.UUID) ([]models.UsersProduct, error) {
	UserProducts := make([]models.UsersProduct, 0)
	err := s.db.Model(&UserProducts).Where("User_id = ?", UserID).Select()
	if err != nil {
		return UserProducts, err
	}
	return UserProducts, nil
}

// CreateUserProduct creates a userProduct using the provided payload
func (s *UserProductService) CreateUserProduct(ctx context.Context, createUserProduct *models.UsersProduct) (*models.UsersProduct, error) {
	createdUserProduct := &models.UsersProduct{}
	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		createdUserProduct, err = s.createUserProduct(tx, createUserProduct)
		return err
	})
	return createdUserProduct, err
}
func (s *UserProductService) createUserProduct(dbSession *pg.Tx, createUserProduct *models.UsersProduct) (*models.UsersProduct, error) {
	if err := createUserProduct.Validate(); err != nil {
		return createUserProduct, err
	}
	_, err := dbSession.Model(createUserProduct).Insert()
	if err != nil {
		return createUserProduct, err
	}
	return createUserProduct, nil
}
