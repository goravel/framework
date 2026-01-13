package client

type FakeSequence interface {
	// Push adds a specific response to the sequence.
	Push(response Response, count ...int) FakeSequence

	// PushStatus adds a status-only response to the sequence.
	PushStatus(status int, count ...int) FakeSequence

	// PushString adds a string-body response to the sequence.
	PushString(body string, status int, count ...int) FakeSequence

	// WhenEmpty defines the default response to return when the sequence is exhausted.
	WhenEmpty(response Response) FakeSequence
}
