# Validations

This package contains ready-to-use validation functions for go-valkit.

## Password Validation

Password validation provides configurable rule enforcement for password fields.

### Usage

```go
package main

import (
	"fmt"

	"github.com/juancwu/go-valkit/validations"
	"github.com/juancwu/go-valkit/validator"
)

type User struct {
	Username string `validate:"required"`
	Password string `validate:"password"` // Uses password validation
}

func main() {
	v := validator.New()

	// Use default password options (min 8 chars, requires uppercase, lowercase, digit, special char)
	options := validations.DefaultPasswordOptions()
	validations.AddPasswordValidation(v, options)

	// Or customize password validation options
	customOptions := validations.DefaultPasswordOptions()
	customOptions.MinLength = 10
	customOptions.RequireSpecialChar = false
	validations.AddPasswordValidation(v, customOptions)

	// Validate a struct
	user := User{
		Username: "johndoe",
		Password: "weakpass", // Will fail validation
	}

	errors, _ := v.Validate(user)
	if len(errors) > 0 {
		fmt.Println(errors[0].Message) // Will print detailed password requirements
	}
}
```

### Configuration Options

The `PasswordOptions` struct allows you to configure:

- `MinLength`: Minimum length of password (default: 8)
- `RequireUppercase`: Requires at least one uppercase letter (default: true)
- `RequireLowercase`: Requires at least one lowercase letter (default: true)
- `RequireDigit`: Requires at least one digit (default: true)
- `RequireSpecialChar`: Requires at least one special character (default: true)

Error messages are automatically generated based on the configured options.

