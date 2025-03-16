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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateParams(tt.message, tt.params)
			assert.Equal(t, tt.expected, result)
		})
	}
}
