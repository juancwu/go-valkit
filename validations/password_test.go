package validations

import (
	"testing"

	"github.com/juancwu/go-valkit/v2/validator"
	"github.com/stretchr/testify/assert"
)

func TestPasswordValidation(t *testing.T) {
	// Test structure with password field
	type User struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"password"`
	}

	tests := []struct {
		name      string
		password  string
		options   PasswordOptions
		expectErr bool
	}{
		{
			name:      "Valid password with default options",
			password:  "StrongP@ss123",
			options:   DefaultPasswordOptions(),
			expectErr: false,
		},
		{
			name:      "Too short password",
			password:  "Short1!",
			options:   DefaultPasswordOptions(),
			expectErr: true,
		},
		{
			name:      "Missing uppercase",
			password:  "weakpass123!",
			options:   DefaultPasswordOptions(),
			expectErr: true,
		},
		{
			name:      "Missing lowercase",
			password:  "STRONGPASS123!",
			options:   DefaultPasswordOptions(),
			expectErr: true,
		},
		{
			name:      "Missing digit",
			password:  "StrongPass!",
			options:   DefaultPasswordOptions(),
			expectErr: true,
		},
		{
			name:      "Missing special character",
			password:  "StrongPass123",
			options:   DefaultPasswordOptions(),
			expectErr: true,
		},
		{
			name:     "Custom options - only lowercase and digit",
			password: "weakpass123",
			options: func() PasswordOptions {
				opts := DefaultPasswordOptions()
				opts.RequireUppercase = false
				opts.RequireSpecialChar = false
				return opts
			}(),
			expectErr: false,
		},
		{
			name:     "Custom length - passing",
			password: "Pass1!",
			options: func() PasswordOptions {
				opts := DefaultPasswordOptions()
				opts.MinLength = 6
				return opts
			}(),
			expectErr: false,
		},
		{
			name:     "Custom length - failing",
			password: "Pass1!",
			options: func() PasswordOptions {
				opts := DefaultPasswordOptions()
				opts.MinLength = 10
				return opts
			}(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validator with password validation
			v := validator.New()
			v.UseJsonTagName()
			err := AddPasswordValidation(v, tt.options)
			assert.NoError(t, err, "Expected no errors after adding password validation")

			// Create user with test password
			user := User{
				Username: "testuser",
				Password: tt.password,
			}

			// Validate user
			err = v.Validate(user)

			// Check results
			if tt.expectErr {
				assert.Error(t, err, "Expected validation to fail")
				if err != nil {
					errors, ok := err.(validator.ValidationErrors)
					assert.True(t, ok, "Expected ValidationErrors type")

					// Find password error
					foundPasswordError := false
					for _, e := range errors {
						if e.Path == "password" && e.Constraint == "password" {
							foundPasswordError = true
							break
						}
					}
					assert.True(t, foundPasswordError, "Expected password validation error")
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass")
			}
		})
	}
}

func TestPasswordErrorMessage(t *testing.T) {
	type User struct {
		Password string `json:"password" validate:"password"`
	}

	// Test various configurations and their resulting error messages
	tests := []struct {
		name           string
		options        PasswordOptions
		expectedPhrase string // Error message should contain this phrase
	}{
		{
			name:           "Default options",
			options:        DefaultPasswordOptions(),
			expectedPhrase: "at least one uppercase letter, at least one lowercase letter, at least one number and at least one special character",
		},
		{
			name: "Custom length only",
			options: func() PasswordOptions {
				opts := DefaultPasswordOptions()
				opts.MinLength = 12
				opts.RequireUppercase = false
				opts.RequireLowercase = false
				opts.RequireDigit = false
				opts.RequireSpecialChar = false
				return opts
			}(),
			expectedPhrase: "at least 12 characters",
		},
		{
			name: "Only digits required",
			options: func() PasswordOptions {
				opts := DefaultPasswordOptions()
				opts.RequireUppercase = false
				opts.RequireLowercase = false
				opts.RequireSpecialChar = false
				return opts
			}(),
			expectedPhrase: "at least one number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validator with password validation
			v := validator.New()
			v.UseJsonTagName()
			err := AddPasswordValidation(v, tt.options)
			assert.NoError(t, err, "Expected no errors after adding password validation")

			// Create user with invalid password to trigger error message
			user := User{
				Password: "a", // Definitely invalid
			}

			// Validate to get error message
			err = v.Validate(user)
			assert.Error(t, err, "Expected validation to fail")

			errors, ok := err.(validator.ValidationErrors)
			assert.True(t, ok, "Expected ValidationErrors type")

			// Find password error message
			var errorMessage string
			for _, e := range errors {
				if e.Path == "password" && e.Constraint == "password" {
					errorMessage = e.Message
					break
				}
			}

			// Check error message contains expected phrase
			assert.Contains(t, errorMessage, tt.expectedPhrase,
				"Error message should contain the expected requirements phrase")
		})
	}
}
