package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
)

// getFullPath generates a clean, dot-separated path string from a FieldError's namespace.
// This function extracts the field path from the validator's namespace, removing the
// struct type name prefix and preserving the rest of the path. The resulting path uses
// the field names as determined by the registered tag name function (e.g., JSON field names
// if UseJsonTagName is called).
//
// For example, given a namespace "Struct.User.Address.Street", it returns "User.Address.Street".
func getFullPath(ve govalidator.FieldError) string {
	namespace := ve.Namespace()
	parts := strings.Split(namespace, ".")
	return strings.Join(parts[1:], ".")
}

// normalizePath standardizes a field path for consistent error key generation.
// It processes a path string by:
//  1. Trimming surrounding whitespace
//  2. Handling empty paths
//  3. Removing all spaces between field names and array indices
//  4. Converting array indices like [0], [1], etc. to the generic form []
//
// This normalization enables array/slice validation errors to be grouped together
// regardless of which specific array index had the error.
func normalizePath(path string) string {
	path = strings.TrimSpace(path)

	if path == "" {
		return ""
	}

	// Remove all spaces between fields/indexing
	path = strings.ReplaceAll(path, " ", "")

	// Remove all digits from brackets
	re := regexp.MustCompile(`\[\d+\]`)
	path = re.ReplaceAllString(path, "[]")

	// Remove any duplicated dots
	re = regexp.MustCompile(`\.+`)
	path = re.ReplaceAllString(path, ".")

	return path
}

// getRawTagMessage extracts validation error messages from struct field tags.
// It first looks for a constraint-specific error message tag (errmsg-{constraint}),
// then falls back to the general error message tag (errmsg) if no constraint-specific tag is found.
//
// Tags follow this format:
//   - errmsg-{constraint}: "Error message for specific constraint"
//   - errmsg: "General error message for any constraint"
//
// The returned message string can include parameter interpolation using {0}, {field}, {param}, etc.
// See the interpolateParams function for details on parameter formatting.
//
// Example:
//
//	type User struct {
//	    Username string `validate:"required,min=3" errmsg-required:"Username is mandatory" errmsg-min:"Username must have at least {param} characters"`
//	    Email    string `validate:"required,email" errmsg:"Email address has an issue"`
//	}
func getRawTagMessage(field reflect.StructField, constraint string) string {
	// Try to get constraint specific error message tag
	message := field.Tag.Get(fmt.Sprintf("errmsg-%s", constraint))

	// Try to get default error message tag
	if message == "" {
		message = field.Tag.Get("errmsg")
	}

	return message
}
