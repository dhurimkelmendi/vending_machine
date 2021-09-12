package controllers

import (
	"fmt"
	"net/http"

	"github.com/dhurimkelmendi/vending_machine/api"
	"github.com/dhurimkelmendi/vending_machine/auth"
	"github.com/dhurimkelmendi/vending_machine/helpers"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/dhurimkelmendi/vending_machine/services"
)

// Controllers is a struct that contains references to all controller instances.
type Controllers struct {
	userService *services.UserService
	Users       *UsersController
}

// Controller is a struct that contains references to error components and responders
type Controller struct {
	errCmp    api.ErrorComponentFn
	responder *api.Responder
}

// AuthenticatedController is a struct used to make routes un-accessible without authorization
type AuthenticatedController struct {
	Controller
	statelessAuthenticationProvider *auth.StatelessAuthenticationProvider
}

// AuthorizationOptions is a struct that contains references to allowed user roles
type AuthorizationOptions struct {
	AllowedUserRoles []models.UserRole
}

var controllersDefaultInstance *Controllers

// GetControllersDefaultInstance returns default instances of all available Controllers
func GetControllersDefaultInstance() *Controllers {
	if controllersDefaultInstance == nil {
		controllersDefaultInstance = &Controllers{
			userService: services.GetUserServiceDefaultInstance(),
			Users:       GetUsersControllerDefaultInstance(),
		}
	}
	return controllersDefaultInstance
}

// AuthenticatedHandlerFunc is a handler function type that requires authorization
type AuthenticatedHandlerFunc func(http.ResponseWriter, *http.Request, auth.UserContext)

// AuthenticationRequired implements user access control
func (cs *Controllers) AuthenticationRequired(c AuthenticatedController, errorContext api.ErrorContext, fn AuthenticatedHandlerFunc, opts AuthorizationOptions) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		errCtx := c.Controller.errCmp(errorContext, r.Header.Get("X-Request-Id"))
		userContext, err := c.statelessAuthenticationProvider.GetCurrentUserContext(r)
		if err != nil {
			c.Controller.responder.Error(w, errCtx(api.ErrUserForbidden, err))
			return
		}

		if !helpers.UserRolesContains(opts.AllowedUserRoles, userContext.Role) {
			c.Controller.responder.Error(w, errCtx(api.ErrUserForbidden, fmt.Errorf("user is forbidden")))
			return
		}

		fn(w, r, *userContext)
	}
}
