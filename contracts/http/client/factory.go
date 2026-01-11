package client

// Factory defines the contract for managing HTTP clients and testing state.
type Factory interface {
	// Request provides access to the default client's request builder methods.
	Request
	// Client returns a request builder for a specific named configuration.
	Client(name ...string) Request
	// Fake intercepts outbound requests and returns stubbed responses.
	Fake(mocks map[string]any) Factory
	// PreventStrayRequests ensures all sent requests match a defined mock.
	PreventStrayRequests() Factory
	// AllowStrayRequests permits specific URL patterns to bypass the mock firewall.
	AllowStrayRequests(patterns []string) Factory
	// Reset restores the factory to its original state and clears all mocks.
	Reset()
	// Response returns a factory for creating stubbed response instances.
	Response() ResponseFactory
	// Sequence creates a builder for defining an ordered sequence of responses.
	Sequence() ResponseSequence
	// AssertSent verifies that at least one request matching the criteria was executed.
	AssertSent(assertion func(req Request) bool) bool
	// AssertSentCount verifies the total number of requests sent matches the expected count.
	AssertSentCount(count int) bool
	// AssertNotSent verifies that no requests matching the criteria were executed.
	AssertNotSent(assertion func(req Request) bool) bool
	// AssertNothingSent verifies that no HTTP requests were executed.
	AssertNothingSent() bool
}
