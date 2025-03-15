package validator

import (
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
