package echo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/juancwu/go-valkit/integrations"
	"github.com/juancwu/go-valkit/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	Name    string   `json:"name" validate:"required,min=3"`
	Email   string   `json:"email" validate:"required,email"`
	Age     int      `json:"age" validate:"min=18"`
	Hobbies []string `json:"hobbies" validate:"required,min=1,dive,required"`
}

// setupEcho creates an Echo instance with test routes
func setupEcho(configFn func(*validator.Validator)) (*echo.Echo, *Validator) {
	e := echo.New()
	var v *Validator

	if configFn != nil {
		ConfigureWithOptions(e, configFn)
		v = e.Validator.(*Validator)
	} else {
		Configure(e)
		v = e.Validator.(*Validator)
	}

	// Define test endpoint
	e.POST("/users", func(c echo.Context) error {
		u := new(TestUser)
		if err := c.Bind(u); err != nil {
			return err
		}

		if err := c.Validate(u); err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"status":  "success",
			"message": "User created",
			"data":    u,
		})
	})

	return e, v
}

// TestBasicValidation tests basic validation success and failure
func TestBasicValidation(t *testing.T) {
	e, _ := setupEcho(nil)

	t.Run("Valid User", func(t *testing.T) {
		// Valid user data
		userData := TestUser{
			Name:    "John Doe",
			Email:   "john@example.com",
			Age:     25,
			Hobbies: []string{"Reading", "Cycling"},
		}

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "success", response["status"])
	})

	t.Run("Invalid User", func(t *testing.T) {
		// Invalid user data (missing required fields, invalid email, under age)
		userData := TestUser{
			Name:    "Jo", // too short
			Email:   "not-an-email",
			Age:     16,         // under minimum
			Hobbies: []string{}, // empty array
		}

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response integrations.DefaultErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Validation failed", response.Message)

		// Should have errors for all fields
		assert.Contains(t, response.Errors, "name")
		assert.Contains(t, response.Errors, "email")
		assert.Contains(t, response.Errors, "age")
		assert.Contains(t, response.Errors, "hobbies")
	})
}

// TestCustomValidationMessages tests custom validation messages
func TestCustomValidationMessages(t *testing.T) {
	e, _ := setupEcho(func(v *validator.Validator) {
		v.SetDefaultMessage("Input is invalid")
		v.SetDefaultTagMessage("required", "Field {0} is required")
		v.SetDefaultTagMessage("email", "Please enter a valid email address")
		v.SetDefaultTagMessage("min", "Minimum value is {2}")

		v.SetConstraintMessage("age", "min", "You must be at least {2} years old")
	})

	// Invalid user data
	userData := TestUser{
		Name:    "",
		Email:   "not-an-email",
		Age:     16,
		Hobbies: nil,
	}

	body, _ := json.Marshal(userData)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response integrations.DefaultErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Check custom messages
	assert.Contains(t, response.Errors["name"][0], "Field name is required")
	assert.Contains(t, response.Errors["email"][0], "valid email address")
	assert.Contains(t, response.Errors["age"][0], "at least 18 years old")
	assert.Contains(t, response.Errors["hobbies"][0], "Field hobbies is required")
}

// Custom error response structure
type CustomErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

// Custom formatter
type CustomFormatter struct{}

func (f *CustomFormatter) Format(ve validator.ValidationErrors) interface{} {
	fields := make(map[string]string)

	for _, err := range ve {
		fields[err.Field] = err.Message
	}

	return CustomErrorResponse{
		Code:    400,
		Message: "Form has errors",
		Fields:  fields,
	}
}

// TestCustomErrorFormatter tests using a custom error formatter
func TestCustomErrorFormatter(t *testing.T) {

	e := echo.New()
	v := NewValidator().WithFormatter(&CustomFormatter{})
	e.Validator = v
	e.HTTPErrorHandler = ValidationErrorHandler(v, e.DefaultHTTPErrorHandler)

	e.POST("/users", func(c echo.Context) error {
		u := new(TestUser)
		if err := c.Bind(u); err != nil {
			return err
		}

		if err := c.Validate(u); err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, map[string]string{"status": "success"})
	})

	// Invalid user data
	userData := TestUser{
		Name: "",
	}

	body, _ := json.Marshal(userData)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response CustomErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	assert.Equal(t, 400, response.Code)
	assert.Equal(t, "Form has errors", response.Message)
	assert.NotEmpty(t, response.Fields)
}

// TestGetValidator tests the GetValidator method
func TestGetValidator(t *testing.T) {
	_, v := setupEcho(nil)

	validator := v.GetValidator()
	assert.NotNil(t, validator)

	// Check that we can modify the validator
	validator.SetDefaultMessage("Custom message")
	assert.Equal(t, "Custom message", validator.DefaultMessage)
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	e, _ := setupEcho(nil)

	t.Run("Empty JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid:json}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Wrong Content Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("plain text"))
		req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
	})
}

// Complex nested test structs
type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	ZipCode string `json:"zipCode" validate:"required,numeric,len=5"`
	Country string `json:"country" validate:"required"`
}

type Contact struct {
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,numeric"`
	Preferred string `json:"preferred" validate:"required,oneof=email phone"`
}

type ComplexUser struct {
	ID         int       `json:"id"`
	Name       string    `json:"name" validate:"required,min=3"`
	Address    Address   `json:"address" validate:"required"`
	Contacts   []Contact `json:"contacts" validate:"required,min=1,dive"`
	Categories []string  `json:"categories" validate:"required,min=1,dive,required,min=2"`
}

func setupNestedValidationTest() *echo.Echo {
	e := echo.New()
	Configure(e)

	e.POST("/complex-users", func(c echo.Context) error {
		user := new(ComplexUser)
		if err := c.Bind(user); err != nil {
			return err
		}

		if err := c.Validate(user); err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"status": "success",
			"data":   user,
		})
	})

	return e
}

