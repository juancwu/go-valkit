package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	JSON_TAG = "json"
	FORM_TAG = "form"
)

// Validator represents a validator with custom error messages
type Validator struct {
	validator       *validator.Validate
	tagName         string            // Tag name to extract field names (e.g., "json", "form")
	defaultMessages map[string]string // Default messages for validation tags
	customMessages  map[string]string // Custom messages for specific field paths
}

// New creates a new Validator instance
func New() *Validator {
	return NewWithConfig(DefaultConfig())
}

// SetDefaultMessage sets a default message for a validation tag
func (v *Validator) SetDefaultMessage(tag, message string) {
	v.defaultMessages[tag] = message
}

// SetCustomMessage sets a custom message for a specific field path
func (v *Validator) SetCustomMessage(path, message string) {
	v.customMessages[path] = message
}

// Validate validates a struct and returns validation errors
func (v *Validator) Validate(s interface{}) (ValidationErrors, error) {
	// Extract field names from struct tags
	fieldNames := v.extractFieldNames(s)

	// Validate the struct
	err := v.validator.Struct(s)
	if err == nil {
		return nil, nil
	}

	// Parse validation errors
	validationErrors := make(ValidationErrors)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			// Get the field path
			fieldPath := v.getFieldPath(e)

			// Get the field name
			fieldName := v.getFieldName(e, fieldNames)

			// Get the error message
			message := v.getErrorMessage(fieldPath, e)

			// Add the error to the validation errors
			validationErrors[fieldPath] = &ValidationError{
				Field:   fieldName,
				Tag:     e.Tag(),
				Value:   e.Value(),
				Message: message,
				Param:   e.Param(),
			}
		}
	}

	return validationErrors, err
}

// extractFieldNames extracts field names from struct tags
func (v *Validator) extractFieldNames(s interface{}) map[string]string {
	fieldNames := make(map[string]string)

	if v.tagName == "" {
		return fieldNames
	}

	// Use reflection to extract field names
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fieldNames
	}

	v.extractFieldNamesRecursive(val.Type(), "", fieldNames)

	return fieldNames
}

// extractFieldNamesRecursive recursively extracts field names from struct tags
func (v *Validator) extractFieldNamesRecursive(t reflect.Type, prefix string, fieldNames map[string]string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the field path
		fieldPath := field.Name
		if prefix != "" {
			fieldPath = prefix + "." + fieldPath
		}

		// Get the tag value
		tag := field.Tag.Get(v.tagName)
		if tag != "" {
			// Extract the field name from the tag
			parts := strings.Split(tag, ",")
			if len(parts) > 0 && parts[0] != "" {
				fieldNames[fieldPath] = parts[0]
			}
		}

		// Recursively extract field names for nested structs
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			v.extractFieldNamesRecursive(fieldType, fieldPath, fieldNames)
		} else if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			// For slices/arrays, check the element type
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}

			if elemType.Kind() == reflect.Struct {
				v.extractFieldNamesRecursive(elemType, fieldPath+"[]", fieldNames)
			}
		}
	}
}

// getFieldPath gets the field path for a validation error
func (v *Validator) getFieldPath(e validator.FieldError) string {
	namespace := e.Namespace()

	// Remove the struct name from the namespace
	parts := strings.Split(namespace, ".")
	if len(parts) > 1 {
		parts = parts[1:]
	}

	// Replace array/slice indices with [] using regex
	path := strings.Join(parts, ".")
	re := regexp.MustCompile(`\[\d+\]`)
	path = re.ReplaceAllString(path, "[]")

	return path
}

// getFieldName gets the field name from the extracted field names
func (v *Validator) getFieldName(e validator.FieldError, fieldNames map[string]string) string {
	// Get the field path
	fieldPath := v.getFieldPath(e)

	// Check if there's a custom field name for this path
	if name, ok := fieldNames[fieldPath]; ok {
		return name
	}

	// Return the original field name
	return e.Field()
}

// getErrorMessage gets the error message for a field path and validation error
func (v *Validator) getErrorMessage(fieldPath string, e validator.FieldError) string {
	// Check if there's a custom message for this field path
	if message, ok := v.customMessages[fieldPath]; ok {
		return message
	}

	// Check if there's a default message for this tag
	if message, ok := v.defaultMessages[e.Tag()]; ok {
		// Replace placeholders in the message
		message = strings.ReplaceAll(message, "{0}", e.Param())
		message = strings.ReplaceAll(message, "{field}", e.Field())

		return message
	}

	// Return the default error message
	return fmt.Sprintf("Validation failed for field %s", fieldPath)
}
