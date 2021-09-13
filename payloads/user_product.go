package payloads

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dhurimkelmendi/vending_machine/models"
	uuid "github.com/satori/go.uuid"
)

// UserProductBuyList is a struct that represents a list of bought products by users
type UserProductBuyList struct {
	UserProductList []*UserProductPurchase `json:"users_products"`
}

// UserChange is a struct that represents the change for a user broken down in coins of 5,10,20,100 cents
type UserChange struct {
	HundredCentCoins int32 `json:"hundred_cent_coins"`
	FiftyCentCoins   int32 `json:"fifty_cent_coins"`
	TwentyCentCoins  int32 `json:"twenty_cent_coins"`
	TenCentCoins     int32 `json:"ten_cent_coins"`
	FiveCentCoins    int32 `json:"five_cent_coins"`
}

// Equals compares two instances of type UserChange
func (p *UserChange) Equals(secondProduct *UserChange) bool {
	if p.HundredCentCoins != secondProduct.HundredCentCoins {
		return false
	}
	if p.FiftyCentCoins != secondProduct.FiftyCentCoins {
		return false
	}
	if p.TwentyCentCoins != secondProduct.TwentyCentCoins {
		return false
	}
	if p.TenCentCoins != secondProduct.TenCentCoins {
		return false
	}
	if p.FiveCentCoins != secondProduct.FiveCentCoins {
		return false
	}
	return true
}

// CreateChangeRepresentation returns a UserChange instance from a given int32
func CreateChangeRepresentation(change int32) *UserChange {
	userChange := &UserChange{}
	if change < 0 {
		return userChange
	}

	changeString := strconv.Itoa(int(change))
	hundreds, err := strconv.Atoi(changeString[0 : len(changeString)-2])
	if err != nil {
		userChange.HundredCentCoins = 0
	}
	userChange.HundredCentCoins = int32(hundreds)
	tens, err := strconv.Atoi(changeString[len(changeString)-2 : len(changeString)-1])
	if err != nil {
		userChange.TenCentCoins = 0
		userChange.FiftyCentCoins = 0
		userChange.TwentyCentCoins = 0
	}
	if tens/5 > 0 {
		userChange.FiftyCentCoins = int32(tens / 5)
		tens -= 5
	}
	if tens/2 > 0 {
		userChange.TwentyCentCoins = int32(tens / 2)
		tens -= 2
	}
	if tens > 0 {
		userChange.TenCentCoins = int32(tens)
	}

	ones, err := strconv.Atoi(string(changeString[len(changeString)-1:]))
	userChange.FiveCentCoins = int32(ones / 5)
	return userChange
}

// UserBuysReport is a struct that represents the products bought from a user
type UserBuysReport struct {
	UserID      uuid.UUID         `json:"user_id"`
	AmountSpent int32             `json:"amount_spent"`
	Change      UserChange        `json:"change"`
	Products    []*models.Product `json:"products"`
}

// UserProductPurchase is a struct that represents the payload for linking a single product to a user
type UserProductPurchase struct {
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	Amount    int32     `json:"amount"`
}

// Validate ensures that all the required fields are present in an instance of *UserProductBuy
func (p *UserProductPurchase) Validate() error {
	if p == nil {
		return fmt.Errorf("request body cannot be null")
	}
	if p.ProductID == uuid.Nil {
		return fmt.Errorf("product_id cannot be null")
	}
	if p.UserID == uuid.Nil {
		return fmt.Errorf("user_id cannot be null")
	}
	if p.Amount == 0 {
		return fmt.Errorf("amount cannot be null")
	}

	return nil
}

// Render is used by go-chi/renderer
func (p *UserProductPurchase) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
