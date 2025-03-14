package validation

import (
	"testing"
)

type TestStruct struct {
	Name      string        `json:"name" validate:"required"`
	Email     string        `json:"email" validate:"required,email"`
	Age       int           `json:"age" validate:"required,gte=18"`
	Addresses []TestAddress `json:"addresses" validate:"dive"`
}

type TestAddress struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	ZipCode string `json:"zip_code" validate:"required"`
}

func TestValidator(t *testing.T) {
	// Create a validator
	v := New("json")

	// Set default messages
	v.SetDefaultMessage("required", "This field is required")
	v.SetDefaultMessage("email", "Invalid email format")
	v.SetDefaultMessage("gte", "Value must be greater than or equal to {0}")

	// Set custom messages
	v.SetCustomMessage("Name", "Please enter your full name")
	v.SetCustomMessage("Addresses[].Street", "Please enter a valid street address")

	// Create a test struct with validation errors
	test := TestStruct{
		Name:  "",
		Email: "not-an-email",
		Age:   16,
		Addresses: []TestAddress{
			{
				Street:  "",
				City:    "New York",
				ZipCode: "10001",
			},
		},
	}

	// Validate the struct
	errs, err := v.Validate(test)

	// Check if there are validation errors
	if err == nil {
		t.Error("Expected validation errors, got nil")
	}

	// Check if the validation errors are correct
	if len(errs) != 4 {
		t.Errorf("Expected 4 validation errors, got %d", len(errs))
	}

	// Check if the custom messages are applied
	if errs["Name"].Message != "Please enter your full name" {
		t.Errorf("Expected custom message for Name, got %s", errs["Name"].Message)
	}

	if errs["Addresses[].Street"].Message != "Please enter a valid street address" {
		t.Errorf("Expected custom message for Addresses[].Street, got %s", errs["Addresses[].Street"].Message)
	}
}
