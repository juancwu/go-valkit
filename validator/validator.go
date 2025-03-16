package validator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validator          *govalidator.Validate
	DefaultMessage     string
	DefaultTagMessages map[string]string
	Messages           ValidationMessages
}

// New creates a new Validator instance with default configuration.
func New() *Validator {
	v := govalidator.New()

	return &Validator{
		validator:          v,
		DefaultMessage:     "Invalid value",
		DefaultTagMessages: make(map[string]string),
		Messages:           NewValidationMessages(),
	}
}

// UseMessages creates a new Validator instance with the same base configuration
// but with different validation messages for the specific context.
// This allows different handlers or methods to have custom error messages.
func (v *Validator) UseMessages(messages ValidationMessages) *Validator {
	newV := &Validator{
		validator:          v.validator,
		DefaultMessage:     v.DefaultMessage,
		DefaultTagMessages: make(map[string]string),
		Messages:           messages,
	}

	newV.DefaultMessage = v.DefaultMessage

	for tag, msg := range v.DefaultTagMessages {
		newV.DefaultTagMessages[tag] = msg
	}

	return newV
}

// Validate performs validation on the provided struct based on its validation tags.
// Returns nil if validation passes, or ValidationErrors containing details about
// validation failures.
func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		validationErrors := ValidationErrors{}

		for _, err := range err.(govalidator.ValidationErrors) {
			path := getFullPath(err)
			normPath := normalizePath(path)

			// Create validation error with basic information
			valError := ValidationError{
				Field:      err.Field(),
				Path:       path,
				Constraint: err.Tag(),
				Param:      err.Param(),
			}

			// Create params for interpolation
			params := CreateValidationParams(valError)

			// Try to get message from ValidationMessages first
			message := v.Messages.ResolveMessage(normPath, err.Tag(), params)

			// Fall back to default message lookup
			if message == "" {
				if msg, ok := v.DefaultTagMessages[err.Tag()]; ok {
					message = interpolateParams(msg, params)
				} else {
					message = interpolateParams(v.DefaultMessage, params)
				}
			}

			valError.Message = message
			validationErrors = append(validationErrors, valError)
		}

		return validationErrors
	}

	return nil
}

// extractArrayIndex tries to extract an array index from a path like "users[2].name"
// Returns the index and a boolean indicating success
func extractArrayIndex(path string) (int, bool) {
	re := regexp.MustCompile(`\[(\d+)\]`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		if idx, err := strconv.Atoi(matches[1]); err == nil {
			return idx, true
		}
	}
	return 0, false
}

// SetDefaultMessage sets the default message that should be used if not path/tag matches
func (v *Validator) SetDefaultMessage(s string) *Validator {
	v.DefaultMessage = s
	return v
}

// SetDefaultTagMessage sets the default message for a specific tag error.
// Example: "required", "email"
func (v *Validator) SetDefaultTagMessage(tag string, s string) *Validator {
	v.DefaultTagMessages[tag] = s
	return v
}

// SetConstraintMessage sets a specific message for a field path and constraint combination.
// Example: v.SetConstraintMessage("user.profile.firstname", "required", "First name is required")
func (v *Validator) SetConstraintMessage(path, constraint, message string) *Validator {
	path = normalizePath(path)
	v.Messages.SetMessage(path, constraint, message)
	return v
}

// SetPathDefaultMessage sets a default message for a field path to use when no constraint-specific
// message is found.
// Example: v.SetPathDefaultMessage("user.profile.firstname", "First name is invalid")
func (v *Validator) SetPathDefaultMessage(path, message string) *Validator {
	path = normalizePath(path)
	v.Messages.SetDefaultMessage(path, message)
	return v
}

// UseJsonTagName configures the validator to use the JSON tag names in error messages
// and field paths instead of the Go struct field names.
//
// This makes error messages more useful when working with JSON APIs, as the field names
// in error messages will match the names used in JSON requests/responses.
//
// Example:
//
//	type User struct {
//	    FirstName string `json:"first_name" validate:"required"`
//	}
//
// Without UseJsonTagName: error path would be "FirstName"
// With UseJsonTagName: error path would be "first_name"
func (v *Validator) UseJsonTagName() *Validator {
	v.validator.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			// If JSON tag is "-" (meaning "don't include in JSON"), use the actual field name
			return field.Name
		}
		return name
	})
	return v
}

func getLegacyMessage(v *Validator, err govalidator.FieldError, path string) string {
	// Normalize path to keep path-key consistent for lookup
	path = normalizePath(path)

	// Match default message by tag
	if msg, ok := v.DefaultTagMessages[err.Tag()]; ok {
		return msg
	}

	// Use default message
	return v.DefaultMessage
}
