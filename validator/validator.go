package validator

import (
	govalidator "github.com/go-playground/validator/v10"
)

type Validator struct {
	validator           *govalidator.Validate
	DefaultMessage      string
	DefaultTagMessages  map[string]string
	CustomFieldMessages map[string]string
}

// New creates a new Validator instance with default configuration.
func New() *Validator {
	v := govalidator.New()

	return &Validator{
		validator:           v,
		DefaultMessage:      "Invalid value",
		DefaultTagMessages:  make(map[string]string),
		CustomFieldMessages: make(map[string]string),
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
			message := getMessageByPath(v, err, path)
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
func (v *Validator) SetFieldMessage(path string, s string) *Validator {
	path = normalizePath(path)
	v.CustomFieldMessages[path] = s
	return v
}

// getMessageByPath gets the error message if matches a path in custom messages, then
// fallback to using tag default messages. If nothing matches, the default message is used.
// It handles path such as:
// - "A.B"
// - "A[]" or "A.B[]"
func getMessageByPath(v *Validator, err govalidator.FieldError, path string) string {
	// Normalize path to keep path-key consistent for lookup
	path = normalizePath(path)

	// Match path base messages
	if msg, ok := v.CustomFieldMessages[path]; ok {
		return msg
	}

	// Match default message by tag
	if msg, ok := v.DefaultTagMessages[err.Tag()]; ok {
		return msg
	}

	// Use default message
	return v.DefaultMessage
}
