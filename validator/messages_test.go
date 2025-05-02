package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationMessages(t *testing.T) {
	messages := NewValidationMessages()

	// Resolve empty when no messages defined
	assert.Empty(t, messages.ResolveMessage("fake", "fake", nil))

	// Add default message
	messages.SetDefaultMessage("fake", "default message")
	assert.Equal(t, "default message", messages.ResolveMessage("fake", "fake", nil))

	// Add specific message
	messages.SetMessage("fake", "fake", "specific message")
	assert.Equal(t, "specific message", messages.ResolveMessage("fake", "fake", nil))

	// Add default message with params
	messages.SetDefaultMessage("fake2", "{0} message")
	assert.Equal(t, "param message", messages.ResolveMessage("fake2", "fake", []interface{}{"param"}))

	// Add specific message with params
	messages.SetMessage("fake", "fake", "Hello {0}, you are {1}")
	assert.Equal(t, "Hello Tim, you are good", messages.ResolveMessage("fake", "fake", []interface{}{"Tim", "good"}))

	// Test with custom parameters
	customParams := make(CustomParams)
	customParams["appName"] = "TestApp"
	customParams["version"] = "1.0.0"

	messages.SetMessage("app", "info", "Welcome to {appName} version {version}")
	assert.Equal(t, "Welcome to TestApp version 1.0.0",
		messages.ResolveMessage("app", "info", nil, customParams))

	// Test custom params with standard params
	messages.SetMessage("user", "login", "User {0} logged into {appName} v{version}")
	assert.Equal(t, "User john logged into TestApp v1.0.0",
		messages.ResolveMessage("user", "login", []interface{}{"john"}, customParams))
}

