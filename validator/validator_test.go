package validator

import (
	"reflect"
	"strings"
	"testing"

	goval "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// TestValidate contains test cases for the Validate function and related functionality
func TestValidate(t *testing.T) {
	// Test that validation correctly reports the expected number of validation errors
	t.Run("Correct ValidationErrors Count", func(t *testing.T) {
		v := New()

		v.UseJsonTagName()

		// Item struct with multiple validation rules
		type Item struct {
			Key     string   `json:"key" validate:"required,alphanum"`
			Value   string   `json:"value" validate:"required,min=3"`
			Details []string `json:"details" validate:"required,min=1,max=2,email"`
		}

		// Parent struct with nested validation rules
		type Teststruct struct {
			Name  string `json:"name" validate:"required"`
			Items []Item `json:"items" validate:"required,min=1,dive"`
		}

		// Create a test struct with intentional validation errors:
		// - Empty Name (fails required)
		// - Key with non-alphanumeric chars (fails alphanum)
		// - Value too short (fails min=3)
		// - Details with invalid email (fails email)
		s := Teststruct{
			Name: "",
			Items: []Item{
				{
					Key:     "]]]]]]",
					Value:   "d",
					Details: []string{"kdjasdlkjlkj"},
				},
			},
		}

		err := v.Validate(s)
		errs, ok := err.(ValidationErrors)
		assert.True(t, ok, "Should be TRUE")

		// We expect exactly 4 validation errors (name, key, value, and details)
		assert.Len(t, errs, 4, "Should have 4 validation errors")
	})

	// Test that nested field paths are correctly generated
	t.Run("Full path to field error", func(t *testing.T) {
		type SubItem struct {
			Name string `json:"name" validate:"required,min=5"`
		}

		type Item struct {
			Sub   SubItem  `json:"sub" validate:"required"`
			Array []string `json:"array" validate:"required,min=1,dive,alpha"`
		}

		v := New()
		v.UseJsonTagName()

		// Create a test item with validation errors:
		// - Sub.Name is too short (fails min=5)
		// - Array[1] contains a non-alpha character (fails alpha)
		item := Item{
			Sub: SubItem{
				Name: "hey", // Too short, min=5
			},
			Array: []string{"a", "1"}, // "1" fails alpha validation
		}

		err := v.Validate(item)
		errs, ok := err.(ValidationErrors)
		assert.True(t, ok, "Should be of type ValidationErrors")

		assert.Len(t, errs, 2, "Should only have 2 errors")

		// First error should be about sub.name failing min=5
		assert.Equal(t, "sub.name", errs[0].Path)
		assert.Equal(t, "min", errs[0].Constraint)
		assert.Equal(t, "5", errs[0].Param)

		// Second error should be about array[1] failing alpha
		assert.Equal(t, "array[1]", errs[1].Path)
		assert.Equal(t, "alpha", errs[1].Constraint)
		assert.Equal(t, "", errs[1].Param)
	})
}

func TestUseJsonTagName(t *testing.T) {
	v := New()
	v.UseJsonTagName()

	type Item struct {
		Name string `json:"name" validate:"required"`
	}

	i := Item{Name: ""}

	err := v.Validate(i)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	assert.Equal(t, "name", errs[0].Path)
	assert.Equal(t, "required", errs[0].Constraint)
}

func TestDefaultTagName(t *testing.T) {
	v := New()

	type Item struct {
		Name string `json:"name" validate:"required"`
	}

	i := Item{Name: ""}

	err := v.Validate(i)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	assert.Equal(t, "Name", errs[0].Path)
	assert.Equal(t, "required", errs[0].Constraint)
}

func TestDefaultMessage(t *testing.T) {
	type Item struct {
		Name string `json:"name" validate:"required"`
	}

	i := Item{Name: ""}

	tests := []struct {
		Name   string
		Input  Item
		Output string
		Action func(v *Validator)
	}{
		{
			Name:   "No default message set",
			Input:  i,
			Output: "Invalid value",
		},
		{
			Name:   "No default message set",
			Input:  i,
			Output: "New default message",
			Action: func(v *Validator) {
				v.SetDefaultMessage("New default message")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			v := New()

			if tc.Action != nil {
				tc.Action(v)
			}

			err := v.Validate(tc.Input)
			errs, ok := err.(ValidationErrors)
			assert.True(t, ok, "Should be of type ValidationErrors")
			assert.Equal(t, tc.Output, errs[0].Message)
		})
	}
}

func TestDefaultTagMessage(t *testing.T) {
	type Item struct {
		Name string `json:"name" validate:"required"`
		Age  uint   `json:"age" validate:"required,min=18"`
	}

	i := Item{Name: "", Age: 13}

	v := New()
	v.UseJsonTagName()

	requiredTagMsg := "This field is required"
	minTagMsg := "This field has minimum set"

	v.
		SetDefaultTagMessage("required", requiredTagMsg).
		SetDefaultTagMessage("min", minTagMsg)

	err := v.Validate(i)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	for _, ve := range errs {
		switch ve.Path {
		case "name":
			assert.Equal(t, requiredTagMsg, ve.Message)
		case "age":
			assert.Equal(t, minTagMsg, ve.Message)
		}
	}
}

func TestNamedParameterInterpolation(t *testing.T) {
	type User struct {
		Username string `json:"username" validate:"required,min=5"`
		Email    string `json:"email" validate:"required,email"`
		Age      int    `json:"age" validate:"required,min=18"`
	}

	// Create a user with validation errors
	user := User{
		Username: "abc",       // Too short, min=5
		Email:    "not-email", // Invalid email
		Age:      16,          // Too young, min=18
	}

	v := New()
	v.UseJsonTagName()

	// Set messages using named parameters
	v.SetDefaultTagMessage("required", "{field} is required")
	v.SetDefaultTagMessage("min", "{field} must be at least {param}")
	v.SetDefaultTagMessage("email", "{field} must be a valid email address")

	err := v.Validate(user)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// We expect 3 validation errors
	assert.Len(t, errs, 3, "Should have 3 validation errors")

	// Check that named parameters were interpolated correctly
	for _, ve := range errs {
		switch ve.Path {
		case "username":
			assert.Equal(t, "username must be at least 5", ve.Message)
		case "email":
			assert.Equal(t, "email must be a valid email address", ve.Message)
		case "age":
			assert.Equal(t, "age must be at least 18", ve.Message)
		}
	}

	// Test mixed positional and named parameters
	v = New()
	v.UseJsonTagName()
	v.SetDefaultTagMessage("min", "Field {0} with value {value} must be at least {param}")

	err = v.Validate(user)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	for _, ve := range errs {
		if ve.Path == "username" && ve.Constraint == "min" {
			assert.Equal(t, "Field username with value abc must be at least 5", ve.Message)
		}
		if ve.Path == "age" && ve.Constraint == "min" {
			assert.Equal(t, "Field age with value 16 must be at least 18", ve.Message)
		}
	}
}

func TestCustomParameterInterpolation(t *testing.T) {
	type Product struct {
		Name  string `json:"name" validate:"required"`
		Price int    `json:"price" validate:"required,min=1"`
	}

	// Create a product with validation errors
	product := Product{
		Name:  "", // Empty name (fails required)
		Price: 0,  // Zero price (fails min=1)
	}

	v := New()
	v.UseJsonTagName()

	// Add custom parameters
	v.AddCustomParam("appName", "TestShop")
	v.AddCustomParam("supportEmail", "support@example.com")

	// Set messages using custom parameters
	v.SetDefaultTagMessage("required", "{field} is required for {appName}")
	v.SetDefaultTagMessage("min", "{field} must be at least {param} for {appName}")

	// Set constraint-specific message with custom parameters
	v.SetConstraintMessage("price", "min", "Price must be positive. Contact {supportEmail} for help.")

	err := v.Validate(product)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// We expect 2 validation errors (name required, price required)
	assert.Len(t, errs, 2, "Should have 2 validation errors")

	// Check that custom parameters were interpolated correctly
	var nameErrorFound, priceErrorFound bool
	for _, ve := range errs {
		switch {
		case ve.Path == "name" && ve.Constraint == "required":
			assert.Equal(t, "name is required for TestShop", ve.Message)
			nameErrorFound = true
		case ve.Path == "price" && ve.Constraint == "required":
			assert.Equal(t, "price is required for TestShop", ve.Message)
			priceErrorFound = true
		}
	}
	assert.True(t, nameErrorFound, "Name required error not found")
	assert.True(t, priceErrorFound, "Price required error not found")

	// Go directly to testing custom parameter interpolation with a specific constraint message
	// that uses a named parameter
	v.SetConstraintMessage("price", "required", "Price field is required. Contact {supportEmail} for support.")

	err = v.Validate(product)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Check that the custom parameter was interpolated
	var customParamFound bool
	for _, ve := range errs {
		if ve.Path == "price" && ve.Constraint == "required" {
			assert.Equal(t, "Price field is required. Contact support@example.com for support.", ve.Message)
			customParamFound = true
		}
	}
	assert.True(t, customParamFound, "Custom parameter interpolation not found")

	// Test removing a custom parameter
	v.RemoveCustomParam("supportEmail")
	v.SetConstraintMessage("price", "required", "Price field needs help. Contact {supportEmail}.")

	err = v.Validate(product)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Check that the removed custom parameter is not interpolated
	for _, ve := range errs {
		if ve.Path == "price" && ve.Constraint == "required" {
			assert.Equal(t, "Price field needs help. Contact {supportEmail}.", ve.Message)
		}
	}
}

func TestRegisterTagNameFunc(t *testing.T) {
	v := New()

	// Register a custom tag name function that uses "form" tags
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := field.Tag.Get("form")
		if name == "" {
			return field.Name
		}
		return name
	})

	type Item struct {
		UserName string `form:"username" validate:"required"`
	}

	i := Item{UserName: ""}

	err := v.Validate(i)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	assert.Equal(t, "username", errs[0].Path)
	assert.Equal(t, "required", errs[0].Constraint)
}

