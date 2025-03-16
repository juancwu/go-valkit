package validator

import (
	"reflect"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validator           *govalidator.Validate
	DefaultMessage      string
	DefaultTagMessages  map[string]string
	CustomFieldMessages map[string]string
	Messages            ValidationMessages
}

// New creates a new Validator instance with default configuration.
func New() *Validator {
	v := govalidator.New()

	return &Validator{
		validator:           v,
		DefaultMessage:      "Invalid value",
		DefaultTagMessages:  make(map[string]string),
		CustomFieldMessages: make(map[string]string),
		Messages:            NewValidationMessages(),
	}
}

// UseMessages creates a new Validator instance with the same base configuration
// but with different validation messages for the specific context.
// This allows different handlers or methods to have custom error messages.
func (v *Validator) UseMessages(messages ValidationMessages) *Validator {
	newV := &Validator{
		validator:           v.validator,
		DefaultMessage:      v.DefaultMessage,
		DefaultTagMessages:  make(map[string]string),
		CustomFieldMessages: make(map[string]string),
		Messages:            messages,
	}

	newV.DefaultMessage = v.DefaultMessage

	for tag, msg := range v.DefaultTagMessages {
		newV.DefaultTagMessages[tag] = msg
	}

	for path, msg := range v.CustomFieldMessages {
		newV.CustomFieldMessages[path] = msg
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

			params := []interface{}{err.Param()}

			message := v.Messages.ResolveMessage(normPath, err.Tag(), params)
			if message == "" {
				message = getMessageByPath(v, err, path)
			}

			valError := ValidationError{
				Field:      err.Field(),
				Path:       path,
				Message:    message,
				Constraint: err.Tag(),
				Param:      err.Param(),
			}
			validationErrors = append(validationErrors, valError)
		}

		return validationErrors
	}

	return nil
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

// SetFieldMessage sets a custom message for a specific field defined by a path.
// A path looks like "user.profile.firstname" or "user.profile.photos[]".
//
// If the validator is using JSON as tag name, then make sure to define the path
// using the json name instead of the field name.
//
// For cases not using JSON as tag name, make sure to match the casing of the field.
// Example: "User.Profile.Firstname"
//
// Setting an index for array fields won't make the message specific to that index.
// Example: "photos[1]" will be normalized into "photos[]"
func (v *Validator) SetFieldMessage(path string, tag string, s string) *Validator {
	path = normalizePath(path)
	v.CustomFieldMessages[path] = s
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
