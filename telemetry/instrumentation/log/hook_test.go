package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/logtest"

	contractslog "github.com/goravel/framework/contracts/log"
)

func TestHookLevels(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		wantLevels []contractslog.Level
	}{
		{
			name:    "enabled returns all levels",
			enabled: true,
			wantLevels: []contractslog.Level{
				contractslog.DebugLevel,
				contractslog.InfoLevel,
				contractslog.WarningLevel,
				contractslog.ErrorLevel,
				contractslog.FatalLevel,
				contractslog.PanicLevel,
			},
		},
		{
			name:       "disabled returns nil",
			enabled:    false,
			wantLevels: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &hook{enabled: tt.enabled}
			assert.Equal(t, tt.wantLevels, h.Levels())
		})
	}
}

func TestHookFire(t *testing.T) {
	const loggerName = "test-logger"
	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	ctx := context.WithValue(context.Background(), "request_id", "req-123")

	tests := []struct {
		name    string
		entry   *TestEntry
		want    logtest.Recording
		wantErr error
	}{
		{
			name: "emits an empty log entry",
			entry: &TestEntry{
				ctx:   context.Background(),
				level: contractslog.InfoLevel,
				time:  now,
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:           context.Background(),
						Timestamp:         now,
						Severity:          log.SeverityInfo,
						SeverityText:      "info",
						Body:              log.StringValue(""),
						ObservedTimestamp: time.Time{},
					},
				},
			},
		},
		{
			name: "emits a debug log with message",
			entry: &TestEntry{
				ctx:     ctx,
				level:   contractslog.DebugLevel,
				time:    now,
				message: "debug message",
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:      ctx,
						Timestamp:    now,
						Severity:     log.SeverityDebug,
						SeverityText: "debug",
						Body:         log.StringValue("debug message"),
					},
				},
			},
		},
		{
			name: "emits an error log with standard fields (code, domain, hint)",
			entry: &TestEntry{
				ctx:     ctx,
				level:   contractslog.ErrorLevel,
				time:    now,
				message: "something went wrong",
				code:    "ERR_500",
				domain:  "payment",
				hint:    "check balance",
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:      ctx,
						Timestamp:    now,
						Severity:     log.SeverityError,
						SeverityText: "error",
						Body:         log.StringValue("something went wrong"),
						Attributes: []log.KeyValue{
							log.String("code", "ERR_500"),
							log.String("domain", "payment"),
							log.String("hint", "check balance"),
						},
					},
				},
			},
		},
		{
			name: "emits a log with context data (With/Data)",
			entry: &TestEntry{
				ctx:   ctx,
				level: contractslog.InfoLevel,
				time:  now,
				with: map[string]any{
					"foo": "bar",
				},
				data: map[string]any{
					"user_id": 42,
					"active":  true,
				},
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:      ctx,
						Timestamp:    now,
						Severity:     log.SeverityInfo,
						SeverityText: "info",
						Body:         log.StringValue(""),
						Attributes: []log.KeyValue{
							log.String("foo", "bar"),
							log.Int64("user_id", 42),
							log.Bool("active", true),
						},
					},
				},
			},
		},
		{
			name: "emits a log with complex types (User, Tags, Request)",
			entry: &TestEntry{
				ctx:   ctx,
				level: contractslog.WarningLevel,
				time:  now,
				user: map[string]any{
					"id":   1,
					"role": "admin",
				},
				tags: []string{"critical", "auth"},
				request: map[string]any{
					"method": "GET",
					"url":    "/login",
				},
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:      ctx,
						Timestamp:    now,
						Severity:     log.SeverityWarn,
						SeverityText: "warning",
						Body:         log.StringValue(""),
						Attributes: []log.KeyValue{
							log.Map("user",
								log.Int64("id", 1),
								log.String("role", "admin"),
							),
							log.Slice("tags",
								log.StringValue("critical"),
								log.StringValue("auth"),
							),
							log.Map("request",
								log.String("method", "GET"),
								log.String("url", "/login"),
							),
						},
					},
				},
			},
		},
		{
			name: "emits a log with Panic level",
			entry: &TestEntry{
				ctx:   ctx,
				level: contractslog.PanicLevel,
				time:  now,
			},
			want: logtest.Recording{
				logtest.Scope{Name: loggerName}: {
					{
						Context:      ctx,
						Timestamp:    now,
						Severity:     log.SeverityFatal4,
						SeverityText: "panic",
						Body:         log.StringValue(""),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := logtest.NewRecorder()
			h := &hook{
				enabled: true,
				logger:  recorder.Logger(loggerName),
			}

			err := h.Fire(tt.entry)
			assert.Equal(t, tt.wantErr, err)

			result := recorder.Result()

			for i := range result[logtest.Scope{Name: loggerName}] {
				result[logtest.Scope{Name: loggerName}][i].ObservedTimestamp = time.Time{}
			}

			logtest.AssertEqual(t, tt.want, result)
		})
	}
}

type TestEntry struct {
	ctx      context.Context
	level    contractslog.Level
	time     time.Time
	message  string
	with     map[string]any
	data     contractslog.Data
	tags     []string
	user     any
	owner    any
	code     string
	domain   string
	hint     string
	request  map[string]any
	response map[string]any
	trace    map[string]any
}

func (e *TestEntry) Context() context.Context  { return e.ctx }
func (e *TestEntry) Level() contractslog.Level { return e.level }
func (e *TestEntry) Time() time.Time           { return e.time }
func (e *TestEntry) Message() string           { return e.message }
func (e *TestEntry) With() map[string]any      { return e.with }
func (e *TestEntry) Data() contractslog.Data   { return e.data }
func (e *TestEntry) Tags() []string            { return e.tags }
func (e *TestEntry) User() any                 { return e.user }
func (e *TestEntry) Owner() any                { return e.owner }
func (e *TestEntry) Code() string              { return e.code }
func (e *TestEntry) Domain() string            { return e.domain }
func (e *TestEntry) Hint() string              { return e.hint }
func (e *TestEntry) Request() map[string]any   { return e.request }
func (e *TestEntry) Response() map[string]any  { return e.response }
func (e *TestEntry) Trace() map[string]any     { return e.trace }
