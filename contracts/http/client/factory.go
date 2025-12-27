package client

type Factory interface {
	// Client returns a named HTTP client.
	// If no name is provided, the default client is returned.
	Client(name ...string) Client
}
