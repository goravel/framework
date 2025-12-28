package client

type Factory interface {
	// Client returns a configured HTTP client by name.
	//
	// If no name is provided, the default client is returned.
	// If the requested client is not configured, this method panics.
	//
	// This method is intended for advanced usage where the caller
	// wants to reuse a client instance or create multiple requests
	// from the same client.
	Client(name ...string) Client

	// Request returns a new HTTP request builder for the given client.
	//
	// This is a convenience method equivalent to:
	//   Http().Client(name).NewRequest()
	//
	// If no name is provided, the default client is used.
	// If the requested client is not configured, this method panics.
	Request(name ...string) Request
}
