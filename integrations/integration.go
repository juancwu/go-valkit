package integrations

import "github.com/juancwu/go-valkit/validator"

// ErrorFormatter converts validation errors to a structure suitable for HTTP responses
type ErrorFormatter interface {
	Format(ve validator.ValidationErrors) interface{}
}

// DefaultErrorResponse is a standard structure for validation error responses
type DefaultErrorResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// DefaultErrorFormatter converts validation errors to a map of field paths to error messages
type DefaultErrorFormatter struct{}

func (f *DefaultErrorFormatter) Format(ve validator.ValidationErrors) interface{} {
	errorMap := make(map[string][]string)

	for _, fieldErr := range ve {
		path := fieldErr.Path
		if _, exists := errorMap[path]; !exists {
			errorMap[path] = []string{}
		}
		errorMap[path] = append(errorMap[path], fieldErr.Message)
	}

	return DefaultErrorResponse{
		Status:  "error",
		Message: "Validation failed",
		Errors:  errorMap,
	}
}
