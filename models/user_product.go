package models

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// UsersProduct is a struct that represents the many-to-many relationship between a user and an product
type UsersProduct struct {
	tableName struct{}  `pg:"users_products"`
	ID        uuid.UUID `json:"id" pg:"id,pk"`
	UserID    uuid.UUID `json:"user_id" pg:"user_id"`
	ProductID uuid.UUID `json:"product_id" pg:"product_id"`
	Amount    int32     `json:"amount" pg:"amount"`
}

// Merge merges two instances of type UserProduct into one
func (u *UsersProduct) Merge(secondProduct UsersProduct) {
	if u.UserID == uuid.Nil {
		u.UserID = secondProduct.UserID
	}
	if u.ProductID == uuid.Nil {
		u.ProductID = secondProduct.ProductID
	}
	if u.Amount == 0 {
		u.Amount = secondProduct.Amount
	}
}

// Equals compares two instances of type UserProduct
func (u *UsersProduct) Equals(secondProduct *UsersProduct) bool {
	if u.ID != secondProduct.ID {
		return false
	}
	if u.UserID != secondProduct.UserID {
		return false
	}
	if u.ProductID != secondProduct.ProductID {
		return false
	}
	if u.Amount != secondProduct.Amount {
		return false
	}
	return true
}

// Validate ensures that all the required fields are present in an instance of *CreateProductPayload
func (u *UsersProduct) Validate() error {
	if u.ProductID == uuid.Nil {
		return fmt.Errorf("product_id is a required field")
	}
	if u.UserID == uuid.Nil {
		return fmt.Errorf("user_id is a required field")
	}
	if u.Amount == 0 {
		return fmt.Errorf("amount is a required field")
	}
	return nil
}

// Render is used by go-chi/renderer
func (u *UsersProduct) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
