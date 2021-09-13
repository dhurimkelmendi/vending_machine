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

// GetUserBuysReport returns all userProducts related to a given user, with the amount spent and change(if any)
func (s *UserProductService) GetUserBuysReport(userID uuid.UUID) (*payloads.UserBuysReport, error) {
	return s.getUserBuysReport(userID)
}
func (s *UserProductService) getUserBuysReport(userID uuid.UUID) (*payloads.UserBuysReport, error) {
	userReport := &payloads.UserBuysReport{UserID: userID}
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
		return &payloads.UserBuysReport{}, err
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
func (s *UserProductService) CreateUserProduct(ctx context.Context, createUserProduct *payloads.UserProductPurchase) (*payloads.UserBuysReport, error) {
	var err error
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		_, err = s.createUserProduct(tx, createUserProduct)
		return err
	})
	userReport := &payloads.UserBuysReport{UserID: createUserProduct.UserID}

	return userReport, err
}
func (s *UserProductService) createUserProduct(dbSession *pg.Tx, createUserProduct *payloads.UserProductPurchase) (*models.UsersProduct, error) {
	createdUserProduct := &models.UsersProduct{
		UserID:    createUserProduct.UserID,
		ProductID: createUserProduct.ProductID,
		Amount:    createUserProduct.Amount,
	}
	if err := createUserProduct.Validate(); err != nil {
		return createdUserProduct, err
	}
	_, err := dbSession.Model(createUserProduct).Insert()
	if err != nil {
		return createdUserProduct, err
	}
	return createdUserProduct, nil
}
