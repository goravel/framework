package log

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/log"
)

func TestLegacyHandlerAdapter(t *testing.T) {
	var (
		ctx        = context.Background()
		code       = "123"
		domain     = "example.com"
		hint       = "hint"
		level      = slog.LevelInfo
		message    = "message"
		owner      = "owner"
		request    = map[string]any{"key": "value"}
		response   = map[string]any{"key": "value"}
		stacktrace = map[string]any{"key": "value"}
		now        = time.Now()
		tags       = []string{"tag1", "tag2"}
		user       = "user"
		with       = map[string]any{"key": "value"}
	)

	tests := []struct {
		name   string
		record slog.Record
		setup  func() slog.Handler
	}{
		{
			name: "full data",
			record: func() slog.Record {
				r := slog.NewRecord(now, level, message, 0)
				r.AddAttrs(slog.Any("root", map[string]any{
					"code":       code,
					"domain":     domain,
					"hint":       hint,
					"owner":      owner,
					"request":    request,
					"response":   response,
					"stacktrace": stacktrace,
					"tags":       tags,
					"user":       user,
					"with":       with,
				}))
				return r
			}(),
			setup: func() slog.Handler {
				// Create a simple test handler
				return &testHandler{enabled: true}
			},
		},
		{
			name: "empty data",
			record: func() slog.Record {
				return slog.NewRecord(now, level, message, 0)
			}(),
			setup: func() slog.Handler {
				return &testHandler{enabled: true}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setup()
			adapter := NewLegacyHandlerAdapter(handler)

			err := adapter.Handle(ctx, tt.record)

			assert.Nil(t, err)
		})
	}
}

func TestCreateEntryFromRecord(t *testing.T) {
	now := time.Now()
	ctx := context.Background()
	
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	record.AddAttrs(slog.Any("root", map[string]any{
		"code":   "123",
		"domain": "test.com",
	}))
	
	entry := createEntryFromRecord(ctx, record)
	
	assert.NotNil(t, entry)
	assert.Equal(t, "test message", entry.Message())
	assert.Equal(t, log.FromSlog(slog.LevelInfo), entry.Level())
	assert.Equal(t, "123", entry.Code())
	assert.Equal(t, "test.com", entry.Domain())
}

// testHandler is a simple handler for testing
type testHandler struct {
	enabled bool
	records []slog.Record
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabled
}

func (h *testHandler) Handle(ctx context.Context, record slog.Record) error {
	h.records = append(h.records, record)
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}
