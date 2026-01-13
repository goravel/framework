package client

import (
	"sync"

	"github.com/goravel/framework/contracts/http/client"
)

var _ client.FakeSequence = (*FakeSequence)(nil)

type FakeSequence struct {
	mu        sync.Mutex
	responses []client.Response
	factory   client.FakeResponse
	whenEmpty client.Response
	current   int
}

func NewFakeSequence(factory client.FakeResponse) *FakeSequence {
	return &FakeSequence{
		factory:   factory,
		responses: make([]client.Response, 0),
	}
}

func (r *FakeSequence) Push(response client.Response, count ...int) client.FakeSequence {
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

func (r *FakeSequence) PushStatus(status int, count ...int) client.FakeSequence {
	return r.Push(r.factory.Status(status), count...)
}

func (r *FakeSequence) PushString(body string, status int, count ...int) client.FakeSequence {
	return r.Push(r.factory.String(body, status), count...)
}

func (r *FakeSequence) WhenEmpty(response client.Response) client.FakeSequence {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.whenEmpty = response

	return r
}

func (r *FakeSequence) GetNext() client.Response {
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

	return nil
}
