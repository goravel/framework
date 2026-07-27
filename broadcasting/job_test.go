package broadcasting

import (
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

func TestBroadcastJob_Handle_Retry(t *testing.T) {
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

	marshalItem := func(conns []string, tries int, backoff, timeout int64) string {
		item := broadcastItem{
			Channels:    []string{"test-channel"},
			Event:       "test.event",
			Payload:     map[string]any{},
			Connections: conns,
			Tries:       tries,
			Backoff:     backoff,
			Timeout:     timeout,
		}
		data, _ := json.Marshal(item)
		return string(data)
	}

	t.Run("single attempt with no retry config", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// "broken" connection doesn't exist → always fails
		payload := marshalItem([]string{"broken"}, 0, 0, 0)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Less(t, elapsed, 50*time.Millisecond, "single attempt should complete quickly")
	})

	t.Run("retries up to tries count", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// 3 tries, no backoff → 3 fast attempts
		payload := marshalItem([]string{"broken"}, 3, 0, 0)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Less(t, elapsed, 50*time.Millisecond, "3 attempts with no backoff should complete quickly")
	})

	t.Run("retries with backoff delay", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// 3 tries, 50ms backoff → 2 sleeps of 50ms each = at least 100ms
		payload := marshalItem([]string{"broken"}, 3, 50, 0)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "should sleep between retries")
		assert.Less(t, elapsed, 500*time.Millisecond, "should not exceed expected backoff time")
	})

	t.Run("timeout stops retries early", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// 10 tries, 100ms backoff, 150ms timeout
		// After 1st attempt + 100ms sleep ≈ 100ms elapsed
		// 100ms + next 100ms backoff >= 150ms → break after 2 attempts
		// Total elapsed ≈ 100ms (far less than 1000ms for 10*100ms)
		payload := marshalItem([]string{"broken"}, 10, 100, 150)

		start := time.Now()
		err := job.Handle(payload)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Less(t, elapsed, 500*time.Millisecond, "timeout should stop retries early")
	})

	t.Run("succeeds within retry limit", func(t *testing.T) {
		job := &BroadcastJob{config: setupConfig(t)}
		// Use an existing "null" connection → never fails (no app dependency)
		payload := marshalItem([]string{"null"}, 3, 0, 0)

		err := job.Handle(payload)
		assert.NoError(t, err)
	})
}
