package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dhurimkelmendi/vending_machine/api"
	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/db"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/dhurimkelmendi/vending_machine/services"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	uuid "github.com/satori/go.uuid"
)

// A ProductsController handles HTTP requests that deal with product.
type ProductsController struct {
	AuthenticatedController
	productService *services.ProductService
}

var productsControllerDefaultInstance *ProductsController

// GetProductsControllerDefaultInstance returns the default instance of ProductController.
func GetProductsControllerDefaultInstance() *ProductsController {
	if productsControllerDefaultInstance == nil {
		productsControllerDefaultInstance = NewProductController(services.GetProductServiceDefaultInstance())
	}

	return productsControllerDefaultInstance
}

// NewProductController create a new instance of a product controller using the supplied product service
func NewProductController(productService *services.ProductService) *ProductsController {
	controller := Controller{
		errCmp:    api.NewErrorComponent(api.CmpController),
		responder: api.GetResponderDefaultInstance(),
	}
	authenticatedController := AuthenticatedController{
		Controller:                      controller,
		statelessAuthenticationProvider: auth.GetStatelessAuthenticationProviderDefaultInstance(),
	}

	return &ProductsController{
		AuthenticatedController: authenticatedController,
		productService:          productService,
	}
}

// GetAllProducts returns all active (non-deleted) products
func (c *ProductsController) GetAllProducts(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxGetProducts, r.Header.Get("X-Request-Id"))
	products, err := c.productService.GetAllProducts()
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrGetProducts, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, products); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot serialize result")), http.StatusBadRequest)
	}
}

// GetProductByID returns the requested product by id
func (c *ProductsController) GetProductByID(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxGetProduct, r.Header.Get("X-Request-Id"))
	urlProductID := chi.URLParam(r, "id")
	productID, err := uuid.FromString(urlProductID)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestParameter, fmt.Errorf("invalid productId, %v", err)), http.StatusBadRequest)
		return
	}

	product, err := c.productService.GetProductByID(productID)
	if err != nil {
		if err == db.ErrNoMatch {
			c.responder.Error(w, errCtx(api.ErrProductNotFound, errors.New("no product with that id")), http.StatusNotFound)
		} else {
			c.responder.Error(w, errCtx(api.ErrGetProduct, err), http.StatusBadRequest)
		}
		return
	}
	if err := render.Render(w, r, product); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot serialize result")), http.StatusBadRequest)
		return
	}
}

// CreateProduct creates a new product and returns product details with an authentication token
func (c *ProductsController) CreateProduct(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxCreateProduct, r.Header.Get("X-Request-Id"))
	product := &payloads.CreateProductPayload{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot decode product")), http.StatusBadRequest)
		return
	}

	if err := product.Validate(); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("request body not valid, missing required fields")), http.StatusBadRequest)
		return
	}

	createdProduct, err := c.productService.CreateProduct(context.Background(), product)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrCreateProduct, err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	c.responder.JSON(w, r, createdProduct, http.StatusCreated)
}

// UpdateProduct update the current products profile
func (c *ProductsController) UpdateProduct(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxUpdateProduct, r.Header.Get("X-Request-Id"))

	product := &payloads.UpdateProductPayload{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("cannot decode product")), http.StatusBadRequest)
		return
	}

	if err := product.Validate(); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("request body not valid")), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	defer r.Body.Close()

	updatedProduct, err := c.productService.UpdateProduct(ctx, product, userContext)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrUpdateProduct, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, updatedProduct); err != nil {
		c.responder.Error(w, errCtx(api.ErrUpdateProduct, err), http.StatusBadRequest)
		return
	}
}

// DeleteProduct deletes the currently authenticated product
func (c *ProductsController) DeleteProduct(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxDeleteProduct, r.Header.Get("X-Request-Id"))
	urlProductID := chi.URLParam(r, "id")
	productID, err := uuid.FromString(urlProductID)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestParameter, fmt.Errorf("invalid productId, %v", err)), http.StatusBadRequest)
		return
	}
	ctx := context.Background()

	if err := c.productService.DeleteProduct(ctx, productID, userContext); err != nil {
		if err == db.ErrNoMatch {
			c.responder.Error(w, errCtx(api.ErrProductNotFound, errors.New("no product with that id")), http.StatusNotFound)
		} else if err == db.ErrUserForbidden {
			c.responder.Error(w, errCtx(api.ErrUserForbidden, err), http.StatusForbidden)
		} else {
			c.responder.Error(w, errCtx(api.ErrDeleteProduct, err), http.StatusBadRequest)
		}
		return
	}
	c.responder.NoContent(w)
}