func TestRegisterValidation(t *testing.T) {
	v := New()
	v.UseJsonTagName()

	// Register a custom validation rule that checks if a string contains "test"
	err := v.RegisterValidation("containstest", func(fl goval.FieldLevel) bool {
		return strings.Contains(fl.Field().String(), "test")
	})
	assert.NoError(t, err)

	type Item struct {
		Message string `json:"message" validate:"containstest"`
	}

	// Test with a value that doesn't contain "test"
	i1 := Item{Message: "hello"}
	err = v.Validate(i1)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")
	assert.Equal(t, "message", errs[0].Path)
	assert.Equal(t, "containstest", errs[0].Constraint)

	// Test with a value that contains "test"
	i2 := Item{Message: "hello test"}
	err = v.Validate(i2)
	assert.NoError(t, err)
}

func TestRegisterAlias(t *testing.T) {
	v := New()
	v.UseJsonTagName()

	// Register an alias for common validation rules
	v.RegisterAlias("username", "required,min=3,max=20,alphanum")

	type User struct {
		UserName string `json:"username" validate:"username"`
	}

	// Test with an invalid username
	u1 := User{UserName: "ab"}
	err := v.Validate(u1)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")
	assert.Equal(t, "username", errs[0].Path)
	assert.Equal(t, "min", errs[0].Constraint)

	// Test with a valid username
	u2 := User{UserName: "validuser123"}
	err = v.Validate(u2)
	assert.NoError(t, err)
}

