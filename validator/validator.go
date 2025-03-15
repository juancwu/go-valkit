package validator

import (
	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *govalidator.Validate
}

// New creates a new Validator instance with default configuration.
func New() *Validator {
	v := govalidator.New()

	return &Validator{
		validator: v,
	}
}

// Validate performs validation on the provided struct based on its validation tags.
// Returns nil if validation passes, or ValidationErrors containing details about
// validation failures.
func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		validationErrors := ValidationErrors{}

		for _, err := range err.(govalidator.ValidationErrors) {
			path := getFullPath(err)
			message := "invalid value"
			valError := ValidationError{
				Path:    path,
				Message: message,
				Tag:     err.Tag(),
				Param:   err.Param(),
			}
			validationErrors = append(validationErrors, valError)
		}

		return validationErrors
	}

	return nil
}
