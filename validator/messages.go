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

// interpolateParams replaces positional placeholders in a message with values from an array
// Placeholders use the format {0}, {1}, {2}, etc.
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
// [0]: Field name
// [1]: Field value (if available, otherwise nil)
// [2]: Constraint parameter (if available, otherwise nil)
func CreateValidationParams(err ValidationError) []interface{} {
	// Always include field name as first parameter
	params := []interface{}{err.Field}

	// Add nil as second parameter (typically would be the field value,
	// but we don't have access to it here)
	params = append(params, nil)

	// Add constraint parameter if it exists
	if err.Param != "" {
		params = append(params, err.Param)
	} else {
		params = append(params, nil)
	}

	return params
}