func TestStructTagErrorMessages(t *testing.T) {
	type User struct {
		Username string `json:"username" validate:"required,min=5" errmsg-required:"Username is mandatory" errmsg-min:"Username must have at least {param} characters"`
		Email    string `json:"email" validate:"required,email" errmsg:"Email field has an issue"`
		Age      int    `json:"age" validate:"required,min=18" errmsg-min:"Age must be at least {param} for {appName}"`
	}

	v := New()
	v.UseJsonTagName()
	v.AddCustomParam("appName", "TestApp")

	// Create a user with validation errors
	user := User{
		Username: "",        // Empty (fails required)
		Email:    "invalid", // Invalid format (fails email)
		Age:      16,        // Too young (fails min=18)
	}

	err := v.Validate(user)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Should have 3 validation errors
	assert.Len(t, errs, 3, "Should have 3 validation errors")

	// Check that struct tag error messages were used correctly
	for _, ve := range errs {
		switch {
		case ve.Path == "username" && ve.Constraint == "required":
			// Should use constraint-specific errmsg-required tag
			assert.Equal(t, "Username is mandatory", ve.Message)
		case ve.Path == "username" && ve.Constraint == "min":
			// Should use constraint-specific errmsg-min tag with parameter interpolation
			assert.Equal(t, "Username must have at least 5 characters", ve.Message)
		case ve.Path == "email" && ve.Constraint == "email":
			// Should use the default errmsg tag for all constraints on this field
			assert.Equal(t, "Email field has an issue", ve.Message)
		case ve.Path == "age" && ve.Constraint == "min":
			// Should use constraint-specific errmsg-min tag with custom parameter interpolation
			assert.Equal(t, "Age must be at least 18 for TestApp", ve.Message)
		}
	}

	// Test fallback to other message resolution when no struct tag is present
	v.SetConstraintMessage("age", "min", "Age must be greater than or equal to {param}")

	err = v.Validate(user)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Find the age.min error and verify it uses the struct tag message (higher priority)
	var ageMinError *ValidationError
	for i, ve := range errs {
		if ve.Path == "age" && ve.Constraint == "min" {
			ageMinError = &errs[i]
			break
		}
	}

	assert.NotNil(t, ageMinError, "Age min error should exist")
	assert.Equal(t, "Age must be at least 18 for TestApp", ageMinError.Message)
}

