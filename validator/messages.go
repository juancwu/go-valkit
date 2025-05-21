package validator

import (
	"fmt"
	"regexp"
	"strconv"
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
func (vm ValidationMessages) ResolveMessage(path, constraint string, params []interface{}, customParams ...CustomParams) string {
	if config, exists := vm[path]; exists {
		if msg, ok := config.Constraints[constraint]; ok {
			return interpolateParams(msg, params, customParams...)
		}

		if config.Default != "" {
			return interpolateParams(config.Default, params, customParams...)
		}
	}

	return ""
}

// interpolateParams replaces placeholders in a message with values from an array
// Supports three types of placeholders:
// 1. Positional placeholders: {0}, {1}, {2}, etc.
// 2. Standard named placeholders: {field}, {value}, {param}, etc.
// 3. Custom named parameters: any name defined in CustomParams
// Also supports escaped braces with double braces: {{placeholder}} -> {placeholder}
//
// Important rules for parameters:
// - Numbers with leading zeros (e.g. {000}) are treated as literals, not parameter indices
// - Parameter names starting with digits (e.g. {0name}) are treated as literals
// - Parameter names can contain digits if they don't start with one (e.g. {name123})
func interpolateParams(message string, params []interface{}, customParams ...CustomParams) string {
	if (params == nil || len(params) == 0) && len(customParams) == 0 {
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

	// Add custom parameters if provided
	if len(customParams) > 0 {
		for _, cp := range customParams {
			for name, value := range cp {
				// Custom parameters override default ones
				namedParams[name] = value
			}
		}
	}

	// Replace named placeholders with parameter values
	// Use a regular expression to match exact placeholder patterns
	placeholderRegex := regexp.MustCompile(`{([^{}]+)}`)
	result = placeholderRegex.ReplaceAllStringFunc(result, func(placeholder string) string {
		// Extract the parameter name without the braces
		paramName := placeholder[1 : len(placeholder)-1]

		// Special case: Skip replacing parameter names that begin with a digit
		// This is to avoid confusion with positional parameters and ensure
		// placeholders like {0name} remain as they are
		if len(paramName) > 0 && paramName[0] >= '0' && paramName[0] <= '9' {
			return placeholder
		}

		// Check if this parameter exists in the named parameters
		if value, exists := namedParams[paramName]; exists {
			if value != nil {
				return fmt.Sprintf("%v", value)
			}
			return "" // Replace nil values with empty string
		}

		// If not found, return the original placeholder
		return placeholder
	})

	// Replace positional placeholders with parameter values
	if params != nil {
		// Define a strict pattern for positional parameters
		// Only match {0}, {1}, {2}, etc. as positional parameters
		// The pattern specifically excludes leading zeros to avoid {000000000} being treated as positional
		positionalRegex := regexp.MustCompile(`{([0-9])}` + "|" + `{([1-9][0-9]+)}`)
		result = positionalRegex.ReplaceAllStringFunc(result, func(placeholder string) string {
			// Extract the index without the braces
			indexStr := placeholder[1 : len(placeholder)-1]

			// Skip if the string starts with a leading zero and has more than one digit
			if len(indexStr) > 1 && indexStr[0] == '0' {
				return placeholder
			}

			index, err := strconv.Atoi(indexStr)

			// Check if this is a valid index in our params array
			if err == nil && index >= 0 && index < len(params) {
				return fmt.Sprintf("%v", params[index])
			}

			// If not found or invalid, return the original placeholder
			return placeholder
		})
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
// These parameters are automatically available in all validation messages and can be referenced in two ways:
// 1. Using positional placeholders: {0}, {1}, {2}
// 2. Using named placeholders: {field}, {value}, {param}
//
// Example:
//
//	// These two formats are equivalent
//	v.SetConstraintMessage("username", "min", "{0} must be at least {2} characters")
//	v.SetConstraintMessage("email", "required", "{field} is required for validation")
//
// Note: Custom parameters added via AddCustomParam() can be used alongside these standard parameters
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
