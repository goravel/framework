package ai

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsai "github.com/goravel/framework/contracts/ai"
	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mockshttp "github.com/goravel/framework/mocks/http"
)

type streamableTestResponse struct {
	text  string
	usage contractsai.Usage
}

func (r *streamableTestResponse) Text() string                      { return r.text }
func (r *streamableTestResponse) Usage() contractsai.Usage          { return r.usage }
func (r *streamableTestResponse) ToolCalls() []contractsai.ToolCall { return nil }

func (r *streamableTestResponse) Then(callback func(contractsai.Response)) contractsai.Response {
	if callback != nil {
		callback(r)
	}

	return r
}

type streamableTestUsage struct {
	input  int
	output int
	total  int
}

func (u *streamableTestUsage) Input() int  { return u.input }
func (u *streamableTestUsage) Output() int { return u.output }
func (u *streamableTestUsage) Total() int  { return u.total }

type streamEventSnapshot struct {
	Type  contractsai.StreamEventType
	Delta string
	Error string
	Usage *streamableTestUsage
}

type recordingStreamWriter struct {
	writes []string

	writeErrAt int
	writeErr   error
	flushErr   error
	flushCount int
}

func (w *recordingStreamWriter) Write(data []byte) (int, error) {
	return w.WriteString(string(data))
}

func (w *recordingStreamWriter) WriteString(data string) (int, error) {
	if w.writeErr != nil && w.writeErrAt > 0 && len(w.writes)+1 == w.writeErrAt {
		return 0, w.writeErr
	}

	w.writes = append(w.writes, data)
	return len(data), nil
}

func (w *recordingStreamWriter) Flush() error {
	w.flushCount++
	return w.flushErr
}

func normalizeStreamEvents(events []contractsai.StreamEvent) []streamEventSnapshot {
	normalized := make([]streamEventSnapshot, 0, len(events))
	for _, event := range events {
		entry := streamEventSnapshot{
			Type:  event.Type,
			Delta: event.Delta,
			Error: event.Error,
		}
		if event.Usage != nil {
			entry.Usage = &streamableTestUsage{
				input:  event.Usage.Input(),
				output: event.Usage.Output(),
				total:  event.Usage.Total(),
			}
		}

		normalized = append(normalized, entry)
	}

	return normalized
}

type StreamableResponseTestSuite struct {
	suite.Suite
}

func TestStreamableResponseTestSuite(t *testing.T) {
	suite.Run(t, &StreamableResponseTestSuite{})
}

func (s *StreamableResponseTestSuite) TestEach() {
	s.Run("returns error when runner is missing", func() {
		stream := NewStreamableResponse(context.Background(), nil)

		err := stream.Each(nil)
		s.Equal(errors.AIStreamRunnerRequired, err)
	})

	s.Run("drains queued events before returning runner error", func() {
		events := []contractsai.StreamEvent{
			{Type: contractsai.StreamEventTypeTextDelta, Delta: "a"},
			{Type: contractsai.StreamEventTypeTextDelta, Delta: "b"},
		}
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(events[0]); err != nil {
				return nil, err
			}
			if err := emit(events[1]); err != nil {
				return nil, err
			}

			return nil, assert.AnError
		})

		var got []contractsai.StreamEvent
		err := stream.Each(func(event contractsai.StreamEvent) error {
			got = append(got, event)
			return nil
		})

		s.Equal(assert.AnError, err)
		s.Equal(normalizeStreamEvents(events), normalizeStreamEvents(got))
	})

	s.Run("aborts runner when callback fails", func() {
		canceled := make(chan struct{})
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "first"}); err != nil {
				return nil, err
			}

			<-ctx.Done()
			close(canceled)
			return nil, ctx.Err()
		})

		err := stream.Each(func(_ contractsai.StreamEvent) error {
			return assert.AnError
		})

		s.Equal(assert.AnError, err)
		select {
		case <-canceled:
		case <-time.After(time.Second):
			s.FailNow("runner context was not canceled")
		}
	})

	s.Run("supports nil callback", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "ignored"}); err != nil {
				return nil, err
			}

			return nil, nil
		})

		err := stream.Each(nil)
		s.NoError(err)
	})
}