func TestInterpolateParams(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		params   []interface{}
		expected string
	}{
		{
			name:     "Nil params",
			message:  "This is a {test} message with {placeholders}",
			params:   nil,
			expected: "This is a {test} message with {placeholders}",
		},
		{
			name:     "Empty params",
			message:  "This is a {test} message with {placeholders}",
			params:   []interface{}{},
			expected: "This is a {test} message with {placeholders}",
		},
		{
			name:     "Single replacement",
			message:  "Value must be at least {0}",
			params:   []interface{}{5},
			expected: "Value must be at least 5",
		},
		{
			name:     "Multiple replacements",
			message:  "Value must be between {0} and {1}",
			params:   []interface{}{5, 10},
			expected: "Value must be between 5 and 10",
		},
		{
			name:     "No placeholders",
			message:  "This message has no placeholders",
			params:   []interface{}{"value"},
			expected: "This message has no placeholders",
		},
		{
			name:     "Placeholder with no corresponding param",
			message:  "Value must be at least {000000000}",
			params:   []interface{}{10},
			expected: "Value must be at least {000000000}",
		},
		{
			name:     "Repeated placeholders",
			message:  "The {0} field is invalid. {0} must be valid.",
			params:   []interface{}{"email"},
			expected: "The email field is invalid. email must be valid.",
		},
		{
			name:     "Different value types - string",
			message:  "The field {0} is invalid",
			params:   []interface{}{"email"},
			expected: "The field email is invalid",
		},
		{
			name:     "Different value types - int",
			message:  "Minimum length is {0}",
			params:   []interface{}{5},
			expected: "Minimum length is 5",
		},
		{
			name:     "Different value types - float",
			message:  "The value must be at least {0}",
			params:   []interface{}{5.75},
			expected: "The value must be at least 5.75",
		},
		{
			name:     "Different value types - bool",
			message:  "Required: {0}",
			params:   []interface{}{true},
			expected: "Required: true",
		},
		{
			name:     "Mixed types",
			message:  "Field {0} must be between {1} and {2} characters and is required: {3}",
			params:   []interface{}{"username", 3, 20, true},
			expected: "Field username must be between 3 and 20 characters and is required: true",
		},
		{
			name:     "Complex placeholder names",
			message:  "Error in {0} and {1}",
			params:   []interface{}{"username", "string"},
			expected: "Error in username and string",
		},

		// Escaped braces cases
		{
			name:     "Simple escaped braces",
			message:  "Use {{param}} syntax for placeholder {0}",
			params:   []interface{}{"value"},
			expected: "Use {param} syntax for placeholder value",
		},
		{
			name:     "Multiple escaped braces",
			message:  "Syntax: {{index}}, {0}, {1}",
			params:   []interface{}{"value1", "value2"},
			expected: "Syntax: {index}, value1, value2",
		},
		{
			name:     "Complex mixed case",
			message:  "Error in {{json.field}} structure. Field {0} value {1} is invalid. Use {{example}} format.",
			params:   []interface{}{"username", "test@123"},
			expected: "Error in {json.field} structure. Field username value test@123 is invalid. Use {example} format.",
		},
		{
			name:     "Mixed numeric placeholders and escaped braces",
			message:  "Error in field {0}: must match format {{0}}-{1}-{2}",
			params:   []interface{}{"date", "yyyy", "mm", "dd"},
			expected: "Error in field date: must match format {0}-yyyy-mm",
		},

		// Named parameter cases
		{
			name:     "Named parameter - field",
			message:  "The {field} must be valid",
			params:   []interface{}{"email", "test@example.com", ""},
			expected: "The email must be valid",
		},
		{
			name:     "Named parameter - value",
			message:  "The value {value} is invalid",
			params:   []interface{}{"email", "test@example.com", ""},
			expected: "The value test@example.com is invalid",
		},
		{
			name:     "Named parameter - param",
			message:  "Value must be at least {param}",
			params:   []interface{}{"length", "", "8"},
			expected: "Value must be at least 8",
		},
		{
			name:     "Mixed named and positional parameters",
			message:  "Field {field} with value {1} must be at least {param} characters",
			params:   []interface{}{"password", "abc", "8"},
			expected: "Field password with value abc must be at least 8 characters",
		},
		{
			name:     "Named parameters with nil values",
			message:  "Field {field} with value {value} must be {param}",
			params:   []interface{}{"email", nil, nil},
			expected: "Field email with value  must be ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateParams(tt.message, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test custom parameters separately
	t.Run("Custom parameters only", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["appName"] = "TestApp"
		customParams["version"] = "1.0.0"

		message := "Welcome to {appName} version {version}"
		expected := "Welcome to TestApp version 1.0.0"

		result := interpolateParams(message, nil, customParams)
		assert.Equal(t, expected, result)
	})

	t.Run("Custom parameters with standard params", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["appName"] = "TestApp"
		customParams["domain"] = "example.com"

		message := "User {field} registered at {domain} using {appName}"
		params := []interface{}{"john.doe"}
		expected := "User john.doe registered at example.com using TestApp"

		result := interpolateParams(message, params, customParams)
		assert.Equal(t, expected, result)
	})

	t.Run("Custom parameters override standard params", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["field"] = "CUSTOM_FIELD"
		customParams["appName"] = "TestApp"

		message := "Field {field} in {appName}"
		params := []interface{}{"username"}
		expected := "Field CUSTOM_FIELD in TestApp"

		result := interpolateParams(message, params, customParams)
		assert.Equal(t, expected, result)
	})

	t.Run("Multiple custom parameter sources", func(t *testing.T) {
		customParams1 := make(CustomParams)
		customParams1["appName"] = "TestApp"

		customParams2 := make(CustomParams)
		customParams2["version"] = "2.0.0"
		customParams2["appName"] = "OverrideApp" // This should take precedence

		message := "Welcome to {appName} version {version}"
		expected := "Welcome to OverrideApp version 2.0.0"

		result := interpolateParams(message, nil, customParams1, customParams2)
		assert.Equal(t, expected, result)
	})

	// Test escaped braces with named parameters
	t.Run("Named parameters with escaped braces", func(t *testing.T) {
		message := "The {field} must use format {{fieldFormat}}"
		params := []interface{}{"email", "test@example.com", ""}
		expected := "The email must use format {fieldFormat}"

		result := interpolateParams(message, params)
		assert.Equal(t, expected, result)
	})

	// Test escaped braces with custom parameters
	t.Run("Custom parameters with escaped braces", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["appName"] = "TestApp"
		customParams["format"] = "JSON"

		message := "{{appName}} requires {{format}} format, actual: {appName}/{format}"
		expected := "{appName} requires {format} format, actual: TestApp/JSON"

		result := interpolateParams(message, nil, customParams)
		assert.Equal(t, expected, result)
	})

	// Test mixed escaped braces with both named and custom parameters
	t.Run("Mixed escaped and non-escaped named and custom parameters", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["appName"] = "TestApp"
		customParams["supportEmail"] = "support@example.com"

		message := "Field {field} with value {{value}} must match {appName} format. Contact {{supportEmail}}."
		params := []interface{}{"email", "test@invalid.com", ""}
		expected := "Field email with value {value} must match TestApp format. Contact {supportEmail}."

		result := interpolateParams(message, params, customParams)
		assert.Equal(t, expected, result)
	})

	// Test with parameter name that contains braces
	t.Run("Custom parameter name containing braces", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["app{Name}"] = "TestApp" // Parameter name contains braces

		message := "Welcome to {app{Name}}"
		expected := "Welcome to {app{Name}}" // Should not replace this

		result := interpolateParams(message, nil, customParams)
		assert.Equal(t, expected, result)
	})

	// Test parameter names starting with digits
	t.Run("Parameter name starting with digit", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["0name"] = "TestValue" // Valid parameter name starting with digit

		// The placeholder {0name} should not be replaced
		message := "Value is {0name}"
		expected := "Value is {0name}"

		result := interpolateParams(message, nil, customParams)
		assert.Equal(t, expected, result)

		// But a custom parameter with that name should be usable in other contexts
		message2 := "The parameter {param} has name 0name"
		expected2 := "The parameter TestValue has name 0name"

		// Adding it as the 3rd parameter (which maps to {param})
		result2 := interpolateParams(message2, []interface{}{"field", "value", "TestValue"}, nil)
		assert.Equal(t, expected2, result2)
	})

	// Test parameter names containing digits (but not at the start)
	t.Run("Parameter name containing digits", func(t *testing.T) {
		customParams := make(CustomParams)
		customParams["name123"] = "TestValue"            // Parameter name with digits
		customParams["prefix_2_suffix"] = "AnotherValue" // Parameter name with digits in the middle

		// The placeholder {name123} should be replaced
		message := "Value is {name123}"
		expected := "Value is TestValue"

		result := interpolateParams(message, nil, customParams)
		assert.Equal(t, expected, result)

		// The placeholder {prefix_2_suffix} should also be replaced
		message2 := "Another value is {prefix_2_suffix}"
		expected2 := "Another value is AnotherValue"

		result2 := interpolateParams(message2, nil, customParams)
		assert.Equal(t, expected2, result2)

		// Both can be used in the same message
		message3 := "Values: {name123} and {prefix_2_suffix}"
		expected3 := "Values: TestValue and AnotherValue"

		result3 := interpolateParams(message3, nil, customParams)
		assert.Equal(t, expected3, result3)
	})
}

