package client

type ResponseSequence interface {
	// Push appends a response to the sequence.
	//
	// You can pass an optional integer as the second argument to specify how many times
	// this response should be returned before moving to the next one.
	//
	// Example:
	//   // Return 500 three times, then 200 once
	//   fail := facades.Http().Response().Status(500)
	//   ok := facades.Http().Response().Success()
	//
	//   facades.Http().Sequence().
	//       Push(fail, 3).
	//       Push(ok)
	Push(response Response, count ...int) ResponseSequence

	// PushStatus is a convenience method to push a simple status code response.
	//
	// Example:
	//   facades.Http().Sequence().PushStatus(404)
	PushStatus(status int, count ...int) ResponseSequence

	// PushString is a convenience method to push a simple string body response.
	//
	// Example:
	//   facades.Http().Sequence().PushString("Hello", 200)
	PushString(body string, status int, count ...int) ResponseSequence

	// WhenEmpty sets the default response to return when the sequence is exhausted.
	//
	// If not set, the sequence will usually return a 404 when it runs out of responses.
	//
	// Example:
	//   facades.Http().Sequence().
	//       PushStatus(200).
	//       WhenEmpty(facades.Http().Response().Status(418))
	WhenEmpty(response Response) ResponseSequence
}
