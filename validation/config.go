package validation

import "github.com/go-playground/validator/v10"

// Config represents the configuration for the validator
type Config struct {
	TagName         string            // The tag name to extract field names (e.g., "json", "form")
	DefaultMessages map[string]string // Default messages for validation tags
	CustomMessages  map[string]string // Custom messages for specific field paths
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		TagName: "json",
		DefaultMessages: map[string]string{
			"required": "This field is required",
			"email":    "Invalid email format",
			"min":      "Value must be at least {0}",
			"max":      "Value must be at most {0}",
			"len":      "Value must be exactly {0} characters long",
		},
		CustomMessages: make(map[string]string),
	}
}

// NewWithConfig creates a new Validator instance with the given configuration
func NewWithConfig(config Config) *Validator {
	v := Validator{
		validator:       validator.New(),
		tagName:         config.TagName,
		defaultMessages: make(map[string]string),
		customMessages:  make(map[string]string),
	}

	// Set default messages
	for tag, message := range config.DefaultMessages {
		v.SetDefaultMessage(tag, message)
	}

	// Set custom messages
	for path, message := range config.CustomMessages {
		v.SetCustomMessage(path, message)
	}

	return &v
}