func (s *StreamableResponseTestSuite) TestThen() {
	makeSuccessStream := func() contractsai.StreamableResponse {
		return NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
				return nil, err
			}

			return &streamableTestResponse{text: "final"}, nil
		})
	}

	s.Run("executes callback after successful completion", func() {
		stream := makeSuccessStream()
		called := 0

		stream.Then(func(resp contractsai.Response) {
			called++
			s.Equal("final", resp.Text())
		})

		err := stream.Each(nil)
		s.NoError(err)
		s.Equal(1, called)
	})

	s.Run("executes callback immediately when called after completion", func() {
		stream := makeSuccessStream()
		s.Require().NoError(stream.Each(nil))

		called := 0
		stream.Then(func(resp contractsai.Response) {
			called++
			s.Equal("final", resp.Text())
		})
		s.Equal(1, called)
	})

	s.Run("does not execute callback when stream already failed", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			return nil, assert.AnError
		})
		s.Require().Equal(assert.AnError, stream.Each(nil))

		called := 0
		stream.Then(func(_ contractsai.Response) {
			called++
		})
		s.Equal(0, called)
	})
}

func (s *StreamableResponseTestSuite) TestHTTPResponse() {
	prepareHTTP := func(expectHeaders bool, code int, stream contractsai.StreamableResponse, options ...contractsai.StreamOption) (*recordingStreamWriter, error) {
		s.T().Helper()

		mockCtx := mockshttp.NewContext(s.T())
		mockContextResponse := mockshttp.NewContextResponse(s.T())
		writer := &recordingStreamWriter{}

		mockCtx.EXPECT().Response().Return(mockContextResponse).Once()
		if expectHeaders {
			mockContextResponse.EXPECT().Header("Content-Type", streamContentType).Return(mockContextResponse).Once()
			mockContextResponse.EXPECT().Header("Cache-Control", streamCacheControl).Return(mockContextResponse).Once()
			mockContextResponse.EXPECT().Header("Connection", streamConnection).Return(mockContextResponse).Once()
		}

		var streamErr error
		mockContextResponse.EXPECT().
			Stream(code, mock.IsType((func(contractshttp.StreamWriter) error)(nil))).
			RunAndReturn(func(_ int, step func(contractshttp.StreamWriter) error) contractshttp.Response {
				streamErr = step(writer)
				return nil
			}).
			Once()

		stream.HTTPResponse(mockCtx, options...)
		return writer, streamErr
	}

	s.Run("renders with default headers and SSE payload", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "hi"}); err != nil {
				return nil, err
			}
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone, Usage: &streamableTestUsage{input: 1, output: 2, total: 3}}); err != nil {
				return nil, err
			}

			return &streamableTestResponse{text: "hi"}, nil
		})

		writer, streamErr := prepareHTTP(true, defaultStreamResponseCode, stream)
		s.NoError(streamErr)
		s.Equal([]string{
			"event: text_delta\n",
			"data: {\"delta\":\"hi\"}\n\n",
			"event: done\n",
			"data: {\"usage\":{\"input\":1,\"output\":2,\"total\":3}}\n\n",
		}, writer.writes)
		s.Equal(2, writer.flushCount)
	})

	s.Run("uses custom render and code", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "a"}); err != nil {
				return nil, err
			}
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeDone}); err != nil {
				return nil, err
			}

			return nil, nil
		})

		var seen []contractsai.StreamEvent
		customRender := func(_ contractshttp.StreamWriter, event contractsai.StreamEvent) error {
			seen = append(seen, event)
			return nil
		}

		writer, streamErr := prepareHTTP(false, 207, stream, WithStreamCode(207), WithStreamRender(customRender))
		s.NoError(streamErr)
		s.Equal(normalizeStreamEvents([]contractsai.StreamEvent{
			{Type: contractsai.StreamEventTypeTextDelta, Delta: "a"},
			{Type: contractsai.StreamEventTypeDone},
		}), normalizeStreamEvents(seen))
		s.Empty(writer.writes)
	})

	s.Run("suppresses provider error when error event is already rendered", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeError, Error: "provider failed"}); err != nil {
				return nil, err
			}

			return nil, assert.AnError
		})

		writer, streamErr := prepareHTTP(true, defaultStreamResponseCode, stream)
		s.NoError(streamErr)
		s.Equal([]string{
			"event: error\n",
			"data: {\"error\":\"provider failed\"}\n\n",
		}, writer.writes)
		s.Equal(1, writer.flushCount)
	})

	s.Run("returns renderer error when renderer fails", func() {
		renderErr := stderrors.New("render failed")
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			if err := emit(contractsai.StreamEvent{Type: contractsai.StreamEventTypeError, Error: "provider failed"}); err != nil {
				return nil, err
			}

			return nil, assert.AnError
		})

		_, streamErr := prepareHTTP(false, defaultStreamResponseCode, stream, WithStreamRender(func(_ contractshttp.StreamWriter, _ contractsai.StreamEvent) error {
			return renderErr
		}))

		s.Equal(renderErr, streamErr)
	})

	s.Run("returns provider error when no error event is emitted", func() {
		stream := NewStreamableResponse(context.Background(), func(ctx context.Context, emit func(contractsai.StreamEvent) error) (contractsai.Response, error) {
			return nil, assert.AnError
		})

		writer, streamErr := prepareHTTP(true, defaultStreamResponseCode, stream)
		s.Equal(assert.AnError, streamErr)
		s.Empty(writer.writes)
		s.Equal(0, writer.flushCount)
	})
}

