package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/dhurimkelmendi/vending_machine/api"
	"github.com/dhurimkelmendi/vending_machine/controllers"
	"github.com/dhurimkelmendi/vending_machine/fixtures"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
)

func TestProductController(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()

	ctrl := controllers.GetControllersDefaultInstance()
	seller := fixture.User.CreateSellerUser(t)
	secondSeller := fixture.User.CreateSellerUser(t)
	product := fixture.Product.CreateProduct(t, seller.ID)
	allUserOptions := controllers.AuthorizationOptions{
		AllowedUserRoles: []models.UserRole{models.UserRoleSeller, models.UserRoleBuyer},
	}

	t.Run("create product", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/products", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxCreateProduct, ctrl.Products.CreateProduct, allUserOptions))

		bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"name":"%s", "seller_id":"%s", "cost": %d, "amount_available": %d}`,
			strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18], seller.ID.String(), gofakeit.Uint32(), gofakeit.Uint32())))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bBuf)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusCreated {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("get product", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/products/%s", product.ID.String())
		r.Get("/api/v1/products/{id}", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxGetProduct, ctrl.Products.GetProductByID, allUserOptions))

		req := httptest.NewRequest(http.MethodGet, URL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("get all products", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/api/v1/products", ctrl.Products.GetAllProducts)
		URL := "/api/v1/products"

		req := httptest.NewRequest(http.MethodGet, URL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("update product", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/products")
		r.Patch("/api/v1/products", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxUpdateProduct, ctrl.Products.UpdateProduct, allUserOptions))

		t.Run("without permission", func(t *testing.T) {
			newDepositAmount := gofakeit.Uint32()
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","cost":%d}`, product.ID.String(), newDepositAmount)))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secondSeller.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("with empty body", func(t *testing.T) {
			bBuf := bytes.NewBuffer([]byte(""))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("with basic attributes", func(t *testing.T) {
			newCost := gofakeit.Uint32()
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","cost":%d}`, product.ID.String(), newCost)))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusOK {
				t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
			}

			body := make(map[string]interface{})
			dec := json.NewDecoder(strings.NewReader(res.Body.String()))
			err := dec.Decode(&body)
			if err != nil {
				t.Fatalf("error decoding response body: %+v", err)
			}

			cost := body["cost"].(float64)
			if int32(cost) != int32(newCost) {
				t.Fatalf("failed to parse body.cost, got: %+v", cost)
			}
		})
	})

	t.Run("delete product", func(t *testing.T) {
		r := chi.NewRouter()

		t.Run("without permission", func(t *testing.T) {
			URL := fmt.Sprintf("/api/v1/products/%s", product.ID.String())
			r.Delete("/api/v1/products/{id}", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxDeleteProduct, ctrl.Products.DeleteProduct, allUserOptions))
			req := httptest.NewRequest(http.MethodDelete, URL, nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secondSeller.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusForbidden {
				t.Fatalf("expected http status code of 403 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})
		t.Run("non-existing product", func(t *testing.T) {
			URL := fmt.Sprintf("/api/v1/products/%s", gofakeit.UUID())
			r.Delete("/api/v1/products/{id}", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxDeleteProduct, ctrl.Products.DeleteProduct, allUserOptions))
			req := httptest.NewRequest(http.MethodDelete, URL, nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secondSeller.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusNotFound {
				t.Fatalf("expected http status code of 404 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("existing product", func(t *testing.T) {
			URL := fmt.Sprintf("/api/v1/products/%s", product.ID.String())
			r.Delete("/api/v1/products/{id}", ctrl.AuthenticationRequired(ctrl.Products.AuthenticatedController, api.CtxDeleteProduct, ctrl.Products.DeleteProduct, allUserOptions))
			t.Run("without permission", func(t *testing.T) {
				req := httptest.NewRequest(http.MethodDelete, URL, nil)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secondSeller.Token))

				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)

				if res.Code != http.StatusForbidden {
					t.Fatalf("expected http status code of 403 but got: %+v, %+v", res.Code, res.Body.String())
				}
			})
			t.Run("as seller", func(t *testing.T) {
				req := httptest.NewRequest(http.MethodDelete, URL, nil)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)

				if res.Code != http.StatusNoContent {
					t.Fatalf("expected http status code of 204 but got: %+v, %+v", res.Code, res.Body.String())
				}
			})
			t.Run("should not allow a second delete", func(t *testing.T) {
				req := httptest.NewRequest(http.MethodDelete, URL, nil)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", seller.Token))

				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)

				if res.Code != http.StatusNotFound {
					t.Fatalf("expected http status code of 404 but got: %+v, %+v", res.Code, res.Body.String())
				}
			})
		})
	})
}
