package api

// ErrorContext is the name of the context for the API error.
type ErrorContext string

// ErrorContextFn is a function that decorate the given errors in the error context of the function.
type ErrorContextFn func(*ResponseError, error) error

// Authentication error contexts
const (
	CtxAuthentication ErrorContext = "ctxAuthentication"
)

// Serializer error contexts
const (
	CtxSerializeUser ErrorContext = "ctxSerializeUser"
)
