package fixtures

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
	"github.com/go-pg/pg/v10"
	uuid "github.com/satori/go.uuid"
)

// ProductFixture is a struct that contains references to the db and ProductService
type ProductFixture struct {
	db             *pg.DB
	productService *services.ProductService
}

var productFixtureDefaultInstance *ProductFixture

// GetProductFixtureDefaultInstance returns the default instance of ProductFixture
func GetProductFixtureDefaultInstance() *ProductFixture {
	if productFixtureDefaultInstance == nil {
		productFixtureDefaultInstance = &ProductFixture{
			db:             db.GetDefaultInstance().GetDB(),
			productService: services.GetProductServiceDefaultInstance(),
		}
	}

	if productFixtureDefaultInstance.productService == nil {
		productFixtureDefaultInstance.productService = services.GetProductServiceDefaultInstance()
	}

	return productFixtureDefaultInstance
}

// CreateProduct creates a product with fake data
func (f *ProductFixture) CreateProduct(t *testing.T, sellerID uuid.UUID) *models.Product {
	product := &payloads.CreateProductPayload{}
	product.Name = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
	product.SellerID = sellerID
	product.Cost = int32(gofakeit.Uint32())
	product.AmountAvailable = int32(gofakeit.Uint32())

	ctx := context.Background()

	if f.productService == nil {
		t.Log("CreateBuyerProduct: fixture.ProductService is nil!")
	}

	createdProduct, err := f.productService.CreateProduct(ctx, product)
	if err != nil {
		return nil
	}
	return createdProduct
}
