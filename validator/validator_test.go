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
