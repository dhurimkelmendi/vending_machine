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

// A UsersController handles HTTP requests that deal with user.
type UsersController struct {
	AuthenticatedController
	userService *services.UserService
}

var usersControllerDefaultInstance *UsersController

// GetUsersControllerDefaultInstance returns the default instance of UserController.
func GetUsersControllerDefaultInstance() *UsersController {
	if usersControllerDefaultInstance == nil {
		usersControllerDefaultInstance = NewUserController(services.GetUserServiceDefaultInstance())
	}

	return usersControllerDefaultInstance
}

// NewUserController create a new instance of a user controller using the supplied user service
func NewUserController(userService *services.UserService) *UsersController {
	controller := Controller{
		errCmp:    api.NewErrorComponent(api.CmpController),
		responder: api.GetResponderDefaultInstance(),
	}
	authenticatedController := AuthenticatedController{
		Controller:                      controller,
		statelessAuthenticationProvider: auth.GetStatelessAuthenticationProviderDefaultInstance(),
	}

	return &UsersController{
		AuthenticatedController: authenticatedController,
		userService:             userService,
	}
}

// CreateUser creates a new user and returns user details with an authentication token
func (c *UsersController) CreateUser(w http.ResponseWriter, r *http.Request) {
	errCtx := c.errCmp(api.CtxCreateUser, r.Header.Get("X-Request-Id"))
	user := &payloads.CreateUserPayload{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot decode user")), http.StatusBadRequest)
		return
	}

	if err := user.Validate(); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("request body not valid, missing required fields")), http.StatusBadRequest)
		return
	}

	createdUser, err := c.userService.CreateUser(context.Background(), user)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrCreateUser, err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	c.responder.JSON(w, r, createdUser, http.StatusCreated)
}

// LoginUser returns the user found from the given username&password combination
func (c *UsersController) LoginUser(w http.ResponseWriter, r *http.Request) {
	errCtx := c.errCmp(api.CtxLoginUser, r.Header.Get("X-Request-Id"))

	loginUser := &payloads.LoginUserPayload{}
	if err := json.NewDecoder(r.Body).Decode(loginUser); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("cannot decode user")), http.StatusBadRequest)
		return
	}
	user, err := c.userService.LoginUser(context.Background(), loginUser)
	if err != nil {
		if err == db.ErrNoMatch {
			c.responder.Error(w, errCtx(api.ErrUserNotFound, errors.New("incorrect username or password")), http.StatusNotFound)
		}
		return
	}

	if err := render.Render(w, r, user); err != nil {
		c.responder.Error(w, errCtx(api.ErrLoginUser, err), http.StatusBadRequest)
		return
	}
}

// GetAllUsers returns all active (non-deleted) users
func (c *UsersController) GetAllUsers(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxGetUsers, r.Header.Get("X-Request-Id"))
	users, err := c.userService.GetAllUsers()
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrGetUsers, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, users); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot serialize result")), http.StatusBadRequest)
	}
}

// GetUserByID returns the requested user by id
func (c *UsersController) GetUserByID(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxGetUser, r.Header.Get("X-Request-Id"))
	urlUserID := chi.URLParam(r, "id")
	userID, err := uuid.FromString(urlUserID)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestParameter, fmt.Errorf("invalid userId, %v", err)), http.StatusBadRequest)
		return
	}

	user, err := c.userService.GetUserByID(userID)
	if err != nil {
		if err == db.ErrNoMatch {
			c.responder.Error(w, errCtx(api.ErrUserNotFound, errors.New("no user with that id")), http.StatusNotFound)
		} else {
			c.responder.Error(w, errCtx(api.ErrGetUser, err), http.StatusBadRequest)
		}
		return
	}

	var res render.Renderer

	// Ensure we don't leak any private details
	if userID != userContext.ID {
		res = payloads.MapUserToUserDetails(user)
	} else {
		res = user
	}

	if err := render.Render(w, r, res); err != nil {
		c.responder.Error(w, errCtx(api.ErrCreatePayload, errors.New("cannot serialize result")), http.StatusBadRequest)
		return
	}
}

