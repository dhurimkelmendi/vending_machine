package models

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// Product is a struct that represents a db row of the Products table
type Product struct {
	tableName       struct{}  `pg:"products"`
	ID              uuid.UUID `pg:"id,pk,type:uuid"`
	SellerID        uuid.UUID `json:"seller_id" pg:"seller_id,fk,type:uuid"`
	Name            string    `json:"name"`
	AmountAvailable int32     `json:"amount_available"`
	Cost            int32     `json:"cost"`
}

// Merge merges two instances of type Product into one
func (p *Product) Merge(secondProduct Product) {
	if p.SellerID == uuid.Nil {
		p.SellerID = secondProduct.SellerID
	}
	if p.Name == "" {
		p.Name = secondProduct.Name
	}
	if p.AmountAvailable == 0 {
		p.AmountAvailable = secondProduct.AmountAvailable
	}
	if p.Cost == 0 {
		p.Cost = secondProduct.Cost
	}
}

// Equals compares two instances of type Product
func (p *Product) Equals(secondProduct *Product) bool {
	if p.ID != secondProduct.ID {
		return false
	}
	if p.Name != secondProduct.Name {
		return false
	}
	if p.AmountAvailable != secondProduct.AmountAvailable {
		return false
	}
	if p.Cost != secondProduct.Cost {
		return false
	}
	return true
}

// Render is used by go-chi/renderer
func (p *Product) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
