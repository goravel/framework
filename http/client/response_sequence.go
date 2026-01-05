package client

import (
	"sync"

	"github.com/goravel/framework/contracts/http/client"
)

var _ client.ResponseSequence = (*ResponseSequence)(nil)

type ResponseSequence struct {
	mu        sync.Mutex
	responses []client.Response
	factory   client.ResponseFactory
	whenEmpty client.Response
	current   int
}

func NewResponseSequence(factory client.ResponseFactory) *ResponseSequence {
	return &ResponseSequence{
		factory:   factory,
		responses: make([]client.Response, 0),
	}
}

func (r *ResponseSequence) Push(response client.Response, count ...int) client.ResponseSequence {
	r.mu.Lock()
	defer r.mu.Unlock()

	times := 1
	if len(count) > 0 && count[0] > 0 {
		times = count[0]
	}

	for i := 0; i < times; i++ {
		r.responses = append(r.responses, response)
	}

	return r
}

func (r *ResponseSequence) PushStatus(status int, count ...int) client.ResponseSequence {
	return r.Push(r.factory.Status(status), count...)
}

func (r *ResponseSequence) PushString(body string, status int, count ...int) client.ResponseSequence {
	return r.Push(r.factory.String(body, status), count...)
}

func (r *ResponseSequence) WhenEmpty(response client.Response) client.ResponseSequence {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.whenEmpty = response

	return r
}

// getNext retrieves the next response in the sequence.
//
// This method is intended for internal use by the MockTransport to
// resolve the response for the current request.
func (r *ResponseSequence) getNext() client.Response {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.current < len(r.responses) {
		response := r.responses[r.current]
		r.current++
		return response
	}

	if r.whenEmpty != nil {
		return r.whenEmpty
	}

	return r.factory.Status(404)
}
