package validator

import (
	"testing"

	goval "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Simple path",
			input:    "user.name",
			expected: "user.name",
		},
		{
			name:     "Path with spaces",
			input:    "user . name",
			expected: "user.name",
		},
		{
			name:     "Path with leading/trailing spaces",
			input:    "  user.name  ",
			expected: "user.name",
		},
		{
			name:     "Path with empty segments",
			input:    "user..name",
			expected: "user.name",
		},
		{
			name:     "Path with array indices",
			input:    "users[0].address",
			expected: "users[].address",
		},
		{
			name:     "Path with multi-dimension array",
			input:    "users[0][1]  [123]",
			expected: "users[][][]",
		},
		{
			name:     "Path with multiple array indices",
			input:    "users[1].addresses[2].street",
			expected: "users[].addresses[].street",
		},
		{
			name:     "Complex path with spaces and indices",
			input:    " users[0] . addresses[123] . street ",
			expected: "users[].addresses[].street",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFullPath(t *testing.T) {
	// Create a struct to validate
	type Address struct {
		Street string `validate:"required"`
	}

	type User struct {
		Address Address `validate:"required"`
	}

	type TestStruct struct {
		User User `validate:"required"`
	}

	// Create an instance with an invalid field
	s := TestStruct{
		User: User{
			Address: Address{
				Street: "", // This will fail the required validation
			},
		},
	}

	// Validate and get the field error
	v := goval.New()
	err := v.Struct(s)

	// Check if there is a validation error
	if err == nil {
		t.Fatal("Expected validation error but got nil")
	}

	validationErrors := err.(goval.ValidationErrors)
	if len(validationErrors) == 0 {
		t.Fatal("Expected at least one validation error")
	}

	// Test getFullPath with a real field error
	result := getFullPath(validationErrors[0])

	// The namespace should contain the path to the Street field
	assert.Contains(t, result, "User.Address.Street")
}
