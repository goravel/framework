package client

// ResponseSequence defines the contract for building an ordered sequence of responses.
type ResponseSequence interface {
	// Push adds a response to the sequence, optionally repeating it multiple times.
	Push(response Response, count ...int) ResponseSequence
	// PushStatus adds a status-only response to the sequence.
	PushStatus(status int, count ...int) ResponseSequence
	// PushString adds a string body response to the sequence.
	PushString(body string, status int, count ...int) ResponseSequence
	// WhenEmpty defines the response to return once the sequence is exhausted.
	WhenEmpty(response Response) ResponseSequence
}
