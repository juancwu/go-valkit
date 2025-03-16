package validations

import (
	"fmt"
	"regexp"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/juancwu/go-valkit/validator"
)

// PasswordOptions holds configuration for password validation requirements
type PasswordOptions struct {
	MinLength          int  // Minimum length of password
	RequireUppercase   bool // Requires at least one uppercase letter
	RequireLowercase   bool // Requires at least one lowercase letter
	RequireDigit       bool // Requires at least one digit
	RequireSpecialChar bool // Requires at least one special character
}

// PasswordOption is a function that configures PasswordOptions
type PasswordOption func(*PasswordOptions)

// DefaultPasswordOptions returns a PasswordOptions struct with default settings
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		MinLength:          8,
		RequireUppercase:   true,
		RequireLowercase:   true,
		RequireDigit:       true,
		RequireSpecialChar: true,
	}
}

// AddPasswordValidation registers password validation with the validator
// It adds a custom validation tag "password" that can be used in struct tags
// Example: `validate:"password"`
func AddPasswordValidation(v *validator.Validator, options PasswordOptions) error {
	// Build error message based on requirements
	errorMsg := fmt.Sprintf("Password must be at least %d characters", options.MinLength)
	requirements := []string{}

	if options.RequireUppercase {
		requirements = append(requirements, "at least one uppercase letter")
	}
	if options.RequireLowercase {
		requirements = append(requirements, "at least one lowercase letter")
	}
	if options.RequireDigit {
		requirements = append(requirements, "at least one number")
	}
	if options.RequireSpecialChar {
		requirements = append(requirements, "at least one special character")
	}

	// Format the error message with all requirements
	if len(requirements) > 0 {
		errorMsg += " and contain "

		for i, req := range requirements {
			if i > 0 {
				if i == len(requirements)-1 {
					errorMsg += " and "
				} else {
					errorMsg += ", "
				}
			}
			errorMsg += req
		}
	}

	// Register the password validation function
	err := v.RegisterValidation("password", func(fl govalidator.FieldLevel) bool {
		password := fl.Field().String()

		// Check length
		if len(password) < options.MinLength {
			return false
		}

		// Check for required character types
		hasUpper := !options.RequireUppercase || regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := !options.RequireLowercase || regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit := !options.RequireDigit || regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial := !options.RequireSpecialChar || regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

		return hasUpper && hasLower && hasDigit && hasSpecial
	})
	if err != nil {
		return err
	}

	// Set default error message for the password validation tag
	v.SetDefaultTagMessage("password", errorMsg)

	return nil
}
