package client

// ResponseFactory defines the contract for creating stubbed response instances.
type ResponseFactory interface {
	// Json creates a JSON response with the application/json content type.
	Json(data any, status int) Response
	// String creates a response with a plain text body.
	String(body string, status int) Response
	// Status creates a response with only a status code and an empty body.
	Status(code int) Response
	// Success creates an empty 200 OK response.
	Success() Response
	// File creates a response containing the contents of the specified file.
	File(path string, status int) Response
	// Make creates a custom response with the specified body, status, and headers.
	Make(body string, status int, headers map[string]string) Response
}
