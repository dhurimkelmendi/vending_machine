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

// ProductService is a struct that contains references to the db and the StatelessAuthenticationProvider
type ProductService struct {
	db        *pg.DB
	stateless *auth.StatelessAuthenticationProvider
}

var productServiceDefaultInstance *ProductService

// GetProductServiceDefaultInstance returns the default instance of ProductService
func GetProductServiceDefaultInstance() *ProductService {
	if productServiceDefaultInstance == nil {
		productServiceDefaultInstance = &ProductService{
			db:        db.GetDefaultInstance().GetDB(),
			stateless: auth.GetStatelessAuthenticationProviderDefaultInstance(),
		}
	}

	return productServiceDefaultInstance
}

// GetAllProducts returns all products
func (s *ProductService) GetAllProducts() (*payloads.ProductList, error) {
	return s.getAllProducts()
}
func (s *ProductService) getAllProducts() (*payloads.ProductList, error) {
	products := make([]*models.Product, 0)

	err := s.db.Model(&products).Select()
	if err != nil {
		return nil, err
	}

	productList := &payloads.ProductList{}
	productList.Products = products

	return productList, nil
}

// GetProductByID returns the requested product by id
func (s *ProductService) GetProductByID(productID uuid.UUID) (*models.Product, error) {
	return s.getProductByID(productID)
}
func (s *ProductService) getProductByID(productID uuid.UUID) (*models.Product, error) {
	product := &models.Product{}
	switch err := s.db.Model(product).Where("id = ?", productID).Select(); err {
	case pg.ErrNoRows:
		return product, db.ErrNoMatch
	default:
		return product, err
	}
}

// CreateProduct creates a product using the provided payload
func (s *ProductService) CreateProduct(ctx context.Context, createProduct *payloads.CreateProductPayload) (*models.Product, error) {
	product := &models.Product{}
	if err := createProduct.Validate(); err != nil {
		return product, err
	}
	var err error
	err = s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		product, err = s.createProduct(tx, createProduct)
		return err
	})
	if err != nil {
		return product, err
	}
	return product, err
}
func (s *ProductService) createProduct(dbSession *pg.Tx, registerProduct *payloads.CreateProductPayload) (*models.Product, error) {
	product := registerProduct.ToProductModel()

	product.ID = uuid.NewV4()
	_, err := dbSession.Model(product).Insert()
	if err != nil {
		return product, err
	}

	return product, nil
}

// UpdateProduct updates the product by id using the provided payload
func (s *ProductService) UpdateProduct(ctx context.Context, updateProduct *payloads.UpdateProductPayload, userContext auth.UserContext) (*models.Product, error) {
	var updatedProduct *models.Product
	existingProduct, err := s.GetProductByID(updateProduct.ID)
	if err != nil {
		return updatedProduct, db.ErrNoMatch
	}
	if userContext.ID != existingProduct.SellerID {
		return updatedProduct, db.ErrUserForbidden
	}
	s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		updatedProduct, err = s.updateProduct(tx, updateProduct)
		return err
	})

	return updatedProduct, err
}
func (s *ProductService) updateProduct(dbSession *pg.Tx, updateProduct *payloads.UpdateProductPayload) (*models.Product, error) {
	product := updateProduct.ToProductModel()
	existingProduct, err := s.GetProductByID(product.ID)
	if err != nil {
		return &models.Product{}, db.ErrNoMatch
	}

	product.Merge(*existingProduct)

	if _, err := dbSession.Model(product).Where("id = ?", product.ID).Update(); err != nil {
		if err == pg.ErrNoRows {
			return product, db.ErrNoMatch
		}
		return product, err
	}
	return product, nil
}

// DeleteProduct deletes the product by id
func (s *ProductService) DeleteProduct(ctx context.Context, productID uuid.UUID, userContext auth.UserContext) error {
	existingProduct, err := s.GetProductByID(productID)
	if err != nil {
		return db.ErrNoMatch
	}
	if userContext.ID != existingProduct.SellerID {
		return db.ErrUserForbidden
	}
	return s.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		return s.deleteProduct(tx, productID)
	})
}
func (s *ProductService) deleteProduct(dbSession *pg.Tx, productID uuid.UUID) error {
	product := &models.Product{ID: productID}

	result, err := dbSession.Model(product).WherePK().Delete()

	if err != nil {
		switch err {
		case pg.ErrNoRows:
			return db.ErrNoMatch
		default:
			return err
		}
	}

	if result.RowsAffected() == 0 {
		err = db.ErrNoMatch
	}

	return err
}
