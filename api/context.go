package api

// ErrorContext is the name of the context for the API error.
type ErrorContext string

// ErrorContextFn is a function that decorate the given errors in the error context of the function.
type ErrorContextFn func(*ResponseError, error) error

// Authentication error contexts
const (
	CtxAuthentication ErrorContext = "ctxAuthentication"
)

// User error contexts
const (
	CtxGetUsers     ErrorContext = "ctxGetUsers"
	CtxGetUser      ErrorContext = "ctxGetUser"
	CtxLoginUser    ErrorContext = "ctxLoginUser"
	CtxCreateUser   ErrorContext = "ctxCreateUser"
	CtxUpdateUser   ErrorContext = "ctxUpdateUser"
	CtxDepositMoney ErrorContext = "ctxDepositMoney"
	CtxResetDeposit ErrorContext = "ctxResetDeposit"
	CtxDeleteUser   ErrorContext = "ctxDeleteUser"
)

// Product error contexts
const (
	CtxGetProducts   ErrorContext = "ctxGetProducts"
	CtxGetProduct    ErrorContext = "ctxGetProduct"
	CtxLoginProduct  ErrorContext = "ctxLoginProduct"
	CtxCreateProduct ErrorContext = "ctxCreateProduct"
	CtxUpdateProduct ErrorContext = "ctxUpdateProduct"
	CtxDeleteProduct ErrorContext = "ctxDeleteProduct"
)

// Serializer error contexts
const (
	CtxSerializeUser ErrorContext = "ctxSerializeUser"
)
