package client

type Factory interface {
	// Request embeds the HTTP verb methods (Get, Post, etc.) and configuration methods (WithHeader, etc.).
	//
	// This embedding allows you to use the default client directly without calling Client().
	//
	// Example:
	//   // Uses the 'default' client defined in config/http.go
	//   facades.Http().Get("/users")
	Request

	// Client switches the context to a specific client configuration.
	//
	// It returns a Request builder pre-configured with the specific client's settings
	// (such as BaseURL, Timeout, and Headers) defined in your configuration file.
	//
	// If no name is provided, the default client is returned.
	//
	// Example:
	//   // Switch to the 'github' client and make a request
	//   facades.Http().Client("github").Post("/charges", data)
	Client(name ...string) Request
}
