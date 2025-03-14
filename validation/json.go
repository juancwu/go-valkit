package validation

import (
	"encoding/json"
)

// JSONError represents a validation error in JSON format
type JSONError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Param   string `json:"param,omitempty"`
}

// JSONErrors represents a collection of validation errors in JSON format
type JSONErrors struct {
	Errors []JSONError `json:"errors"`
}

// ToJSON converts validation errors to a JSON-friendly format
func (errs ValidationErrors) ToJSON() JSONErrors {
	jsonErrors := make([]JSONError, 0, len(errs))

	for _, err := range errs {
		jsonErrors = append(jsonErrors, JSONError{
			Field:   err.Field,
			Message: err.Message,
			Tag:     err.Tag,
			Param:   err.Param,
		})
	}

	return JSONErrors{
		Errors: jsonErrors,
	}
}

// MarshalJSON implements the json.Marshaler interface
func (errs ValidationErrors) MarshalJSON() ([]byte, error) {
	return json.Marshal(errs.ToJSON())
}