func TestNestedValidation(t *testing.T) {
	e := setupNestedValidationTest()

	t.Run("Valid Complex User", func(t *testing.T) {
		userData := ComplexUser{
			Name: "John Doe",
			Address: Address{
				Street:  "123 Main St",
				City:    "New York",
				ZipCode: "10001",
				Country: "USA",
			},
			Contacts: []Contact{
				{
					Email:     "john@example.com",
					Phone:     "5551234567",
					Preferred: "email",
				},
			},
			Categories: []string{"technology", "programming"},
		}

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest(http.MethodPost, "/complex-users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Invalid Complex User", func(t *testing.T) {
		// Create user with various validation errors
		userData := ComplexUser{
			Name: "Jo", // too short
			Address: Address{
				// Missing street
				City:    "New York",
				ZipCode: "ABC12", // not numeric
				// Missing country
			},
			Contacts: []Contact{
				{
					Email:     "not-an-email",
					Phone:     "abc123", // not numeric
					Preferred: "mail",   // not in oneof
				},
			},
			Categories: []string{"a", ""}, // second item empty
		}

		body, _ := json.Marshal(userData)
		req := httptest.NewRequest(http.MethodPost, "/complex-users", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response integrations.DefaultErrorResponse
		json.Unmarshal(rec.Body.Bytes(), &response)

		// Check that we have error messages for the nested fields
		assert.Contains(t, response.Errors, "name")
		assert.Contains(t, response.Errors, "address.street")
		assert.Contains(t, response.Errors, "address.zipCode")
		assert.Contains(t, response.Errors, "address.country")
		assert.Contains(t, response.Errors, "contacts[0].email")
		assert.Contains(t, response.Errors, "contacts[0].phone")
		assert.Contains(t, response.Errors, "contacts[0].preferred")
		assert.Contains(t, response.Errors, "categories[1]")
	})
}

func TestCustomNestedValidationMessages(t *testing.T) {
	// Set up Echo with custom validation messages for nested fields
	e := echo.New()

	ConfigureWithOptions(e, func(v *validator.Validator) {
		// Set default message formats
		v.SetDefaultTagMessage("required", "{0} is required")
		v.SetDefaultTagMessage("email", "{0} must be a valid email")
		v.SetDefaultTagMessage("min", "{0} should be minimum {2}")

		// Set specific field messages
		v.SetConstraintMessage("address.street", "required", "Street address cannot be empty")
		v.SetConstraintMessage("address.zipCode", "numeric", "ZIP code must contain only numbers")
		v.SetConstraintMessage("contacts[].email", "email", "Contact email must be valid")
		v.SetConstraintMessage("categories[]", "min", "Each category must be at least {2} characters")
	})

	e.POST("/complex-users", func(c echo.Context) error {
		user := new(ComplexUser)
		if err := c.Bind(user); err != nil {
			return err
		}

		if err := c.Validate(user); err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, map[string]string{"status": "success"})
	})

	// Test with invalid data
	userData := ComplexUser{
		Name: "Jo", // too short
		Address: Address{
			// Missing street
			City:    "New York",
			ZipCode: "ABC12", // not numeric
			Country: "USA",
		},
		Contacts: []Contact{
			{
				Email:     "not-an-email",
				Phone:     "5551234567",
				Preferred: "email",
			},
		},
		Categories: []string{"a", "programming"},
	}

	body, _ := json.Marshal(userData)
	req := httptest.NewRequest(http.MethodPost, "/complex-users", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response integrations.DefaultErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Check custom error messages
	assert.Contains(t, response.Errors["name"][0], "should be minimum 3")
	assert.Contains(t, response.Errors["address.street"][0], "Street address cannot be empty")
	assert.Contains(t, response.Errors["address.zipCode"][0], "ZIP code must contain only numbers")
	assert.Contains(t, response.Errors["contacts[0].email"][0], "Contact email must be valid")
	assert.Contains(t, response.Errors["categories[0]"][0], "Each category must be at least 2 characters")
}

// TestArrayValidationWithMultipleErrors checks if multiple errors in the same array element
// are properly reported and formatted
func TestArrayValidationWithMultipleErrors(t *testing.T) {
	e := echo.New()
	Configure(e)

	type ArrayItem struct {
		ID   int    `json:"id" validate:"required,min=1"`
		Name string `json:"name" validate:"required,min=3"`
		Code string `json:"code" validate:"required,len=6"`
	}

	type ArrayTest struct {
		Items []ArrayItem `json:"items" validate:"required,min=1,dive"`
	}

	e.POST("/array-test", func(c echo.Context) error {
		test := new(ArrayTest)
		if err := c.Bind(test); err != nil {
			return err
		}

		if err := c.Validate(test); err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, map[string]string{"status": "success"})
	})

	// Create test with multiple errors in each array item
	testData := ArrayTest{
		Items: []ArrayItem{
			{
				ID:   0,    // min=1 error
				Name: "A",  // min=3 error
				Code: "12", // len=6 error
			},
			{
				ID:   2,
				Name: "", // required error
				Code: "123456",
			},
		},
	}

	body, _ := json.Marshal(testData)
	req := httptest.NewRequest(http.MethodPost, "/array-test", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response integrations.DefaultErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Check that we have all the expected error fields
	assert.Contains(t, response.Errors, "items[0].id")
	assert.Contains(t, response.Errors, "items[0].name")
	assert.Contains(t, response.Errors, "items[0].code")
	assert.Contains(t, response.Errors, "items[1].name")

	// Make sure we don't have errors for valid fields
	assert.NotContains(t, response.Errors, "items[1].id")
	assert.NotContains(t, response.Errors, "items[1].code")
}
