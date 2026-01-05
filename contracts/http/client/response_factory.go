package client

type ResponseFactory interface {
	// Json creates a JSON response.
	//
	// It automatically sets the "Content-Type" header to "application/json"
	// and marshals the given data.
	//
	// Example:
	//   facades.Http().Response().Json(map[string]any{"id": 1}, 200)
	Json(data any, status int) Response

	// String creates a simple string response.
	//
	// Example:
	//   facades.Http().Response().String("Hello World", 200)
	String(body string, status int) Response

	// Status creates a response with the given status code and no-body.
	//
	// Example:
	//   facades.Http().Response().Status(404)
	Status(code int) Response

	// Success creates an empty 200 OK response.
	//
	// Example:
	//   facades.Http().Response().Success()
	Success() Response

	// File creates a response that serves a file's contents.
	//
	// Useful for testing file download endpoints.
	//
	// Example:
	//   facades.Http().Response().File("/path/to/image.png", 200)
	File(path string, status int) Response

	// Make creates a fully custom response.
	//
	// Use this when you need specific headers or cookies that simpler methods don't support.
	//
	// Example:
	//   headers := map[string]string{"X-Custom": "Value"}
	//   facades.Http().Response().Make("body", 200, headers)
	Make(body string, status int, headers map[string]string) Response
}
