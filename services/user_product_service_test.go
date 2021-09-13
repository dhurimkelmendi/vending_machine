package services_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/fixtures"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
)

func TestUserProductService(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()
	service := services.GetUserProductServiceDefaultInstance()
	seller := fixture.User.CreateSellerUser(t)
	buyer := fixture.User.CreateBuyerUser(t)
	product := fixture.Product.CreateProduct(t, seller.ID)
	userProduct := fixture.UserProduct.CreateUserProduct(t, product.ID, buyer.ID)
	ctx := context.Background()

	t.Run("create userProduct", func(t *testing.T) {
		t.Run("create userProduct with all fields", func(t *testing.T) {
			userProductToCreate := &models.UsersProduct{}
			userProductToCreate.ProductID = product.ID
			userProductToCreate.UserID = seller.ID
			userProductToCreate.Amount = int32(gofakeit.Uint16())
			createdUserProduct, err := service.CreateUserProduct(ctx, userProductToCreate)
			if err != nil {
				t.Fatalf("error while creating userProduct %+v", err)
			}
			userProductToCreate.ID = createdUserProduct.ID
			if !userProductToCreate.Equals(createdUserProduct) {
				t.Fatalf("create userProduct failed: %+v \n received: %+v, %+v", userProductToCreate, createdUserProduct, err)
			}
		})
		t.Run("create userProduct without product_id", func(t *testing.T) {
			userProductToCreate := &models.UsersProduct{}
			userProductToCreate.UserID = seller.ID
			userProductToCreate.Amount = int32(gofakeit.Uint16())
			_, err := service.CreateUserProduct(ctx, userProductToCreate)
			if err == nil {
				t.Fatal("expected create to fail without product_id, create was allowed")
			}
		})
		t.Run("create userProduct without user_id", func(t *testing.T) {
			userProductToCreate := &models.UsersProduct{}
			userProductToCreate.ProductID = product.ID
			userProductToCreate.Amount = int32(gofakeit.Uint16())
			_, err := service.CreateUserProduct(ctx, userProductToCreate)
			if err == nil {
				t.Fatal("expected create to fail without user_id, create was allowed")
			}
		})
		t.Run("create userProduct without amount", func(t *testing.T) {
			userProductToCreate := &models.UsersProduct{}
			userProductToCreate.ProductID = product.ID
			userProductToCreate.UserID = seller.ID
			_, err := service.CreateUserProduct(ctx, userProductToCreate)
			if err == nil {
				t.Fatal("expected create to fail without amount, create was allowed")
			}
		})
	})

	t.Run("get user report", func(t *testing.T) {
		userReport, err := service.GetUserPurchasesReport(buyer.ID)
		if err != nil {
			t.Fatalf("generate user report failed: %+v", err)
		}
		if userReport.UserID != buyer.ID {
			t.Fatalf("user report generated for wrong user, expected: %s, got %s", buyer.ID, userReport.UserID)
		}
		totalSpendExpected := product.Cost * userProduct.Amount
		if userReport.AmountSpent != totalSpendExpected {
			t.Fatalf("user report generated wrong amount_spent, expected: %d, got %d", totalSpendExpected, userReport.AmountSpent)
		}
		productIsInList := false
		for _, p := range userReport.Products {
			if p.Equals(product) {
				productIsInList = true
			}
		}
		if !productIsInList {
			t.Fatalf("user report generated wrong products list, expected it to contain: %+v, got %+v", product, userReport.Products)
		}
		change := buyer.Deposit - totalSpendExpected
		expectedReportChange := payloads.CreateChangeRepresentation(change)
		if !expectedReportChange.Equals(&userReport.Change) {
			t.Fatalf("user report generated wrong change report, expected it to contain: %+v, got %+v", expectedReportChange, userReport.Change)
		}
	})
}