func TestDefaultStreamRender(t *testing.T) {
	tests := []struct {
		name         string
		event        contractsai.StreamEvent
		writer       *recordingStreamWriter
		expectWrites []string
		expectFlush  int
		expectErr    error
	}{
		{
			name:         "renders text delta payload",
			event:        contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "hello"},
			writer:       &recordingStreamWriter{},
			expectWrites: []string{"event: text_delta\n", "data: {\"delta\":\"hello\"}\n\n"},
			expectFlush:  1,
		},
		{
			name: "renders usage payload",
			event: contractsai.StreamEvent{
				Type:  contractsai.StreamEventTypeDone,
				Usage: &streamableTestUsage{input: 3, output: 4, total: 7},
			},
			writer:       &recordingStreamWriter{},
			expectWrites: []string{"event: done\n", "data: {\"usage\":{\"input\":3,\"output\":4,\"total\":7}}\n\n"},
			expectFlush:  1,
		},
		{
			name: "renders tool call payload",
			event: contractsai.StreamEvent{
				Type: contractsai.StreamEventTypeToolCall,
				ToolCalls: []contractsai.ToolCall{{
					ID:   "call_1",
					Name: "get_weather",
					Args: map[string]any{"city": "London"},
				}},
			},
			writer:       &recordingStreamWriter{},
			expectWrites: []string{"event: tool_call\n", "data: {\"tool_calls\":[{\"id\":\"call_1\",\"name\":\"get_weather\",\"args\":{\"city\":\"London\"}}]}\n\n"},
			expectFlush:  1,
		},
		{
			name:      "returns first write error",
			event:     contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "hello"},
			writer:    &recordingStreamWriter{writeErrAt: 1, writeErr: assert.AnError},
			expectErr: assert.AnError,
		},
		{
			name:   "returns second write error",
			event:  contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "hello"},
			writer: &recordingStreamWriter{writeErrAt: 2, writeErr: assert.AnError},
			expectWrites: []string{
				"event: text_delta\n",
			},
			expectErr: assert.AnError,
		},
		{
			name:   "returns flush error",
			event:  contractsai.StreamEvent{Type: contractsai.StreamEventTypeTextDelta, Delta: "hello"},
			writer: &recordingStreamWriter{flushErr: assert.AnError},
			expectWrites: []string{
				"event: text_delta\n",
				"data: {\"delta\":\"hello\"}\n\n",
			},
			expectFlush: 1,
			expectErr:   assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := defaultStreamRender(tt.writer, tt.event)

			assert.Equal(t, tt.expectErr, err)
			assert.Equal(t, tt.expectWrites, tt.writer.writes)
			assert.Equal(t, tt.expectFlush, tt.writer.flushCount)
		})
	}
}
