package fixtures

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/services"
	"github.com/go-pg/pg/v10"
	uuid "github.com/satori/go.uuid"
)

// UserProductFixture is a struct that contains references to the db and UserProductService
type UserProductFixture struct {
	db                 *pg.DB
	userProductService *services.UserProductService
}

var userProductFixtureDefaultInstance *UserProductFixture

// GetUserProductFixtureDefaultInstance returns the default instance of UserProductFixture
func GetUserProductFixtureDefaultInstance() *UserProductFixture {
	if userProductFixtureDefaultInstance == nil {
		userProductFixtureDefaultInstance = &UserProductFixture{
			db:                 db.GetDefaultInstance().GetDB(),
			userProductService: services.GetUserProductServiceDefaultInstance(),
		}
	}
	return userProductFixtureDefaultInstance
}

// CreateUserProduct creates an userProduct with fake data
func (f *UserProductFixture) CreateUserProduct(t *testing.T, productID uuid.UUID, userID uuid.UUID) *models.UsersProduct {
	userProduct := &models.UsersProduct{}
	userProduct.ProductID = productID
	userProduct.UserID = userID
	userProduct.Amount = int32(gofakeit.Uint16())
	ctx := context.Background()
	createdUserProduct, err := f.userProductService.CreateUserProduct(ctx, userProduct)
	if err != nil {
		return nil
	}
	return createdUserProduct
}
