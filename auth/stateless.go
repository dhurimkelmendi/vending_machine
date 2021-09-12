package auth

import (
	"errors"
	"net/http"

	"github.com/dhurimkelmendi/vending_machine/api"
	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/models"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"

	uuid "github.com/satori/go.uuid"
)

// StatelessAuthenticationProvider provides stateless authentication.
type StatelessAuthenticationProvider struct {
	errCmp    api.ErrorComponentFn
	TokenAuth *jwtauth.JWTAuth
}

// UserContext contains user details for the current request context
type UserContext struct {
	ID   uuid.UUID
	Role models.UserRole
}

var statelessAuthenticationProviderDefaultInstance *StatelessAuthenticationProvider

// GetStatelessAuthenticationProviderDefaultInstance returns the default instance of StatelessAuthenticationProvider
func GetStatelessAuthenticationProviderDefaultInstance() *StatelessAuthenticationProvider {
	if statelessAuthenticationProviderDefaultInstance == nil {
		jwtTokenAuth := jwtauth.New("HS256", []byte(config.GetDefaultInstance().JWTSecret), nil)

		statelessAuthenticationProviderDefaultInstance = &StatelessAuthenticationProvider{
			errCmp:    api.NewErrorComponent(api.CmpAuthentication),
			TokenAuth: jwtTokenAuth,
		}
	}
	return statelessAuthenticationProviderDefaultInstance
}

// Authenticator ensures that the request has an auth token
func (p *StatelessAuthenticationProvider) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errCtx := p.errCmp(api.CtxAuthentication)
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Error(w, errCtx(api.ErrInvalidAuth, errors.New("authorization header is invalid")).Error(), http.StatusUnauthorized)
			return
		}

		if token == nil || jwt.Validate(token) != nil {
			http.Error(w, errCtx(api.ErrInvalidAuth, errors.New("authorization header is invalid")).Error(), http.StatusUnauthorized)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

// GetCurrentUserContext gets the user's context from the provided authentication token
func (p *StatelessAuthenticationProvider) GetCurrentUserContext(r *http.Request) (*UserContext, error) {
	token, err := jwtauth.VerifyRequest(p.TokenAuth, r, jwtauth.TokenFromHeader)
	ctx := jwtauth.NewContext(r.Context(), token, err)
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	sub := token.Subject()
	if sub == "" {
		return nil, errors.New("missing subject claim")
	}

	userID, err := uuid.FromString(sub)
	if err != nil {
		return nil, errors.New("invalid subject claim")
	}

	userRole, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("invalid user role claim")
	}

	return &UserContext{
		ID:   userID,
		Role: models.UserRole(userRole),
	}, nil
}

// CreateUserAuthToken creates a JWT authentication token for the supplied user
func (p *StatelessAuthenticationProvider) CreateUserAuthToken(user *models.User) (string, error) {
	claims := map[string]interface{}{
		"sub":      user.ID.String(),
		"username": user.Username,
		"role":     user.Role,
	}
	jwtauth.SetIssuedNow(claims)

	_, tokenString, err := p.TokenAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