// UpdateUser updates the current users profile
func (c *UsersController) UpdateUser(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxUpdateUser, r.Header.Get("X-Request-Id"))

	user := &payloads.UpdateUserPayload{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("cannot decode user")), http.StatusBadRequest)
		return
	}
	user.ID = userContext.ID

	if err := user.Validate(); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("request body not valid")), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	defer r.Body.Close()

	updatedUser, err := c.userService.UpdateUser(ctx, user)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrUpdateUser, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, updatedUser); err != nil {
		c.responder.Error(w, errCtx(api.ErrUpdateUser, err), http.StatusBadRequest)
		return
	}
}

// DepositMoney updates current users deposit amount
func (c *UsersController) DepositMoney(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxDepositMoney, r.Header.Get("X-Request-Id"))

	depositMoney := &payloads.DepositMoneyPayload{}
	if err := json.NewDecoder(r.Body).Decode(depositMoney); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("cannot decode deposit payload")), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	defer r.Body.Close()

	updatedUser, err := c.userService.DepositMoney(ctx, depositMoney, userContext.ID)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrDepositMoney, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, updatedUser); err != nil {
		c.responder.Error(w, errCtx(api.ErrDepositMoney, err), http.StatusBadRequest)
		return
	}
}

// ResetDeposit reset current users deposit amount
func (c *UsersController) ResetDeposit(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxDepositMoney, r.Header.Get("X-Request-Id"))

	ctx := context.Background()
	defer r.Body.Close()

	updatedUser, err := c.userService.ResetDeposit(ctx, userContext.ID)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrResetDeposit, err), http.StatusBadRequest)
		return
	}

	if err := render.Render(w, r, updatedUser); err != nil {
		c.responder.Error(w, errCtx(api.ErrResetDeposit, err), http.StatusBadRequest)
		return
	}
}

// DeleteUser deletes the currently authenticated user
func (c *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxDeleteUser, r.Header.Get("X-Request-Id"))

	ctx := context.Background()

	if err := c.userService.DeleteUser(ctx, userContext.ID); err != nil {
		if err == db.ErrNoMatch {
			c.responder.Error(w, errCtx(api.ErrUserNotFound, errors.New("no user with that id")), http.StatusNotFound)
		} else if err == db.ErrUserForbidden {
			c.responder.Error(w, errCtx(api.ErrUserForbidden, err), http.StatusForbidden)
		} else {
			c.responder.Error(w, errCtx(api.ErrDeleteUser, err), http.StatusBadRequest)
		}
		return
	}
	c.responder.NoContent(w)
}

// BuyProduct links a given user to the provided product using the request payload
func (c *UsersController) BuyProduct(w http.ResponseWriter, r *http.Request, userContext auth.UserContext) {
	errCtx := c.errCmp(api.CtxCreateUser, r.Header.Get("X-Request-Id"))

	userProduct := &payloads.UserProductPurchase{}

	if err := json.NewDecoder(r.Body).Decode(userProduct); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, fmt.Errorf("cannot decode user_product payload: %v", err)), http.StatusBadRequest)
		return
	}

	if err := userProduct.Validate(); err != nil {
		c.responder.Error(w, errCtx(api.ErrInvalidRequestPayload, errors.New("request body not valid, missing required fields")), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	defer r.Body.Close()

	userReport, err := c.userService.BuyProduct(ctx, userProduct)
	if err != nil {
		c.responder.Error(w, errCtx(api.ErrBuyProduct, err), http.StatusBadRequest)
		return
	}
	if err := render.Render(w, r, userReport); err != nil {
		c.responder.Error(w, errCtx(api.ErrResetDeposit, err), http.StatusBadRequest)
		return
	}
}
