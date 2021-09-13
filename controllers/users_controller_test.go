package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhurimkelmendi/vending_machine/api"
	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/controllers"
	"github.com/dhurimkelmendi/vending_machine/fixtures"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/payloads"
	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
)

func TestUserController(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()

	ctrl := controllers.GetControllersDefaultInstance()
	buyerUser := fixture.User.CreateBuyerUser(t)
	secondBuyerUser := fixture.User.CreateBuyerUser(t)
	sellerUser := fixture.User.CreateSellerUser(t)
	product := fixture.Product.CreateProduct(t, sellerUser.ID)
	allUserOptions := controllers.AuthorizationOptions{
		AllowedUserRoles: []models.UserRole{models.UserRoleSeller, models.UserRoleBuyer},
	}
	buyerOnlyOptions := controllers.AuthorizationOptions{
		AllowedUserRoles: []models.UserRole{models.UserRoleBuyer},
	}

	// generate non-existing user token by using user.ID = -1, which doesn't exist because user.ID is autoincrement
	invalidUserToken, _ := GetInvalidAuthToken()

	t.Run("create user", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/users", ctrl.Users.CreateUser)

		bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s", "password":"123456789", "role": "%s", "deposit": %d}`,
			strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18], models.UserRoleBuyer, int32(rand.Intn(1000)+rand.Intn(1000))*5)))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bBuf)

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusCreated {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("get user", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/users/%s", buyerUser.ID.String())
		r.Get("/api/v1/users/{id}", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxGetUser, ctrl.Users.GetUserByID, allUserOptions))

		req := httptest.NewRequest(http.MethodGet, URL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("login user", func(t *testing.T) {
		r := chi.NewRouter()
		URL := "/api/v1/users/login"
		r.Post("/api/v1/users/login", ctrl.Users.LoginUser)
		bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s","password":"%s"}`, buyerUser.Username, buyerUser.Password)))

		req := httptest.NewRequest(http.MethodPost, URL, bBuf)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("get all users", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/api/v1/users", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxGetUsers, ctrl.Users.GetAllUsers, allUserOptions))
		URL := "/api/v1/users"

		req := httptest.NewRequest(http.MethodGet, URL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("update user", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/user")
		r.Patch("/api/v1/user", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxUpdateUser, ctrl.Users.UpdateUser, allUserOptions))

		t.Run("without permission", func(t *testing.T) {
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","username":"%s"}`, buyerUser.ID, strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18])))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", invalidUserToken))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("with empty body", func(t *testing.T) {
			bBuf := bytes.NewBuffer([]byte(""))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("with basic attributes", func(t *testing.T) {
			newUsername := strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18]
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","username":"%s"}`, buyerUser.ID, newUsername)))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

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

			username := body["username"].(string)
			if username != newUsername {
				t.Fatalf("failed to parse body.username, got: %+v", username)
			}
		})
	})
	t.Run("deposit money", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/deposit")
		r.Post("/api/v1/deposit", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxDepositMoney, ctrl.Users.DepositMoney, buyerOnlyOptions))

		t.Run("as seller(without permission)", func(t *testing.T) {
			acceptableDepositAmountValues := config.GetDefaultInstance().AcceptableDepositAmountValues
			newDepositAmount := acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","deposit":%d}`, sellerUser.ID, newDepositAmount)))
			req := httptest.NewRequest(http.MethodPost, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sellerUser.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusForbidden {
				t.Fatalf("expected http status code of 403 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})
		t.Run("as buyer", func(t *testing.T) {
			t.Run("acceptable amount", func(t *testing.T) {
				acceptableDepositAmountValues := config.GetDefaultInstance().AcceptableDepositAmountValues
				newDepositAmount := acceptableDepositAmountValues[rand.Intn(len(acceptableDepositAmountValues))]
				bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","deposit_amount":%d}`, buyerUser.ID, newDepositAmount)))
				req := httptest.NewRequest(http.MethodPost, URL, bBuf)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))
				oldDepositAmount := buyerUser.Deposit
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

				deposit := body["deposit"].(float64)
				if int32(deposit) != (oldDepositAmount + int32(newDepositAmount)) {
					t.Fatalf("unexpected deposit amount, got: %+v", deposit)
				}
			})
			t.Run("unacceptable amount", func(t *testing.T) {
				newDepositAmount := 222
				bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","deposit_amount":%d}`, buyerUser.ID, newDepositAmount)))
				req := httptest.NewRequest(http.MethodPost, URL, bBuf)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))

				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)

				body := make(map[string]interface{})
				dec := json.NewDecoder(strings.NewReader(res.Body.String()))
				err := dec.Decode(&body)
				if err != nil {
					t.Fatalf("error decoding response body: %+v", err)
				}
				if res.Code != http.StatusBadRequest {
					t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
				}
			})
		})
	})
	t.Run("reset deposit", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/reset")
		r.Post("/api/v1/reset", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxDepositMoney, ctrl.Users.ResetDeposit, buyerOnlyOptions))

		t.Run("as seller(without permission)", func(t *testing.T) {
			bBuf := bytes.NewBuffer([]byte(""))
			req := httptest.NewRequest(http.MethodPost, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sellerUser.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusForbidden {
				t.Fatalf("expected http status code of 403 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})
		t.Run("as buyer", func(t *testing.T) {
			bBuf := bytes.NewBuffer([]byte(""))
			req := httptest.NewRequest(http.MethodPost, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", secondBuyerUser.Token))
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

			deposit := body["deposit"].(float64)
			if int32(deposit) != 0 {
				t.Fatalf("unexpected deposit amount, got: %+v", deposit)
			}
		})
	})

	t.Run("delete user", func(t *testing.T) {
		r := chi.NewRouter()

		t.Run("without permission / non-existing user", func(t *testing.T) {
			URL := "/api/v1/user"
			r.Delete("/api/v1/user", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxDeleteUser, ctrl.Users.DeleteUser, allUserOptions))
			req := httptest.NewRequest(http.MethodDelete, URL, nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", invalidUserToken))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusNotFound {
				t.Fatalf("expected http status code of 404 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("existing user", func(t *testing.T) {
			newUser := fixture.User.CreateBuyerUser(t)

			URL := "/api/v1/user"
			r.Delete("/api/v1/user", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxDeleteUser, ctrl.Users.DeleteUser, allUserOptions))
			req := httptest.NewRequest(http.MethodDelete, URL, nil)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", newUser.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusNoContent {
				t.Fatalf("expected http status code of 204 but got: %+v, %+v", res.Code, res.Body.String())
			}

			t.Run("should not allow a second delete", func(t *testing.T) {
				req := httptest.NewRequest(http.MethodDelete, URL, nil)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", newUser.Token))

				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)

				if res.Code != http.StatusNotFound {
					t.Fatalf("expected http status code of 404 but got: %+v, %+v", res.Code, res.Body.String())
				}
			})
		})
	})
	t.Run("user buys product", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/deposit")
		r.Post("/api/v1/deposit", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxBuyProduct, ctrl.Users.BuyProduct, buyerOnlyOptions))
		t.Run("with sufficient deposit", func(t *testing.T) {
			productAmount := 1
			oldDepositAmount := buyerUser.Deposit
			newDepositAmount := oldDepositAmount

			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"user_id":"%s","product_id":"%s","amount":%d}`, buyerUser.ID.String(), product.ID.String(), productAmount)))
			req := httptest.NewRequest(http.MethodPost, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", buyerUser.Token))
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusOK {
				t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
			}

			userReport := &payloads.UserBuysReport{}
			dec := json.NewDecoder(strings.NewReader(res.Body.String()))
			err := dec.Decode(&userReport)
			if err != nil {
				t.Fatalf("error decoding response body: %+v", err)
			}
			newDepositAmount = oldDepositAmount - userReport.AmountSpent
			if newDepositAmount < 0 {
				t.Fatalf("expected deposit not to be negative after purchase, got %+v", newDepositAmount)
			}
		})
	})
}
