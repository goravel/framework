package client

type Factory interface {
	Request

	// Client retrieves a configured Client instance by name.
	//
	// If no name is provided, the "default" client is returned.
	// If the requested client name is not defined in the configuration,
	// this method may panic or return a default, depending on implementation policy.
	Client(name ...string) Client

	// Request is a convenience alias for Client(name...).NewRequest().
	// It immediately starts building a request using the specified (or default) client.
	Request(name ...string) Request
}
