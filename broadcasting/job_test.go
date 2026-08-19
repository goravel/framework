package broadcasting

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestBroadcastJob_Signature(t *testing.T) {
	job := &BroadcastJob{}
	assert.Equal(t, "goravel_broadcast", job.Signature())
}

func TestBroadcastJob_Handle_InvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "no arguments",
			args: nil,
		},
		{
			name: "empty arguments",
			args: []any{},
		},
		{
			name: "non-string argument",
			args: []any{42},
		},
		{
			name: "wrong type argument",
			args: []any{map[string]any{"key": "value"}},
		},
		{
			name: "invalid JSON",
			args: []any{"not-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &BroadcastJob{}
			err := job.Handle(tt.args...)
			assert.Error(t, err)
		})
	}
}

func TestBroadcastJob_Handle(t *testing.T) {
	setupConfig := func(t *testing.T) *mocksconfig.Config {
		mockConfig := mocksconfig.NewConfig(t)
		c := &Config{
			Default: "null",
			Connections: map[string]broadcasting.ConnectionConfig{
				"null": {Driver: "null"},
			},
		}
		mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.Anything).Run(func(key string, rawVal any) {
			cfg := rawVal.(*Config)
			*cfg = *c
		}).Return(nil).Once()
		return mockConfig
	}

	marshalItem := func(conns []string, timeout int64, tries int, backoff []int64) string {
		item := broadcastItem{
			Channels:    []string{"test-channel"},
			Event:       "test.event",
			Payload:     map[string]any{},
			Connections: conns,
			Timeout:     timeout,
			Tries:       tries,
			Backoff:     backoff,
		}
		data, _ := json.Marshal(item)
		return string(data)
	}

	t.Run("single attempt with no timeout", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// "broken" connection doesn't exist → always fails (single attempt).
		payload := marshalItem([]string{"broken"}, 0, 0, nil)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Less(t, elapsed, 50*time.Millisecond, "single attempt should complete quickly")
	})

	t.Run("single attempt fails on broken connection regardless of timeout", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"broken"}, 100, 0, nil)

		err := job.Handle(payload)
		assert.Error(t, err)
	})

	t.Run("succeeds with null connection", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"null"}, 0, 0, nil)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})

	t.Run("uses broadcast default connection when none specified", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem(nil, 0, 0, nil)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})

	t.Run("timeout is honored by driver via ctx", func(t *testing.T) {
		// The null driver ignores ctx, so a small timeout still succeeds.
		// This asserts the job synthesizes a bounded context without panicking.
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"null"}, 50, 0, nil)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})
}

