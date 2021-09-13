package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/fixtures"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
	uuid "github.com/satori/go.uuid"
)

func TestProductService(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()

	service := services.GetProductServiceDefaultInstance()
	buyer := fixture.User.CreateBuyerUser(t)
	seller := fixture.User.CreateSellerUser(t)
	product := fixture.Product.CreateProduct(t, seller.ID)
	sellerUserContext := auth.UserContext{
		ID:   seller.ID,
		Role: seller.Role,
	}
	ctx := context.Background()

	t.Run("create product", func(t *testing.T) {
		t.Run("create product with all fields", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.Name = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			productToCreate.SellerID = seller.ID
			productToCreate.Cost = int32(gofakeit.Uint32())
			productToCreate.AmountAvailable = int32(gofakeit.Uint32())
			createdProduct, err := service.CreateProduct(ctx, productToCreate)
			if err != nil {
				t.Fatalf("error while creating product %+v", err)
			}
			productToCreateModel := productToCreate.ToProductModel()
			productToCreateModel.ID = createdProduct.ID
			if !productToCreateModel.Equals(createdProduct) {
				t.Fatalf("create product failed: %+v \n received: %+v, %+v", productToCreateModel, createdProduct, err)
			}
		})
		t.Run("with existing name", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.Name = product.Name
			productToCreate.SellerID = seller.ID
			productToCreate.Cost = int32(gofakeit.Uint32())
			productToCreate.AmountAvailable = int32(gofakeit.Uint32())
			_, err := service.CreateProduct(ctx, productToCreate)
			if err == nil {
				t.Fatalf("expected duplicate product to fail %+v", err)
			}
		})
		t.Run("without name", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.SellerID = seller.ID
			productToCreate.Cost = int32(gofakeit.Uint32())
			productToCreate.AmountAvailable = int32(gofakeit.Uint32())
			_, err := service.CreateProduct(ctx, productToCreate)
			if err == nil {
				t.Fatalf("expected create product to fail without name, update was allowed, %+v", err)
			}
		})
		t.Run("without seller_id", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.Name = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			productToCreate.Cost = int32(gofakeit.Uint32())
			productToCreate.AmountAvailable = int32(gofakeit.Uint32())
			_, err := service.CreateProduct(ctx, productToCreate)
			if err == nil {
				t.Fatalf("expected create product to fail without seller_id, update was allowed, %+v", err)
			}
		})
		t.Run("without cost", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.Name = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			productToCreate.SellerID = seller.ID
			productToCreate.AmountAvailable = int32(gofakeit.Uint32())
			_, err := service.CreateProduct(ctx, productToCreate)
			if err == nil {
				t.Fatalf("expected create product to fail without cost, update was allowed, %+v", err)
			}
		})
		t.Run("without amount_available", func(t *testing.T) {
			productToCreate := &payloads.CreateProductPayload{}
			productToCreate.Name = strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			productToCreate.Cost = int32(gofakeit.Uint32())
			productToCreate.SellerID = seller.ID
			_, err := service.CreateProduct(ctx, productToCreate)
			if err == nil {
				t.Fatalf("expected create product to fail without amount_available, update was allowed, %+v", err)
			}
		})
	})

	t.Run("get product by id", func(t *testing.T) {
		_, err := service.GetProductByID(product.ID)
		if err != nil {
			t.Fatalf("could not retreive existing product by ID: %d, %+v", product.ID, err)
		}
	})

	t.Run("get all products", func(t *testing.T) {
		_, err := service.GetAllProducts()
		if err != nil {
			t.Fatalf("could not retreive products: %+v", err)
		}
	})

	t.Run("update product", func(t *testing.T) {
		t.Run("with basic attributes", func(t *testing.T) {
			productToUpdate := &payloads.UpdateProductPayload{}
			productToUpdate.ID = product.ID
			newCost := int32(gofakeit.Uint32())
			productToUpdate.Cost = newCost
			updatedProduct, err := service.UpdateProduct(ctx, productToUpdate, sellerUserContext)
			if err != nil {
				t.Fatalf("update product failed: %+v", err)
			}
			if updatedProduct.Cost != newCost {
				t.Fatalf("expected cost to be %d, got: %+v", newCost, updatedProduct.Cost)
			}
		})
		t.Run("with protected attributes", func(t *testing.T) {
			productToUpdate := &payloads.UpdateProductPayload{}
			newID := uuid.NewV4()
			productToUpdate.ID = newID
			newCost := int32(gofakeit.Uint32())
			productToUpdate.Cost = newCost
			_, err := service.UpdateProduct(ctx, productToUpdate, sellerUserContext)
			if err == nil {
				t.Fatal("expected id not to be updated, update was allowed")
			}
		})
	})

	t.Run("delete product", func(t *testing.T) {
		t.Run("requested by owner", func(t *testing.T) {
			err := service.DeleteProduct(ctx, product.ID, sellerUserContext)
			if err != nil {
				t.Fatalf("delete product failed: %+v", err)
			}
		})
		t.Run("requested by buyer", func(t *testing.T) {
			buyerUserContext := auth.UserContext{
				ID:   buyer.ID,
				Role: buyer.Role,
			}
			err := service.DeleteProduct(ctx, product.ID, buyerUserContext)
			if err == nil {
				t.Fatalf("expected delete not to be allowed for non-owners, delete was allowed: %+v", err)
			}
		})
	})

}
