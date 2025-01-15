package log

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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
		level      = logrus.InfoLevel
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
		entry       *logrus.Entry
		setup       func()
		expectError error
	}{
		{
			name: "full data",
			entry: &logrus.Entry{
				Context: ctx,
				Data: logrus.Fields{
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
				Level:   level,
				Time:    now,
				Message: message,
			},
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
					level:      log.Level(level),
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
			entry: &logrus.Entry{
				Context: ctx,
				Data:    logrus.Fields{},
				Level:   level,
				Time:    now,
				Message: message,
			},
			setup: func() {
				mockHook.EXPECT().Fire(&Entry{
					ctx:     ctx,
					data:    map[string]any{},
					level:   log.Level(level),
					message: message,
					time:    now,
				}).Return(nil).Once()
			},
		},
		{
			name: "Fire returns error",
			entry: &logrus.Entry{
				Context: ctx,
				Data:    logrus.Fields{},
				Level:   level,
				Time:    now,
				Message: message,
			},
			setup: func() {
				mockHook.EXPECT().Fire(&Entry{
					ctx:     ctx,
					data:    map[string]any{},
					level:   log.Level(level),
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

			err := hook.Fire(tt.entry)

			assert.Equal(t, tt.expectError, err)
		})
	}
}
