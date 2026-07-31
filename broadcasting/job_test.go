package broadcasting

import (
	"context"
	"encoding/json"
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

	marshalItem := func(conns []string, timeout int64) string {
		item := broadcastItem{
			Channels:    []string{"test-channel"},
			Event:       "test.event",
			Payload:     map[string]any{},
			Connections: conns,
			Timeout:     timeout,
		}
		data, _ := json.Marshal(item)
		return string(data)
	}

	t.Run("single attempt with no timeout", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// "broken" connection doesn't exist → always fails (single attempt).
		payload := marshalItem([]string{"broken"}, 0)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Less(t, elapsed, 50*time.Millisecond, "single attempt should complete quickly")
	})

	t.Run("single attempt fails on broken connection regardless of timeout", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"broken"}, 100)

		err := job.Handle(payload)
		assert.Error(t, err)
	})

	t.Run("succeeds with null connection", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"null"}, 0)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})

	t.Run("uses broadcast default connection when none specified", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem(nil, 0)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})

	t.Run("timeout is honored by driver via ctx", func(t *testing.T) {
		// The null driver ignores ctx, so a small timeout still succeeds.
		// This asserts the job synthesizes a bounded context without panicking.
		job := &BroadcastJob{config: setupConfig(t)}
		payload := marshalItem([]string{"null"}, 50)

		err := job.Handle(payload)
		assert.NoError(t, err)
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
