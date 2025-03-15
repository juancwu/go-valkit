package validator

import (
	"strings"

	goval "github.com/go-playground/validator/v10"
)

// getFullPath generates a clean, dot-separated path string from a FieldError's namespace.
// This function extracts the field path from the validator's namespace, removing the
// struct type name prefix and preserving the rest of the path. The resulting path uses
// the field names as determined by the registered tag name function (e.g., JSON field names
// if UseJsonTagName is called).
//
// For example, given a namespace "Struct.User.Address.Street", it returns "User.Address.Street".
func getFullPath(ve goval.FieldError) string {
	namespace := ve.Namespace()
	parts := strings.Split(namespace, ".")
	parts = parts[1:] // Skip the first part which is the struct type name
	return strings.Join(parts, ".")
}