func TestCustomParamsWithStructTagMessages(t *testing.T) {
	type Product struct {
		Name     string `json:"name" validate:"required" errmsg-required:"Product name is required for {company} catalog"`
		Price    int    `json:"price" validate:"required,min=1" errmsg-min:"Price must be at least {param} {currency}"`
		SKU      string `json:"sku" validate:"required" errmsg:"SKU error for {company} ({department})"`
		Quantity int    `json:"quantity" validate:"min=0" errmsg:"Quantity issue reported by {supportEmail}"`
	}

	// Create a product with validation errors
	product := Product{
		Name:     "", // Empty (fails required)
		Price:    0,  // Invalid (fails min=1)
		SKU:      "", // Empty (fails required)
		Quantity: -1, // Invalid (fails min=0)
	}

	// Test with multiple custom parameters
	v := New()
	v.UseJsonTagName()
	v.AddCustomParam("company", "ACME Corp")
	v.AddCustomParam("currency", "USD")
	v.AddCustomParam("department", "Electronics")
	v.AddCustomParam("supportEmail", "support@acme.com")

	err := v.Validate(product)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// We expect 4 validation errors
	assert.Len(t, errs, 4, "Should have 4 validation errors")

	// Check that custom parameters were interpolated in the struct tag messages
	for _, ve := range errs {
		switch {
		case ve.Path == "name" && ve.Constraint == "required":
			assert.Equal(t, "Product name is required for ACME Corp catalog", ve.Message)
		case ve.Path == "price" && ve.Constraint == "min":
			assert.Equal(t, "Price must be at least 1 USD", ve.Message)
		case ve.Path == "sku" && ve.Constraint == "required":
			assert.Equal(t, "SKU error for ACME Corp (Electronics)", ve.Message)
		case ve.Path == "quantity" && ve.Constraint == "min":
			assert.Equal(t, "Quantity issue reported by support@acme.com", ve.Message)
		}
	}

	// Test chained validator with additional custom parameters
	customV := v.UseMessages(NewValidationMessages())
	customV.AddCustomParam("company", "XYZ Inc")
	customV.AddCustomParam("newParam", "test123")

	err = customV.Validate(product)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Find name error and verify custom parameters from chained validator are used
	var nameError *ValidationError
	for i, ve := range errs {
		if ve.Path == "name" && ve.Constraint == "required" {
			nameError = &errs[i]
			break
		}
	}

	assert.NotNil(t, nameError, "Name error should exist")
	assert.Equal(t, "Product name is required for XYZ Inc catalog", nameError.Message,
		"Custom parameter should be overridden in the chained validator")

	// Try with new parameter that only exists in the chained validator
	type Contact struct {
		Email string `json:"email" validate:"required,email" errmsg-email:"Email format issue ({newParam})"`
	}

	contact := Contact{
		Email: "notanemail",
	}

	err = customV.Validate(contact)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	assert.Len(t, errs, 1, "Should have 1 validation error")
	assert.Equal(t, "Email format issue (test123)", errs[0].Message)
}

