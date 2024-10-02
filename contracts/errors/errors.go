package errors

// Error is the interface that wraps the basic error methods
type Error interface {
	// Args allows setting arguments for the placeholders in the text
	Args(...any) Error
	// Error implements the error interface and formats the error string
	Error() string
	// Location explicitly sets the location in the error message
	Location(string) Error
	// WithLocation enables or disables the inclusion of the location in the error message
	WithLocation(bool) Error
}
