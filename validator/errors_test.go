package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationErrorError(t *testing.T) {
	ve := ValidationError{
		Field:      "name",
		Path:       "user.name",
		Message:    "Name is required",
		Constraint: "required",
		Param:      "",
		Actual:     "",
	}

	expected := "ValidationError (user.name | required): Name is required"
	assert.Equal(t, expected, ve.Error())
}

func TestValidationErrorsError(t *testing.T) {
	ves := ValidationErrors{
		{
			Field:      "name",
			Path:       "user.name",
			Message:    "Name is required",
			Constraint: "required",
			Param:      "",
			Actual:     "",
		},
		{
			Field:      "email",
			Path:       "user.email",
			Message:    "Email is invalid",
			Constraint: "email",
			Param:      "",
			Actual:     "invalid",
		},
	}

	expected := "ValidationError (user.name | required): Name is required;ValidationError (user.email | email): Email is invalid;"
	assert.Equal(t, expected, ves.Error())
}

func TestGroupErrorsByPath(t *testing.T) {
	ves := ValidationErrors{
		{
			Field:      "name",
			Path:       "user.name",
			Message:    "Name is required",
			Constraint: "required",
			Param:      "",
			Actual:     "",
		},
		{
			Field:      "name",
			Path:       "user.name",
			Message:    "Name must be at least 3 characters",
			Constraint: "min",
			Param:      "3",
			Actual:     "ab",
		},
		{
			Field:      "email",
			Path:       "user.email",
			Message:    "Email is invalid",
			Constraint: "email",
			Param:      "",
			Actual:     "invalid",
		},
		{
			Field:      "password",
			Path:       "user.password",
			Message:    "Password is required",
			Constraint: "required",
			Param:      "",
			Actual:     "",
		},
	}

	grouped := ves.GroupErrorsByPath()

	// Should have 3 groups: user.name, user.email, and user.password
	assert.Len(t, grouped, 3)

	// user.name should have 2 errors
	assert.Len(t, grouped["user.name"], 2)

	// user.email should have 1 error
	assert.Len(t, grouped["user.email"], 1)

	// user.password should have 1 error
	assert.Len(t, grouped["user.password"], 1)

	// Verify constraints for user.name errors
	constraints := []string{}
	for _, err := range grouped["user.name"] {
		constraints = append(constraints, err.Constraint)
	}
	assert.ElementsMatch(t, []string{"required", "min"}, constraints)

	// Verify error properties are preserved
	for _, err := range grouped["user.name"] {
		if err.Constraint == "min" {
			assert.Equal(t, "3", err.Param)
			assert.Equal(t, "ab", err.Actual)
		}
	}
}

func TestErrorsForPath(t *testing.T) {
	ves := ValidationErrors{
		{
			Field:      "name",
			Path:       "user.name",
			Message:    "Name is required",
			Constraint: "required",
			Param:      "",
			Actual:     "",
		},
		{
			Field:      "name",
			Path:       "user.name",
			Message:    "Name must be at least 3 characters",
			Constraint: "min",
			Param:      "3",
			Actual:     "ab",
		},
		{
			Field:      "email",
			Path:       "user.email",
			Message:    "Email is invalid",
			Constraint: "email",
			Param:      "",
			Actual:     "invalid",
		},
	}

	nameErrors := ves.ErrorsForPath("user.name")

	// Should have 2 errors for user.name
	assert.Len(t, nameErrors, 2)

	// Verify constraints for user.name errors
	constraints := []string{}
	for _, err := range nameErrors {
		constraints = append(constraints, err.Constraint)
	}
	assert.ElementsMatch(t, []string{"required", "min"}, constraints)

	// Error properties should be preserved
	for _, err := range nameErrors {
		assert.Equal(t, "user.name", err.Path)
		assert.Equal(t, "name", err.Field)
		if err.Constraint == "min" {
			assert.Equal(t, "3", err.Param)
			assert.Equal(t, "ab", err.Actual)
		}
	}

	// Should have 1 error for user.email
	emailErrors := ves.ErrorsForPath("user.email")
	assert.Len(t, emailErrors, 1)
	assert.Equal(t, "email", emailErrors[0].Constraint)

	// Should have none for a non-existent path
	nonExistentErrors := ves.ErrorsForPath("non.existent")
	assert.Len(t, nonExistentErrors, 0)
}
