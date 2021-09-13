package api

import "net/http"

// Ensure ResponseError conforms to the error interface.
var _ error = &ResponseError{}

// ResponseError represents an error from the API.
type ResponseError struct {
	Component  ErrorComponent
	Context    ErrorContext
	ContextID  string // ID of a context (i.e. request)
	Message    string
	Code       string
	Status     int
	InnerError error
}

// Error returns the error message with error code as a string.
func (e *ResponseError) Error() string {
	msg := e.Code + ": " + e.Message
	if e.InnerError != nil {
		msg = msg + "; " + e.InnerError.Error()
	}
	return msg
}

// NewResponseError returns new API error.
func NewResponseError(code, message string, statuses ...int) *ResponseError {
	apiErr := &ResponseError{Message: message, Code: code}
	if len(statuses) > 0 && statuses[0] > 0 {
		apiErr.Status = statuses[0]
	}
	return apiErr
}

// All error codes
var (
	// Payload parsing/serializing errors
	ErrInvalidRequestPayload   = NewResponseError("errInvalidRequestPayload", "request payload is invalid", http.StatusBadRequest)
	ErrInvalidRequestParameter = NewResponseError("errInvalidRequestParameter", "request parameter is invalid", http.StatusBadRequest)
	ErrCreatePayload           = NewResponseError("errCreatePayload", "unable to generate response payload")

	// Auth errors
	ErrInvalidAuth    = NewResponseError("errInvalidAuth", "invalid authorization", http.StatusUnauthorized)
	ErrUserForbidden  = NewResponseError("errUserForbidden", "user is not permitted", http.StatusForbidden)
	ErrCreateUserAuth = NewResponseError("errCreateUserAuth", "unable to authorize user")

	// User errors
	ErrUserNotFound = NewResponseError("errUserNotFound", "unable to find user", http.StatusNotFound)
	ErrGetUsers     = NewResponseError("errFindUser", "unable to get users")
	ErrGetUser      = NewResponseError("errFindUser", "unable to get user")
	ErrLoginUser    = NewResponseError("errLoginUser", "unable to login user")
	ErrCreateUser   = NewResponseError("errCreateUser", "unable to register user")
	ErrUpdateUser   = NewResponseError("errUpdateUser", "unable to update user")
	ErrDeleteUser   = NewResponseError("errDeleteUser", "unable to delete user")

	// Product errors
	ErrProductNotFound = NewResponseError("errProductNotFound", "unable to find user", http.StatusNotFound)
	ErrGetProducts     = NewResponseError("errFindProduct", "unable to get users")
	ErrGetProduct      = NewResponseError("errFindProduct", "unable to get user")
	ErrCreateProduct   = NewResponseError("errCreateProduct", "unable to register user")
	ErrUpdateProduct   = NewResponseError("errUpdateProduct", "unable to update user")
	ErrDeleteProduct   = NewResponseError("errDeleteProduct", "unable to delete user")
)
