package api

// ErrorComponent is the name of the component for the API error.
type ErrorComponent string

// ErrorComponentFn is a function that creates error contexts for the function's component.
type ErrorComponentFn func(ErrorContext, ...string) ErrorContextFn

// NewErrorComponent creates a new error component context creator function.
func NewErrorComponent(componentName ErrorComponent) ErrorComponentFn {
	return func(contextName ErrorContext, contextIDs ...string) ErrorContextFn {
		return func(apiErr *ResponseError, innerErr error) error {
			apiErr.Component = componentName

			apiErr.Context = contextName
			if len(contextIDs) > 0 && len(contextIDs[0]) > 0 {
				apiErr.ContextID = contextIDs[0]
			}

			apiErr.InnerError = innerErr

			return apiErr
		}
	}
}

const (
	// CmpAuthentication is the error component that represents all Authentication level errors.
	CmpAuthentication ErrorComponent = "cmpAuthentication"

	// CmpController is the error component that represents all controller level errors.
	CmpController ErrorComponent = "cmpController"

	// CmpAdminController is the error component that represents all controller level errors.
	CmpAdminController ErrorComponent = "cmpAdminController"

	// CmpSerializer is the error component that represents all serializer level errors.
	CmpSerializer ErrorComponent = "cmpSerializer"

	// CmpService is the error component that represents all service level errors.
	CmpService ErrorComponent = "cmpService"
)