func TestCreateValidationParams(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected []interface{}
	}{
		{
			name: "Basic fields",
			err: ValidationError{
				Field:      "name",
				Path:       "user.name",
				Constraint: "required",
				Param:      "",
			},
			expected: []interface{}{"name", nil, nil},
		},
		{
			name: "With constraint param",
			err: ValidationError{
				Field:      "age",
				Path:       "user.age",
				Constraint: "min",
				Param:      "18",
			},
			expected: []interface{}{"age", nil, "18"},
		},
		{
			name: "With array index",
			err: ValidationError{
				Field:      "email",
				Path:       "users[2].email",
				Constraint: "email",
				Param:      "",
			},
			expected: []interface{}{"email", nil, nil},
		},
		{
			name: "With all fields",
			err: ValidationError{
				Field:      "score",
				Path:       "users[3].scores[1]",
				Constraint: "max",
				Param:      "100",
			},
			expected: []interface{}{"score", nil, "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateValidationParams(tt.err)

			// Check that we have the expected number of parameters
			assert.Equal(t, len(tt.expected), len(result),
				"Expected %d parameters, got %d", len(tt.expected), len(result))

			// Check each parameter in the expected order
			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected, result[i],
						"Parameter at position %d should be %v but got %v", i, expected, result[i])
				}
			}
		})
	}
}
