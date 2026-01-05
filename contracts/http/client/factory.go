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

	// Fake allows you to instruct the HTTP client to return stubbed / dummy responses
	// when requests are made.
	//
	// This is useful for testing to prevent actual network calls to third-party APIs.
	//
	// NOTE: When using the global Facade (facades.Http()), calling Fake() modifies global state.
	// Therefore, tests using Fake() cannot be run in parallel (t.Parallel()).
	//
	// The matches map keys can be:
	//  1. A URL pattern (e.g., "github.com/*" or "api.stripe.com/v1/charges").
	//  2. A Client Name (e.g., "github" - matches the BaseURL of the 'github' client).
	//  3. A Client Name + Path (e.g., "github:/repos/*").
	//
	// The values can be:
	//  - string: Returns a 200 OK with the string body.
	//  - int: Returns a response with this status code and empty body.
	//  - client.Response: A response object created via facades.Http().Response().
	//  - func(client.Request) client.Response: A dynamic callback.
	//
	// Example:
	//   facades.Http().Fake(map[string]any{
	//       "github": facades.Http().Response().Json(data, 200),
	//       "google.com/*": "Hello World",
	//   })
	Fake(mocks map[string]any)

	// Reset clears all fakes, recorded requests, and cached clients.
	//
	// It restores the factory to its original state, making real network calls.
	// You should call this in the defer statement of your tests.
	//
	// Example:
	//   defer facades.Http().Reset()
	Reset()

	// Response returns a factory for creating fake response instances.
	//
	// Use this to generate rich responses (JSON, Headers, Status) for use in
	// Fake() or Sequence().
	//
	// Example:
	//   resp := facades.Http().Response().Json(map[string]string{"foo": "bar"}, 200)
	Response() ResponseFactory

	// Sequence creates a fluent builder for defining a sequence of responses.
	//
	// This is useful when you want to mock a specific URL to return different
	// responses in order (e.g., Fail 500 -> Fail 500 -> Success 200).
	//
	// Example:
	//   facades.Http().Sequence().
	//       PushStatus(500).
	//       Push("OK", 200)
	Sequence() ResponseSequence

	// AssertSent asserts that a request matching the given truth test was sent.
	//
	// It returns true if at least one request matched the criteria.
	//
	// Example:
	//   facades.Http().AssertSent(func(req client.Request) bool {
	//       return req.Url() == "https://api.github.com/user" && req.Method() == "GET"
	//   })
	AssertSent(assertion func(req Request) bool) bool

	// AssertSentCount asserts that the given number of requests were sent in total.
	//
	// Example:
	//   facades.Http().AssertSentCount(3)
	AssertSentCount(count int) bool

	// AssertNotSent asserts that no request matching the given truth test was sent.
	//
	// Example:
	//   facades.Http().AssertNotSent(func(req client.Request) bool {
	//       return req.Client() == "stripe"
	//   })
	AssertNotSent(assertion func(req Request) bool) bool

	// AssertNothingSent asserts that no requests were sent at all.
	AssertNothingSent() bool
}
