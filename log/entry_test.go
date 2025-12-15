package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/log"
)

func TestEntry(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	
	entry := &Entry{
		ctx:     ctx,
		code:    "TEST001",
		domain:  "test.domain",
		hint:    "test hint",
		level:   log.InfoLevel,
		message: "test message",
		owner:   "test owner",
		time:    now,
		user:    "test user",
		tags:    []string{"tag1", "tag2"},
		request: map[string]any{"method": "GET"},
		response: map[string]any{"status": 200},
		stacktrace: map[string]any{"file": "test.go"},
		with: map[string]any{"key": "value"},
		data: map[string]any{"data": "value"},
	}
	
	assert.Equal(t, "TEST001", entry.Code())
	assert.Equal(t, ctx, entry.Context())
	assert.Equal(t, "test.domain", entry.Domain())
	assert.Equal(t, "test hint", entry.Hint())
	assert.Equal(t, log.InfoLevel, entry.Level())
	assert.Equal(t, "test message", entry.Message())
	assert.Equal(t, "test owner", entry.Owner())
	assert.Equal(t, now, entry.Time())
	assert.Equal(t, "test user", entry.User())
	assert.Equal(t, []string{"tag1", "tag2"}, entry.Tags())
	assert.Equal(t, map[string]any{"method": "GET"}, entry.Request())
	assert.Equal(t, map[string]any{"status": 200}, entry.Response())
	assert.Equal(t, map[string]any{"file": "test.go"}, entry.Trace())
	assert.Equal(t, map[string]any{"key": "value"}, entry.With())
	assert.Equal(t, log.Data{"data": "value"}, entry.Data())
}