func TestUseMessages(t *testing.T) {
	// Create a base validator with some default messages
	baseValidator := New()
	baseValidator.SetDefaultMessage("Base default message")
	baseValidator.SetDefaultTagMessage("required", "Base required message")
	baseValidator.SetDefaultTagMessage("min", "Base minimum message")

	// Create custom messages
	customMessages := NewValidationMessages()
	customMessages.SetDefaultMessage("user.name", "Custom name message")
	customMessages.SetMessage("user.email", "email", "Custom email message")

	// Create a new validator with custom messages
	v := baseValidator.UseMessages(customMessages)

	// Test that the default message was copied
	assert.Equal(t, "Base default message", v.DefaultMessage)

	// Test that the default tag messages were copied
	assert.Equal(t, "Base required message", v.DefaultTagMessages["required"])
	assert.Equal(t, "Base minimum message", v.DefaultTagMessages["min"])

	// Test that the custom messages were set correctly
	assert.Equal(t, "Custom name message", v.Messages.ResolveMessage("user.name", "anything", nil))
	assert.Equal(t, "Custom email message", v.Messages.ResolveMessage("user.email", "email", nil))

	// Test that original validator isn't affected by changes to the new one
	v.SetDefaultMessage("New default message")
	assert.Equal(t, "Base default message", baseValidator.DefaultMessage)
	assert.Equal(t, "New default message", v.DefaultMessage)
}

func TestErrmsgTagFallback(t *testing.T) {
	// Define a struct with both specific and generic error message tags
	type User struct {
		// Only has errmsg-required but no errmsg-min
		Username string `json:"username" validate:"required,min=5" errmsg-required:"Username field is required"`

		// Has both errmsg-email and errmsg
		Email string `json:"email" validate:"required,email" errmsg-email:"Invalid email format" errmsg:"General email error"`

		// Only has generic errmsg for all constraints
		Age int `json:"age" validate:"required,min=18" errmsg:"Age validation failed"`
	}

	v := New()
	v.UseJsonTagName()

	user := User{
		Username: "abc",          // Fails min=5
		Email:    "not-an-email", // Fails email validation
		Age:      16,             // Fails min=18
	}

	err := v.Validate(user)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// We expect 3 validation errors
	assert.Len(t, errs, 3, "Should have 3 validation errors")

	// Check the error messages
	for _, ve := range errs {
		switch {
		case ve.Path == "username" && ve.Constraint == "min":
			// Should fall back to default message since no errmsg-min or errmsg is provided
			assert.NotEqual(t, "Username field is required", ve.Message,
				"Should not use errmsg-required for min constraint")
			assert.NotContains(t, ve.Message, "Username field is required",
				"Message should not contain content from errmsg-required tag")

		case ve.Path == "email" && ve.Constraint == "email":
			// Should use the specific constraint message
			assert.Equal(t, "Invalid email format", ve.Message,
				"Should use errmsg-email tag for email constraint")

		case ve.Path == "age" && ve.Constraint == "min":
			// Should fall back to the generic errmsg
			assert.Equal(t, "Age validation failed", ve.Message,
				"Should use generic errmsg tag when no constraint-specific tag exists")
		}
	}

	// Test with a field that has errmsg but not errmsg-{constraint}
	v = New()
	v.UseJsonTagName()

	type Product struct {
		Name string `json:"name" validate:"required,min=3" errmsg:"Name has validation issues"`
	}

	product := Product{
		Name: "a", // Fails min=3
	}

	err = v.Validate(product)
	errs, ok = err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	assert.Len(t, errs, 1, "Should have 1 validation error")
	assert.Equal(t, "min", errs[0].Constraint, "Constraint should be 'min'")
	assert.Equal(t, "Name has validation issues", errs[0].Message,
		"Should fall back to errmsg tag when errmsg-min is not found")
}

