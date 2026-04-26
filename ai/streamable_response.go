package ai

import (
	"context"
	"encoding/json"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
)

var (
	_ contractsai.StreamableResponse = (*streamableResponse)(nil)
)

const (
	defaultStreamResponseCode = 200
	streamContentType         = "text/event-stream"
	streamCacheControl        = "no-cache"
	streamConnection          = "keep-alive"
)

type StreamRunner func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error)

type streamableResponse struct {
	ctx    context.Context
	runner StreamRunner

	mu       sync.Mutex
	cond     *sync.Cond
	started  bool
	finished bool
	cancel   context.CancelFunc

	events   []contractsai.StreamEvent
	response contractsai.Response
	err      error

	thenCallbacks []func(contractsai.Response)
}

func NewStreamableResponse(ctx context.Context, runner StreamRunner) contractsai.StreamableResponse {
	if ctx == nil {
		ctx = context.Background()
	}

	stream := &streamableResponse{
		ctx:    ctx,
		runner: runner,
	}
	stream.cond = sync.NewCond(&stream.mu)

	return stream
}

func (r *streamableResponse) Each(callback func(contractsai.StreamEvent) error) error {
	r.start()

	for {
		r.mu.Lock()
		for len(r.events) == 0 && !r.finished {
			r.cond.Wait()
		}

		if len(r.events) > 0 {
			event := r.events[0]
			r.events[0] = contractsai.StreamEvent{}
			r.events = r.events[1:]
			if len(r.events) == 0 {
				r.events = nil
			}
			r.mu.Unlock()

			if callback != nil {
				if err := callback(event); err != nil {
					r.abort()
					return err
				}
			}

			continue
		}

		err := r.err
		r.mu.Unlock()
		return err
	}
}

func (r *streamableResponse) Then(callback func(contractsai.Response)) contractsai.StreamableResponse {
	if callback == nil {
		return r
	}

	r.mu.Lock()
	if r.finished && r.err == nil && r.response != nil {
		response := r.response
		r.mu.Unlock()
		callback(response)
		return r
	}

	if !r.finished {
		r.thenCallbacks = append(r.thenCallbacks, callback)
	}
	r.mu.Unlock()

	return r
}

func (r *streamableResponse) HTTPResponse(ctx contractshttp.Context, options ...contractsai.StreamOption) contractshttp.Response {
	ops := contractsai.StreamOptions{
		Code: defaultStreamResponseCode,
	}
	for _, option := range options {
		if option != nil {
			option(&ops)
		}
	}

	render := ops.Render
	response := ctx.Response()
	if render == nil {
		render = defaultStreamRender
		response = response.
			Header("Content-Type", streamContentType).
			Header("Cache-Control", streamCacheControl).
			Header("Connection", streamConnection)
	}

	return response.Stream(ops.Code, func(w contractshttp.StreamWriter) error {
		hasProviderErrorEvent := false
		rendererFailed := false
		err := r.Each(func(event contractsai.StreamEvent) error {
			if event.Type == contractsai.StreamEventTypeError {
				hasProviderErrorEvent = true
			}

			if renderErr := render(w, event); renderErr != nil {
				rendererFailed = true
				return renderErr
			}

			return nil
		})
		if err != nil && hasProviderErrorEvent && !rendererFailed {
			return nil
		}

		return err
	})
}

type streamToolCallPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Args any    `json:"args,omitempty"`
}

type streamEventPayload struct {
	Delta     string                  `json:"delta,omitempty"`
	Usage     *streamUsagePayload     `json:"usage,omitempty"`
	Error     string                  `json:"error,omitempty"`
	ToolCalls []streamToolCallPayload `json:"tool_calls,omitempty"`
}

type streamUsagePayload struct {
	Input  int `json:"input"`
	Output int `json:"output"`
	Total  int `json:"total"`
}

func defaultStreamRender(w contractshttp.StreamWriter, event contractsai.StreamEvent) error {
	payload := streamEventPayload{
		Delta: event.Delta,
		Error: event.Error,
	}
	if event.Usage != nil {
		payload.Usage = &streamUsagePayload{
			Input:  event.Usage.Input(),
			Output: event.Usage.Output(),
			Total:  event.Usage.Total(),
		}
	}
	for _, tc := range event.ToolCalls {
		payload.ToolCalls = append(payload.ToolCalls, streamToolCallPayload{
			ID:   tc.ID,
			Name: tc.Name,
			Args: tc.Args,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := w.WriteString("event: " + string(event.Type) + "\n"); err != nil {
		return err
	}
	if _, err := w.WriteString("data: " + string(data) + "\n\n"); err != nil {
		return err
	}

	return w.Flush()
}

func (r *streamableResponse) start() {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return
	}
	r.started = true

	streamCtx, cancel := context.WithCancel(r.ctx)
	r.cancel = cancel
	r.mu.Unlock()

	go r.run(streamCtx)
}

func (r *streamableResponse) run(ctx context.Context) {
	if r.runner == nil {
		r.complete(nil, errors.AIStreamRunnerRequired)
		return
	}

	response, err := r.runner(ctx, func(event contractsai.StreamEvent) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		r.mu.Lock()
		r.events = append(r.events, event)
		r.cond.Broadcast()
		r.mu.Unlock()

		return nil
	})
	r.complete(response, err)
}

func (r *streamableResponse) complete(response contractsai.Response, err error) {
	var callbacks []func(contractsai.Response)

	r.mu.Lock()
	r.response = response
	r.err = err
	if err == nil && response != nil {
		callbacks = append(callbacks, r.thenCallbacks...)
	}
	r.thenCallbacks = nil
	r.mu.Unlock()

	if err == nil && response != nil {
		for _, callback := range callbacks {
			callback(response)
		}
	}

	r.mu.Lock()
	r.err = err
	r.finished = true
	cancel := r.cancel
	r.cond.Broadcast()
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (r *streamableResponse) abort() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}
