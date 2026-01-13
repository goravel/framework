package client

type FakeResponse interface {
	// File creates a mock response using the contents of a file at the specified path.
	File(path string, status int) Response

	// Json creates a mock response with a JSON body and "application/json" content type.
	Json(code int, obj any) Response

	// Make constructs a custom mock response with the specified body, status, and headers.
	Make(body string, status int, headers map[string]string) Response

	// OK creates a generic 200 OK mock response with an empty body.
	OK() Response

	// Status creates a mock response with the specified status code and an empty body.
	Status(code int) Response

	// String creates a mock response with a raw string body.
	String(body string, status int) Response
}