func TestExtractArrayIndex(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedIndex int
		expectedFound bool
	}{
		{
			name:          "Simple path without index",
			path:          "user.name",
			expectedIndex: 0,
			expectedFound: false,
		},
		{
			name:          "Path with index",
			path:          "users[2].name",
			expectedIndex: 2,
			expectedFound: true,
		},
		{
			name:          "Path with multiple indices - should get first index",
			path:          "users[1].addresses[3].street",
			expectedIndex: 1,
			expectedFound: true,
		},
		{
			name:          "Path with non-numeric index",
			path:          "users[abc].name",
			expectedIndex: 0,
			expectedFound: false,
		},
		{
			name:          "Path with empty brackets",
			path:          "users[].name",
			expectedIndex: 0,
			expectedFound: false,
		},
		{
			name:          "Path with negative index",
			path:          "users[-1].name",
			expectedIndex: 0, // Function should only match \d+ (only positive digits)
			expectedFound: false,
		},
		{
			name:          "Path with very large index",
			path:          "users[9999999].name",
			expectedIndex: 9999999,
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, found := extractArrayIndex(tt.path)
			assert.Equal(t, tt.expectedIndex, index)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}

func TestSetConstraintMessage(t *testing.T) {
	v := New()

	// Set a constraint message
	v.SetConstraintMessage("user.name", "required", "Name is required")

	// Verify that the message was set correctly
	message := v.Messages.ResolveMessage("user.name", "required", nil)
	assert.Equal(t, "Name is required", message)

	// Test path normalization with array indices
	v.SetConstraintMessage("users[0].address", "required", "Address is required")
	message = v.Messages.ResolveMessage("users[].address", "required", nil)
	assert.Equal(t, "Address is required", message)

	// Test with spaces in the path
	v.SetConstraintMessage(" user.email ", "email", "Email is invalid")
	message = v.Messages.ResolveMessage("user.email", "email", nil)
	assert.Equal(t, "Email is invalid", message)

	// Test with multiple array indices
	v.SetConstraintMessage("users[1].addresses[2].street", "required", "Street is required")
	message = v.Messages.ResolveMessage("users[].addresses[].street", "required", nil)
	assert.Equal(t, "Street is required", message)

	// Test overwriting existing message
	v.SetConstraintMessage("user.name", "required", "First name is required")
	message = v.Messages.ResolveMessage("user.name", "required", nil)
	assert.Equal(t, "First name is required", message)

	// Test with interpolation params
	v.SetConstraintMessage("user.age", "min", "Age must be at least {2}")
	params := []interface{}{"age", nil, "18"}
	message = v.Messages.ResolveMessage("user.age", "min", params)
	assert.Equal(t, "Age must be at least 18", message)
}

func TestSetPathDefaultMessage(t *testing.T) {
	v := New()

	// Set a default message for a path
	v.SetPathDefaultMessage("user.name", "Name is invalid")

	// Verify that the message is used for any constraint
	message := v.Messages.ResolveMessage("user.name", "anything", nil)
	assert.Equal(t, "Name is invalid", message)

	// Test that a specific constraint message takes precedence
	v.SetConstraintMessage("user.name", "required", "Name is required")
	message = v.Messages.ResolveMessage("user.name", "required", nil)
	assert.Equal(t, "Name is required", message)

	// But the default is still used for other constraints
	message = v.Messages.ResolveMessage("user.name", "min", nil)
	assert.Equal(t, "Name is invalid", message)

	// Test path normalization
	v.SetPathDefaultMessage("users[0].address", "Address is invalid")
	message = v.Messages.ResolveMessage("users[].address", "min", nil)
	assert.Equal(t, "Address is invalid", message)

	// Test with parameter interpolation
	v.SetPathDefaultMessage("user.age", "Age {0} is invalid")
	params := []interface{}{"age", nil, nil}
	message = v.Messages.ResolveMessage("user.age", "anything", params)
	assert.Equal(t, "Age age is invalid", message)

	// Test overwriting existing message
	v.SetPathDefaultMessage("user.name", "User name is invalid")
	message = v.Messages.ResolveMessage("user.name", "min", nil)
	assert.Equal(t, "User name is invalid", message)
}

func TestNestedStructErrmsgTags(t *testing.T) {
	// Define a nested struct with errmsg tags at different levels
	type Address struct {
		Street  string `json:"street" validate:"required" errmsg-required:"Street is a required field"`
		City    string `json:"city" validate:"required" errmsg-required:"City cannot be empty"`
		ZipCode string `json:"zipCode" validate:"required,len=5" errmsg-len:"Zip code must be exactly {param} characters"`
	}

	type User struct {
		Name     string         `json:"name" validate:"required" errmsg-required:"User name is required"`
		Email    string         `json:"email" validate:"required,email" errmsg-email:"Please provide a valid email address"`
		Address  []Address      `json:"addresses" validate:"required,dive"`
		Metadata map[string]int `json:"metadata" validate:"required,dive,gt=0" errmsg-required:"Metadata is required" errmsg-gt:"{field} must be greater than {param}"`
	}

	// Create a user with validation errors at different nesting levels
	user := User{
		Name:  "",         // Missing name
		Email: "invalid@", // Invalid email format
		Address: []Address{
			{
				Street:  "",    // Missing street
				City:    "",    // Missing city
				ZipCode: "123", // Zip code too short
			},
		},
		Metadata: map[string]int{
			"exp": 0,
		},
	}

	v := New()
	v.UseJsonTagName()

	err := v.Validate(user)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// Create a map to easily check the error messages for each path
	errorMap := make(map[string]string)
	for _, ve := range errs {
		errorMap[ve.Path+":"+ve.Constraint] = ve.Message
	}

	t.Log(errorMap)

	// Check that the error messages from the struct tags are correctly applied
	assert.Equal(t, "User name is required", errorMap["name:required"])
	assert.Equal(t, "Please provide a valid email address", errorMap["email:email"])
	assert.Equal(t, "Street is a required field", errorMap["addresses[0].street:required"])
	assert.Equal(t, "City cannot be empty", errorMap["addresses[0].city:required"])
	assert.Equal(t, "Zip code must be exactly 5 characters", errorMap["addresses[0].zipCode:len"])
	assert.Equal(t, "metadata[exp] must be greater than 0", errorMap["metadata[exp]:gt"])
}

func TestNestedArrayCustomParams(t *testing.T) {
	// Define a nested struct with errmsg tags that use custom parameters
	type Contact struct {
		Type  string `json:"type" validate:"required,oneof=home work mobile" errmsg-oneof:"Contact type must be one of: {param}"`
		Value string `json:"value" validate:"required" errmsg-required:"Contact {field} is required for {appName}"`
	}

	type Profile struct {
		Bio      string    `json:"bio" validate:"required,min=10" errmsg-min:"Bio must be at least {param} characters long ({appName} requirement)"`
		Contacts []Contact `json:"contacts" validate:"required,min=1,dive" errmsg-min:"At least {param} contact is required for {appName}"`
	}

	type User struct {
		Username string  `json:"username" validate:"required" errmsg-required:"Username is required for {appName}"`
		Profile  Profile `json:"profile" validate:"required"`
	}

	// Create a user with validation errors at different nesting levels
	user := User{
		Username: "",
		Profile: Profile{
			Bio: "Too short", // Bio too short
			Contacts: []Contact{
				{
					Type:  "invalid", // Invalid type
					Value: "",        // Missing value
				},
			},
		},
	}

	v := New()
	v.UseJsonTagName()
	v.AddCustomParam("appName", "TestingApp")

	err := v.Validate(user)
	errs, ok := err.(ValidationErrors)
	assert.True(t, ok, "Should be of type ValidationErrors")

	// We expect 4 validation errors
	assert.Len(t, errs, 4, "Should have 4 validation errors")

	// Create a map to easily check the error messages for each path
	errorMap := make(map[string]string)
	for _, ve := range errs {
		errorMap[ve.Path+":"+ve.Constraint] = ve.Message
	}

	// Check that the error messages from the struct tags are correctly applied
	// and custom parameters are interpolated properly
	assert.Equal(t, "Username is required for TestingApp", errorMap["username:required"])
	assert.Equal(t, "Bio must be at least 10 characters long (TestingApp requirement)", errorMap["profile.bio:min"])
	assert.Equal(t, "Contact type must be one of: home work mobile", errorMap["profile.contacts[0].type:oneof"])
	assert.Equal(t, "Contact value is required for TestingApp", errorMap["profile.contacts[0].value:required"])
}
