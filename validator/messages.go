package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationMessageConfig defines messages for a specific path and constraint
type ValidationMessageConfig struct {
	Default     string            // Default message for path
	Constraints map[string]string // Constraint-specific messages
}

// ValidationMessages maps normalized paths to their validation configurations
type ValidationMessages map[string]ValidationMessageConfig

// NewValidationMessages creates a new validation message configuration
func NewValidationMessages() ValidationMessages {
	return make(ValidationMessages)
}

// SetMessage sets a constraint-specific message for a path
func (vm ValidationMessages) SetMessage(path, constraint, message string) {
	config, exists := vm[path]
	if !exists {
		config = ValidationMessageConfig{
			Constraints: make(map[string]string),
		}
	}
	config.Constraints[constraint] = message
	vm[path] = config
}

// SetDefaultMessage sets the default message for a path
func (vm ValidationMessages) SetDefaultMessage(path, message string) {
	config, exists := vm[path]
	if !exists {
		config = ValidationMessageConfig{
			Constraints: make(map[string]string),
		}
	}
	config.Default = message
	vm[path] = config
}

// ResolveMessage gets the appropriate message for a path and constraint.
// This method will return an emtpy string if no message was set for path and constraint.
func (vm ValidationMessages) ResolveMessage(path, constraint string, params []interface{}) string {
	if config, exists := vm[path]; exists {
		if msg, ok := config.Constraints[constraint]; ok {
			return interpolateParams(msg, params)
		}

		if config.Default != "" {
			return interpolateParams(config.Default, params)
		}
	}

	return ""
}

// interpolateParams replaces placeholders in a message with values from an array
// Supports two types of placeholders:
// 1. Positional placeholders: {0}, {1}, {2}, etc.
// 2. Named placeholders: {field}, {value}, {param}, etc.
// Also supports escaped braces with double braces: {{placeholder}} -> {placeholder}
func interpolateParams(message string, params []interface{}) string {
	if params == nil || len(params) == 0 {
		return message
	}

	// Regular expression to find escaped braces (double braces)
	escapedPattern := regexp.MustCompile(`{{([^{}]*?)}}`)

	// Store all escaped sequences for later restoration
	var replacements []struct {
		placeholder string
		replacement string
	}

	// Find all escaped sequences and generate unique placeholders
	result := escapedPattern.ReplaceAllStringFunc(message, func(match string) string {
		// Generate a unique placeholder that's unlikely to occur in the message
		placeholder := fmt.Sprintf("__ESCAPED_BRACE_%d__", len(replacements))

		// Extract content between double braces and create the single-brace version
		inner := match[2 : len(match)-2] // Remove {{ and }}
		replacement := "{" + inner + "}"

		// Store for later restoration
		replacements = append(replacements, struct {
			placeholder string
			replacement string
		}{
			placeholder: placeholder,
			replacement: replacement,
		})

		return placeholder
	})

	// Define named parameter mapping
	namedParams := map[string]interface{}{}
	if len(params) > 0 {
		namedParams["field"] = params[0]
	}
	if len(params) > 1 {
		namedParams["value"] = params[1]
	}
	if len(params) > 2 {
		namedParams["param"] = params[2]
	}

	// Replace named placeholders with parameter values
	for name, value := range namedParams {
		placeholder := fmt.Sprintf("{%s}", name)
		if value != nil {
			stringValue := fmt.Sprintf("%v", value)
			result = strings.Replace(result, placeholder, stringValue, -1)
		} else {
			// Replace nil values with empty string
			result = strings.Replace(result, placeholder, "", -1)
		}
	}

	// Replace positional placeholders with parameter values
	for i, param := range params {
		placeholder := fmt.Sprintf("{%d}", i)
		stringValue := fmt.Sprintf("%v", param)
		result = strings.Replace(result, placeholder, stringValue, -1)
	}

	// Restore all escaped placeholders with their single-brace versions
	for _, item := range replacements {
		result = strings.Replace(result, item.placeholder, item.replacement, 1)
	}

	return result
}

// CreateValidationParams creates a slice of parameters from validation error data
// Returns parameters in a consistent order:
// [0]/field: Field name
// [1]/value: Field value (if available, otherwise nil)
// [2]/param: Constraint parameter (if available, otherwise nil)
//
// These can be used with both positional {0}, {1}, {2} and named {field}, {value}, {param} placeholders
func CreateValidationParams(err ValidationError) []interface{} {
	// Always include field name as first parameter and field value as second parameter
	params := []interface{}{err.Field, err.Actual}

	// Add constraint parameter if it exists
	if err.Param != "" {
		params = append(params, err.Param)
	} else {
		params = append(params, nil)
	}

	return params
}
