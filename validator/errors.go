package validator

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error for a specific field.
// It includes the field path, error message, and metadata about the validation rule.
type ValidationError struct {
	Field      string      `json:"field"`                // Field name of the leaf in path
	Path       string      `json:"path"`                 // JSON path to the field with the error
	Message    string      `json:"message"`              // Human-readable error message
	Constraint string      `json:"constraint,omitempty"` // Validation tag that failed (e.g., "required", "min")
	Param      string      `json:"param,omitempty"`      // Parameter for the validation tag (e.g., "5" for min=5)
	Actual     interface{} `json:"actual,omitempty"`     // Actual value that failed validation
}

// Error implements the error interface to allow ValidationError to be used as an error.
// Returns a formatted error string including the field path and error message.
func (ve ValidationError) Error() string {
	return fmt.Sprintf("ValidationError (%s | %s): %s", ve.Path, ve.Constraint, ve.Message)
}

// ValidationErrors is a collection of ValidationError objects.
// This type allows returning multiple validation errors at once.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
// It concatenates all individual validation error messages with semicolons.
func (ve ValidationErrors) Error() string {
	var builder strings.Builder
	for _, err := range ve {
		builder.WriteString(err.Error())
		builder.WriteString(";")
	}
	return builder.String()
}

// GroupErrorsByPath groups validation errors by their path
func (ve ValidationErrors) GroupErrorsByPath() map[string]ValidationErrors {
	groupedErrors := make(map[string]ValidationErrors)

	for _, err := range ve {
		groupedErrors[err.Path] = append(groupedErrors[err.Path], err)
	}

	return groupedErrors
}

// ErrorsForPath returns all errors for a specific path
func (ve ValidationErrors) ErrorsForPath(path string) ValidationErrors {
	var pathErrors ValidationErrors

	for _, err := range ve {
		if err.Path == path {
			pathErrors = append(pathErrors, err)
		}
	}

	return pathErrors
}
