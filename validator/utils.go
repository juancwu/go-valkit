package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
// For slices and arrays, it preserves the indices, e.g., "User.Addresses[0].Street".
func getFullPath(ve govalidator.FieldError) string {
	// Get the field error's namespace
	namespace := ve.Namespace()

	// The namespace includes the top-level struct type name at the beginning,
	// so we need to remove that to get the actual field path
	parts := strings.SplitN(namespace, ".", 2)
	if len(parts) < 2 {
		return namespace // Return as is if there's no dot
	}

	// parts[1] now contains the actual field path without the top-level struct name
	// This will include array/slice indices in the format: "Items[0].Name"
	return parts[1]
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

// getStructFieldFromNamespace traverses the struct hierarchy to find the field specified by the namespace.
// It takes:
// - structType: The root struct type to start searching from
// - namespace: The full struct namespace (e.g. "User.Address.Street", "User.Items[0].Name", "User.Metadata[key]")
// - leafFieldName: The name of the leaf field in the namespace (e.g. "Street", "Name", "key")
//
// Returns:
// - The reflect.StructField of the field if found
// - A boolean indicating whether the field was found
func getStructFieldFromNamespace(structType reflect.Type, namespace string, leafFieldName string) (reflect.StructField, bool) {
	if structType == nil {
		return reflect.StructField{}, false
	}

	// Ensure we're working with a struct type
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}

	// If it's a nested field, we need to parse the namespace to find it
	// The namespace is "StructType.Field1.Field2.Field3"
	parts := strings.Split(namespace, ".")
	if len(parts) <= 1 {
		// If no dots in the namespace, just try to find the field directly
		field, found := structType.FieldByName(leafFieldName)
		return field, found
	}

	current := structType

	// Go through each part of the namespace except the last one (which is the target field)
	for i := 1; i < len(parts); i++ {
		part := parts[i]

		// Handle arrays/slices with indices: "Field[0]" -> "Field"
		// Or maps with keys: "Field[key]" -> "Field"
		fieldName := part
		arrayIndex := -1
		isMapKey := false

		// Check for array/slice index or map key
		reArrayIndex := regexp.MustCompile(`^(\w+)\[(\d+)\]$`)
		arrayMatches := reArrayIndex.FindStringSubmatch(part)

		if len(arrayMatches) == 3 {
			// It's an array/slice with numeric index
			fieldName = arrayMatches[1]
			index, err := strconv.Atoi(arrayMatches[2])
			if err == nil {
				arrayIndex = index
			}
		} else {
			// Check for map key pattern: Field[key]
			reMapKey := regexp.MustCompile(`^(\w+)\[([^\]]+)\]$`)
			mapMatches := reMapKey.FindStringSubmatch(part)
			if len(mapMatches) == 3 {
				fieldName = mapMatches[1]
				isMapKey = true
			}
		}

		// If this is the leaf field, return it directly
		if fieldName == leafFieldName && i == len(parts)-1 {
			field, found := current.FieldByName(fieldName)
			return field, found
		}

		// Find the field by name to continue traversing
		field, found := current.FieldByName(fieldName)
		if !found {
			return reflect.StructField{}, false
		}

		fieldType := field.Type

		if arrayIndex >= 0 && (fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice) {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Handle map values
		if fieldType.Kind() == reflect.Map {
			// If this is directly the target field and it has error message tags, return it
			if isMapKey && i == len(parts)-1 {
				return field, true
			}

			fieldType = fieldType.Elem()
		}

		if i < len(parts)-1 && fieldType.Kind() != reflect.Struct {
			return reflect.StructField{}, false
		}

		current = fieldType
	}

	if current.Kind() == reflect.Struct {
		field, found := current.FieldByName(leafFieldName)
		return field, found
	}

	return reflect.StructField{}, false
}