func TestBroadcastJob_ShouldRetry(t *testing.T) {
	// marshalItem produces a valid payload whose connection is absent from the
	// config, so Handle fails inside broadcastToConns AFTER storing the parsed
	// item — the realistic pre-ShouldRetry state, since ShouldRetry is only
	// ever consulted after a task failed.
	marshalItem := func(tries int, backoff []int64) string {
		item := broadcastItem{
			Channels:    []string{"test-channel"},
			Event:       "test.event",
			Payload:     map[string]any{},
			Connections: []string{"broken"}, // not in config → driver-path failure
			Tries:       tries,
			Backoff:     backoff,
		}
		data, _ := json.Marshal(item)
		return string(data)
	}

	newJob := func(t *testing.T) *BroadcastJob {
		mockConfig := mocksconfig.NewConfig(t)
		c := &Config{
			Default: "null",
			Connections: map[string]broadcasting.ConnectionConfig{
				"null": {Driver: "null"},
			},
		}
		mockConfig.EXPECT().UnmarshalKey("broadcasting", mock.Anything).Run(func(key string, rawVal any) {
			cfg := rawVal.(*Config)
			*cfg = *c
		}).Return(nil).Maybe()
		return &BroadcastJob{config: mockConfig}
	}

	tests := []struct {
		name     string
		payload  string
		attempt  int
		maxTries int // the queue worker's tries (ignored when item.Tries > 0)
		err      error
		want     bool
		wantD    time.Duration
	}{
		{
			// No retry policy declared and the worker runs with Tries=1
			// (the default): single-shot.
			name:     "tries zero is single-shot",
			payload:  marshalItem(0, nil),
			attempt:  1,
			maxTries: 1,
			want:     false,
			wantD:    0,
		},
		{
			// No retry policy declared: the worker's Tries=3 governs.
			name:     "tries zero defers to worker tries first attempt retries",
			payload:  marshalItem(0, nil),
			attempt:  1,
			maxTries: 3,
			want:     true,
			wantD:    0,
		},
		{
			name:     "tries zero defers to worker tries second attempt retries",
			payload:  marshalItem(0, nil),
			attempt:  2,
			maxTries: 3,
			want:     true,
			wantD:    0,
		},
		{
			name:     "tries zero defers to worker tries third attempt stops",
			payload:  marshalItem(0, nil),
			attempt:  3,
			maxTries: 3,
			want:     false,
			wantD:    0,
		},
		{
			// Worker Tries=0 keeps the existing single-shot behavior.
			name:     "tries zero with worker tries zero is single-shot",
			payload:  marshalItem(0, nil),
			attempt:  1,
			maxTries: 0,
			want:     false,
			wantD:    0,
		},
		{
			name:     "tries 3 first attempt retries",
			payload:  marshalItem(3, nil),
			attempt:  1,
			maxTries: 1,
			want:     true,
			wantD:    0,
		},
		{
			name:     "tries 3 second attempt retries",
			payload:  marshalItem(3, nil),
			attempt:  2,
			maxTries: 1,
			want:     true,
			wantD:    0,
		},
		{
			name:     "tries 3 third attempt stops",
			payload:  marshalItem(3, nil),
			attempt:  3,
			maxTries: 1,
			want:     false,
			wantD:    0,
		},
		{
			// A broadcast-declared Tries always wins over the worker's:
			// with Tries=3 and worker maxTries=1, attempt 2 must retry.
			name:     "item tries overrides worker tries",
			payload:  marshalItem(3, nil),
			attempt:  2,
			maxTries: 1,
			want:     true,
			wantD:    0,
		},
		{
			name:     "backoff first attempt",
			payload:  marshalItem(4, []int64{1000, 2000}),
			attempt:  1,
			maxTries: 1,
			err:      errors.New("test error"), // err is ignored by ShouldRetry
			want:     true,
			wantD:    1 * time.Second,
		},
		{
			name:     "backoff second attempt",
			payload:  marshalItem(4, []int64{1000, 2000}),
			attempt:  2,
			maxTries: 1,
			want:     true,
			wantD:    2 * time.Second,
		},
		{
			name:     "backoff last value repeats",
			payload:  marshalItem(4, []int64{1000, 2000}),
			attempt:  3,
			maxTries: 1,
			want:     true,
			wantD:    2 * time.Second,
		},
		{
			name:     "backoff with attempt 0 is single-shot fallback",
			payload:  marshalItem(4, []int64{1000, 2000}),
			attempt:  0,
			maxTries: 1,
			want:     false,
			wantD:    0,
		},
		{
			name:     "backoff stop at final attempt before index",
			payload:  marshalItem(2, []int64{1000, 2000}),
			attempt:  2,
			maxTries: 1,
			want:     false,
			wantD:    0,
		},
		{
			name:     "backoff last attempt stops",
			payload:  marshalItem(4, []int64{1000, 2000}),
			attempt:  4,
			maxTries: 1,
			want:     false,
			wantD:    0,
		},
		{
			// Backoff is honored during worker-driven retries too (no
			// BroadcastTries, worker Tries=3).
			name:     "backoff applies during worker fallback",
			payload:  marshalItem(0, []int64{1000, 2000}),
			attempt:  2,
			maxTries: 3,
			want:     true,
			wantD:    2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := newJob(t)
			// Handle fails at the driver path ("broken" connection), which
			// retains the parsed item for ShouldRetry to read.
			assert.Error(t, job.Handle(tt.payload))

			retryable, delay := job.ShouldRetry(tt.err, tt.attempt, tt.maxTries)
			assert.Equal(t, tt.want, retryable)
			assert.Equal(t, tt.wantD, delay)
		})
	}

	// The following cases exercise stateful behavior — establishing a stale
	// item, then clearing or overriding it via a second Handle call — rather
	// than the single-shot decision the table above covers. They need multiple
	// sequential Handle calls, so they stay as separate t.Run cases.
	t.Run("invalid payload clears stale item", func(t *testing.T) {
		job := &BroadcastJob{}
		err := job.Handle("not-json")
		assert.Error(t, err)

		retryable, delay := job.ShouldRetry(nil, 1, 1)
		assert.False(t, retryable)
		assert.Equal(t, time.Duration(0), delay)

		// item == nil must stay the safe single-shot fallback even when the
		// worker's maxTries would otherwise allow a retry: with maxTries=3
		// and attempt=1 a worker-tries fallback would return true, so this
		// locks the "never the worker's tries" invariant.
		retryable, delay = job.ShouldRetry(nil, 1, 3)
		assert.False(t, retryable)
		assert.Equal(t, time.Duration(0), delay)
	})

	t.Run("wrong arity clears stale item", func(t *testing.T) {
		job := newJob(t)
		// The failed broadcast retains the parsed item (driver-path failure),
		// exactly the stale state a concurrent failed task would leave.
		assert.Error(t, job.Handle(marshalItem(3, nil)))
		retryable, _ := job.ShouldRetry(nil, 1, 1)
		assert.True(t, retryable) // stale policy present (Tries=3)

		assert.Error(t, job.Handle())
		retryable, _ = job.ShouldRetry(nil, 1, 1)
		assert.False(t, retryable)
	})

	t.Run("successful handle clears stale item", func(t *testing.T) {
		job := newJob(t)
		assert.Error(t, job.Handle(marshalItem(3, nil)))
		retryable, _ := job.ShouldRetry(nil, 1, 1)
		assert.True(t, retryable) // stale policy present (Tries=3)

		// A successful task releases the payload, so any concurrent policy
		// read falls back to the safe single-shot default instead of reading
		// the wrong task's policy.
		item := broadcastItem{
			Channels:    []string{"test-channel"},
			Event:       "test.event",
			Payload:     map[string]any{},
			Connections: []string{"null"},
			Tries:       5,
		}
		data, _ := json.Marshal(item)
		assert.NoError(t, job.Handle(string(data)))

		retryable, _ = job.ShouldRetry(nil, 1, 1)
		assert.False(t, retryable)
	})

	t.Run("non-string arg clears stale item", func(t *testing.T) {
		job := newJob(t)
		assert.Error(t, job.Handle(marshalItem(3, nil)))
		retryable, _ := job.ShouldRetry(nil, 1, 1)
		assert.True(t, retryable) // stale policy present (Tries=3)

		assert.Error(t, job.Handle(42))
		retryable, _ = job.ShouldRetry(nil, 1, 1)
		assert.False(t, retryable)
	})
}

func TestWithTimeout(t *testing.T) {
	t.Run("returns parent unchanged when timeoutMs <= 0", func(t *testing.T) {
		parent := context.Background()
		ctx, cancel := withTimeout(parent, 0)
		defer cancel()
		assert.Equal(t, parent, ctx)

		ctx2, cancel2 := withTimeout(parent, -1)
		defer cancel2()
		assert.Equal(t, parent, ctx2)
	})

	t.Run("derives bounded child when timeoutMs > 0", func(t *testing.T) {
		ctx, cancel := withTimeout(context.Background(), 50)
		defer cancel()
		_, ok := ctx.Deadline()
		assert.True(t, ok, "ctx should have a deadline")
	})

	t.Run("nil parent falls back to Background", func(t *testing.T) {
		ctx, cancel := withTimeout(nil, 0) //nolint:staticcheck // Testing nil parent behavior
		defer cancel()
		assert.NotNil(t, ctx)
	})

	t.Run("nil parent with timeout derives from Background", func(t *testing.T) {
		ctx, cancel := withTimeout(nil, 100) //nolint:staticcheck // Testing nil parent behavior
		defer cancel()
		_, ok := ctx.Deadline()
		assert.True(t, ok, "ctx should have a deadline even with nil parent")
	})
}
