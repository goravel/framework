package log

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/log"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestHook_Fire(t *testing.T) {
	var (
		mockHook *mockslog.Hook

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
		name        string
		record      slog.Record
		setup       func()
		expectError error
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
			setup: func() {
				mockHook.EXPECT().Fire(&Entry{
					ctx:  ctx,
					code: code,
					data: map[string]any{
						"root": map[string]any{
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
						},
					},
					domain:     domain,
					hint:       hint,
					level:      log.FromSlog(level),
					message:    message,
					owner:      owner,
					request:    request,
					response:   response,
					stacktrace: stacktrace,
					tags:       tags,
					time:       now,
					user:       user,
					with:       with,
				}).Return(nil).Once()
			},
		},
		{
			name: "empty data",
			record: func() slog.Record {
				return slog.NewRecord(now, level, message, 0)
			}(),
			setup: func() {
				mockHook.EXPECT().Fire(&Entry{
					ctx:     ctx,
					data:    map[string]any{},
					level:   log.FromSlog(level),
					message: message,
					time:    now,
				}).Return(nil).Once()
			},
		},
		{
			name: "Fire returns error",
			record: func() slog.Record {
				return slog.NewRecord(now, level, message, 0)
			}(),
			setup: func() {
				mockHook.EXPECT().Fire(&Entry{
					ctx:     ctx,
					data:    map[string]any{},
					level:   log.FromSlog(level),
					message: message,
					time:    now,
				}).Return(assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHook = mockslog.NewHook(t)
			tt.setup()

			hook := &Hook{instance: mockHook}

			err := hook.Fire(ctx, tt.record)

			assert.Equal(t, tt.expectError, err)
		})
	}
}
