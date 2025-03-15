package validator

import (
	"fmt"
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
func (vm ValidationMessages) ResolveMessage(path, constraint string, params map[string]interface{}) string {
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

// interpolateParams replaces placeholders in message with values
func interpolateParams(message string, params map[string]interface{}) string {
	if params == nil {
		return message
	}

	result := message
	for key, value := range params {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.Replace(result, placeholder, replacement, -1)
	}

	return result
}
