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

func TestUserController(t *testing.T) {
	t.Parallel()
	fixture := fixtures.GetFixturesDefaultInstance()

	ctrl := controllers.GetControllersDefaultInstance()
	user := fixture.User.CreateBuyerUser(t)
	allUserOptions := controllers.AuthorizationOptions{
		AllowedUserRoles: []models.UserRole{models.UserRoleSeller, models.UserRoleBuyer},
	}

	// generate non-existing user token by using user.ID = -1, which doesn't exist because user.ID is autoincrement
	invalidUserToken, _ := GetInvalidAuthToken()

	t.Run("create user", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/api/v1/users", ctrl.Users.CreateUser)

		bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s", "password":"123456789", "role": "%s", "deposit": %d}`,
			strings.Replace(uuid.NewV4().String(), "-", "_", -1)[0:18], models.UserRoleBuyer, gofakeit.Uint32())))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bBuf)

		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		if res.Code != http.StatusCreated {
			t.Fatalf("expected http status code of 200 but got: %+v, %+v", res.Code, res.Body.String())
		}
	})

	t.Run("get user", func(t *testing.T) {
		r := chi.NewRouter()
		URL := fmt.Sprintf("/api/v1/users/%s", user.ID.String())
		r.Get("/api/v1/users/{id}", ctrl.AuthenticationRequired(ctrl.Users.AuthenticatedController, api.CtxGetUser, ctrl.Users.GetUserByID, allUserOptions))

		req := httptest.NewRequest(http.MethodGet, URL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.Token))

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
		bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s","password":"%s"}`, user.Username, user.Password)))

		req := httptest.NewRequest(http.MethodPost, URL, bBuf)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.Token))

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
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.Token))

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
			newDepositAmount := gofakeit.Uint32()
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","deposit":%d}`, user.ID, newDepositAmount)))
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
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.Token))

			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("expected http status code of 400 but got: %+v, %+v", res.Code, res.Body.String())
			}
		})

		t.Run("with basic attributes", func(t *testing.T) {
			newDepositAmount := gofakeit.Uint32()
			bBuf := bytes.NewBuffer([]byte(fmt.Sprintf(`{"id":"%s","deposit":%d}`, user.ID, newDepositAmount)))
			req := httptest.NewRequest(http.MethodPatch, URL, bBuf)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.Token))

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
			if int32(deposit) != int32(newDepositAmount) {
				t.Fatalf("failed to parse body.deposit, got: %+v", deposit)
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
}
