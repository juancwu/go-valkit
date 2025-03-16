# go-valkit

A powerful and flexible Go validation library that enhances [go-playground/validator](https://github.com/go-playground/validator)
with better error handling, customizable messages, and framework integrations.

[![Enforce Proper Code Formatting](https://github.com/juancwu/go-valkit/actions/workflows/code-formatting.yml/badge.svg)](https://github.com/juancwu/go-valkit/actions/workflows/code-formatting.yml)
[![Go Build Check](https://github.com/juancwu/go-valkit/actions/workflows/build-check.yml/badge.svg)](https://github.com/juancwu/go-valkit/actions/workflows/build-check.yml)

## Features

- **Enhanced Error Handling**: Structured validation errors with field paths, constraints, and parameters
- **Customizable Messages**: Set default, tag-specific, or field-specific validation messages
- **Message Interpolation**: Support for variable placeholders in error messages
- **Framework Integration**: Easy integration with Echo and extensible for other frameworks
- **JSON Field Name Support**: Use JSON field names in error messages for better API responses

## Installation

```bash
go get github.com/juancwu/go-valkit
```

## Quick Start

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/juancwu/go-valkit/validator"
)

type User struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=18"`
}

func main() {
	// Create a new validator
	v := validator.New()
	v.UseJsonTagName() // Use JSON field names in error messages

	// Set custom validation messages
	v.SetDefaultTagMessage("required", "This field is required")
	v.SetDefaultTagMessage("min", "Minimum value is {2}")
	v.SetDefaultTagMessage("email", "Must be a valid email address")

	// Create a user with validation errors
	user := User{
		Username: "jo", // too short
		Email:    "not-an-email",
		Age:      16,   // too young
	}

	// Validate the user
	err := v.Validate(user)
	if err != nil {
		// Type assertion to get structured validation errors
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				fmt.Printf("Field: %s, Error: %s\n", e.Path, e.Message)
			}
		}
	}
}
```

### Echo Framework Integration

```go
package main

import (
	"github.com/juancwu/go-valkit/integrations/echo"
	"github.com/juancwu/go-valkit/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func main() {
	e := echo.New()

	// Configure Echo with go-valkit
	echo.ConfigureWithOptions(e, func(v *validator.Validator) {
		v.SetDefaultTagMessage("required", "This field is required")
		v.SetDefaultTagMessage("email", "Must be a valid email address")
		v.SetDefaultTagMessage("min", "Must be at least {2} characters long")
	})

	e.POST("/login", func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		if err := c.Validate(&req); err != nil {
			// Error handling is already configured
			return err
		}

		// Process valid request...
		return c.JSON(http.StatusOK, map[string]string{
			"status": "success",
			"message": "Login successful",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
```

## Advanced Usage

### Field-Specific Messages

```go
v := validator.New()
v.UseJsonTagName()

// Set specific message for email field when it fails email validation
v.SetConstraintMessage("user.email", "email", "Please enter a valid email address")

// Set specific message for username field when it fails required validation
v.SetConstraintMessage("username", "required", "Username cannot be empty")

// Set default message for a specific field
v.SetPathDefaultMessage("age", "Age must be valid")
```

### Message Interpolation

Messages support parameter interpolation:

- `{0}`: Field name
- `{1}`: Field value (when available)
- `{2}`: Constraint parameter (e.g., "8" for min=8)

```go
v.SetDefaultTagMessage("min", "{0} must be at least {2} characters long")
v.SetDefaultTagMessage("required", "{0} is required")
```

### Custom Error Formatting

You can implement custom error formatters for HTTP responses:

```go
type MyErrorFormatter struct{}

func (f *MyErrorFormatter) Format(ve validator.ValidationErrors) interface{} {
	// Format errors as needed
	// ...
}

// Use with Echo
validator := echo.NewValidator()
validator.WithFormatter(&MyErrorFormatter{})
```

## License

This project is licensed under the [MIT License](LICENSE).
