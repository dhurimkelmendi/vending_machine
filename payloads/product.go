package payloads

import (
	"fmt"
	"net/http"

	"github.com/dhurimkelmendi/vending_machine/models"
)

// ProductList is a struct that contains a reference to a slice of type *models.Product
type ProductList struct {
	Products []*models.Product `json:"products"`
}

// Render is used by go-chi/renderer
func (pl *ProductList) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// CreateProductPayload for registering a new product
type CreateProductPayload struct {
	tableName       struct{} `pg:"products"`
	Name            string   `json:"name"`
	AmountAvailable int32    `json:"amount_available"`
	Cost            int32    `json:"cost"`
}

// ToProductModel converts an instance of type *RegisterProductPayload to *models.Product type
func (p *CreateProductPayload) ToProductModel() *models.Product {
	return &models.Product{
		Name:            p.Name,
		AmountAvailable: p.AmountAvailable,
		Cost:            p.Cost,
	}
}

// Validate ensures that all the required fields are present in an instance of *RegisterProductPayload
func (p *CreateProductPayload) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is a required field")
	}
	if p.AmountAvailable == 0 {
		return fmt.Errorf("amount_available is a required field")
	}
	if p.Cost == 0 {
		return fmt.Errorf("cost is a required field")
	}

	return nil
}

// Render is used by go-chi/renderer
func (p *CreateProductPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// UpdateProductPayload is a struct that represents the payload that is expected when updating a product
type UpdateProductPayload struct {
	tableName struct{} `pg:"products"`
	models.Product
}

// ToProductModel converts an instance of type *UpdateProductPayload to *models.Product type
func (p *UpdateProductPayload) ToProductModel() *models.Product {
	return &models.Product{
		ID:              p.ID,
		Name:            p.Name,
		SellerID:        p.SellerID,
		AmountAvailable: p.AmountAvailable,
		Cost:            p.Cost,
	}
}

// Validate ensures that all the required fields are present in an instance of *UpdateProductPayload
func (p *UpdateProductPayload) Validate() error {
	if p == nil {
		return fmt.Errorf("request body cannot be null")
	}
	return nil
}

// Render is used by go-chi/renderer
func (p *UpdateProductPayload) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
