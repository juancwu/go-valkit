package validator

import (
	"reflect"
	"strings"
)

// UseJsonTagName configures the validator to use the JSON tag names in error messages
// and field paths instead of the Go struct field names.
//
// This makes error messages more useful when working with JSON APIs, as the field names
// in error messages will match the names used in JSON requests/responses.
//
// Example:
//   type User struct {
//       FirstName string `json:"first_name" validate:"required"`
//   }
//
// Without UseJsonTagName: error path would be "FirstName"
// With UseJsonTagName: error path would be "first_name"
func (v *Validator) UseJsonTagName() {
	v.validator.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			// If JSON tag is "-" (meaning "don't include in JSON"), use the actual field name
			return field.Name
		}
		return name
	})
}
