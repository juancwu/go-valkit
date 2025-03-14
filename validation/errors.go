package validation

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Tag     string
	Value   interface{}
	Message string
	Param   string
}

// ValidationErrors represents a map of validation errors
type ValidationErrors map[string]*ValidationError
