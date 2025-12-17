package log

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/logtest"

	contractslog "github.com/goravel/framework/contracts/log"
)

type HandlerTestSuite struct {
	suite.Suite
	recorder   *logtest.Recorder
	handler    *handler
	loggerName string
	ctx        context.Context
	now        time.Time
}

func (s *HandlerTestSuite) SetupTest() {
	s.loggerName = "test-logger"
	s.recorder = logtest.NewRecorder()
	s.handler = &handler{
		logger: s.recorder.Logger(s.loggerName),
	}
	s.now = time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	type ctxKey string
	s.ctx = context.WithValue(context.Background(), ctxKey("request_id"), "req-123")
}

func (s *HandlerTestSuite) TestEnabled() {
	s.True(s.handler.Enabled(contractslog.LevelDebug))
	s.True(s.handler.Enabled(contractslog.LevelInfo))
	s.True(s.handler.Enabled(contractslog.LevelError))
}

func (s *HandlerTestSuite) TestHandleEmptyEntry() {
	entry := &TestEntry{
		ctx:   context.Background(),
		level: contractslog.LevelInfo,
		time:  s.now,
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      context.Background(),
				Timestamp:    s.now,
				Severity:     log.SeverityInfo,
				SeverityText: "info",
				Body:         log.StringValue(""),
			},
		},
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) TestHandleDebugWithMessage() {
	entry := &TestEntry{
		ctx:     s.ctx,
		level:   contractslog.LevelDebug,
		time:    s.now,
		message: "debug message",
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      s.ctx,
				Timestamp:    s.now,
				Severity:     log.SeverityDebug,
				SeverityText: "debug",
				Body:         log.StringValue("debug message"),
			},
		},
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) TestHandleErrorWithStandardFields() {
	entry := &TestEntry{
		ctx:     s.ctx,
		level:   contractslog.LevelError,
		time:    s.now,
		message: "something went wrong",
		code:    "ERR_500",
		domain:  "payment",
		hint:    "check balance",
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      s.ctx,
				Timestamp:    s.now,
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
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) TestHandleWithContextData() {
	entry := &TestEntry{
		ctx:   s.ctx,
		level: contractslog.LevelInfo,
		time:  s.now,
		with: map[string]any{
			"foo": "bar",
		},
		data: map[string]any{
			"user_id": 42,
			"active":  true,
		},
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      s.ctx,
				Timestamp:    s.now,
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
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) TestHandleWithComplexTypes() {
	entry := &TestEntry{
		ctx:   s.ctx,
		level: contractslog.LevelWarning,
		time:  s.now,
		user: map[string]any{
			"id":   1,
			"role": "admin",
		},
		tags: []string{"critical", "auth"},
		request: map[string]any{
			"method": "GET",
			"url":    "/login",
		},
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      s.ctx,
				Timestamp:    s.now,
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
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) TestHandleWithPanicLevel() {
	entry := &TestEntry{
		ctx:   s.ctx,
		level: contractslog.LevelPanic,
		time:  s.now,
	}

	err := s.handler.Handle(entry)
	s.NoError(err)

	result := s.recorder.Result()
	s.normalizeObservedTimestamp(result)

	expected := logtest.Recording{
		logtest.Scope{Name: s.loggerName}: {
			{
				Context:      s.ctx,
				Timestamp:    s.now,
				Severity:     log.SeverityFatal4,
				SeverityText: "panic",
				Body:         log.StringValue(""),
			},
		},
	}

	logtest.AssertEqual(s.T(), expected, result)
}

func (s *HandlerTestSuite) normalizeObservedTimestamp(result logtest.Recording) {
	for scope := range result {
		for i := range result[scope] {
			result[scope][i].ObservedTimestamp = time.Time{}
		}
	}
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
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
