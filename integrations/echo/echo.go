// Package echo provides Echo framework integration for the validator library
package echo

import (
	"net/http"

	"github.com/juancwu/go-valkit/integrations"
	"github.com/juancwu/go-valkit/validator"
	"github.com/labstack/echo/v4"
)

// Validator implements echo.Validator interface using our validator library
type Validator struct {
	validator *validator.Validator
	formatter integrations.ErrorFormatter
}

// NewValidator creates a new validator for Echo using default settings
func NewValidator() *Validator {
	v := validator.New()
	v.UseJsonTagName() // Use JSON field names in error messages

	return &Validator{
		validator: v,
		formatter: &integrations.DefaultErrorFormatter{},
	}
}

// Validate implements echo.Validator interface
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Validate(i)
}

// GetValidator returns the underlying validator instance for customization
func (v *Validator) GetValidator() *validator.Validator {
	return v.validator
}

// WithFormatter sets a custom error formatter
func (v *Validator) WithFormatter(formatter integrations.ErrorFormatter) *Validator {
	v.formatter = formatter
	return v
}

// Configure sets up Echo with the validator and appropriate error handling
func Configure(e *echo.Echo) {
	v := NewValidator()
	e.Validator = v
	e.HTTPErrorHandler = ValidationErrorHandler(v, e.DefaultHTTPErrorHandler)
}

// ValidationErrorHandler returns an echo.HTTPErrorHandler for validation errors
func ValidationErrorHandler(v *Validator, next echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if ve, ok := err.(validator.ValidationErrors); ok {
			response := v.formatter.Format(ve)
			c.JSON(http.StatusBadRequest, response)
			return
		}

		// For other error types, use the default handler
		next(err, c)
	}
}

// ConfigureWithOptions sets up Echo with a customized validator
func ConfigureWithOptions(e *echo.Echo, configFn func(*validator.Validator)) {
	v := NewValidator()

	if configFn != nil {
		configFn(v.validator)
	}

	e.Validator = v
	e.HTTPErrorHandler = ValidationErrorHandler(v, e.DefaultHTTPErrorHandler)
}
