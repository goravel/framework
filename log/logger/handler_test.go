package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

type FileHandlerTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	json       foundation.Json
	buffer     *bytes.Buffer
}

func TestFileHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FileHandlerTestSuite))
}

func (s *FileHandlerTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.json = json.New()
	s.buffer = new(bytes.Buffer)
}

func (s *FileHandlerTestSuite) TestNewFileHandler() {
	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)
	s.NotNil(handler)
	s.Equal(s.buffer, handler.writer)
	s.Equal(s.mockConfig, handler.config)
	s.Equal(s.json, handler.json)
	s.Equal(log.LevelDebug, handler.level)
}

func (s *FileHandlerTestSuite) TestEnabled() {
	tests := []struct {
		name          string
		handlerLevel  log.Level
		recordLevel   slog.Level
		expectEnabled bool
	}{
		{
			name:          "debug handler allows debug",
			handlerLevel:  log.LevelDebug,
			recordLevel:   slog.LevelDebug,
			expectEnabled: true,
		},
		{
			name:          "debug handler allows info",
			handlerLevel:  log.LevelDebug,
			recordLevel:   slog.LevelInfo,
			expectEnabled: true,
		},
		{
			name:          "info handler blocks debug",
			handlerLevel:  log.LevelInfo,
			recordLevel:   slog.LevelDebug,
			expectEnabled: false,
		},
		{
			name:          "error handler blocks warning",
			handlerLevel:  log.LevelError,
			recordLevel:   slog.LevelWarn,
			expectEnabled: false,
		},
		{
			name:          "error handler allows error",
			handlerLevel:  log.LevelError,
			recordLevel:   slog.LevelError,
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := NewFileHandler(s.buffer, s.mockConfig, s.json, tt.handlerLevel)
			result := handler.Enabled(log.Level(tt.recordLevel))
			s.Equal(tt.expectEnabled, result)
		})
	}
}

func (s *FileHandlerTestSuite) TestHandle() {
	s.mockConfig.EXPECT().GetString("app.env").Return("test").Maybe()

	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)

	entry := &mockEntry{
		time:    time.Now(),
		level:   log.LevelInfo,
		message: "test message",
	}

	err := handler.Handle(entry)
	s.Nil(err)
	s.Contains(s.buffer.String(), "test.info: test message")
}

func TestFormatStackTrace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid stack trace with file and method",
			input:    "main.functionName:/path/to/file.go:42",
			expected: "/path/to/file.go:42 [main.functionName]\n",
		},
		{
			name:     "Valid stack trace without method",
			input:    "/path/to/file.go:42",
			expected: "/path/to/file.go:42\n",
		},
		{
			name:     "No colons in stack trace",
			input:    "invalidstacktrace",
			expected: "invalidstacktrace\n",
		},
		{
			name:     "Single colon in stack trace",
			input:    "file.go:42",
			expected: "file.go:42\n",
		},
		{
			name:     "Edge case: Empty string",
			input:    "",
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStackTrace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    log.Level
		expected string
	}{
		{log.LevelDebug, "debug"},
		{log.LevelInfo, "info"},
		{log.LevelWarning, "warning"},
		{log.LevelError, "error"},
		{log.LevelFatal, "fatal"},
		{log.LevelPanic, "panic"},
		{log.Level(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

type ConsoleHandlerTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	json       foundation.Json
}

func TestConsoleHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleHandlerTestSuite))
}

func (s *ConsoleHandlerTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.json = json.New()
}

func (s *ConsoleHandlerTestSuite) TestNewConsoleHandler() {
	handler := NewConsoleHandler(s.mockConfig, s.json)
	s.NotNil(handler)
	s.NotNil(handler.FileHandler)
	s.Equal(log.LevelDebug, handler.level)
}

type mockEntry struct {
	time       time.Time
	ctx        context.Context
	owner      any
	user       any
	data       log.Data
	request    map[string]any
	response   map[string]any
	stacktrace map[string]any
	with       map[string]any
	code       string
	domain     string
	hint       string
	message    string
	tags       []string
	level      log.Level
}

func (e *mockEntry) Code() string {
	return e.code
}

func (e *mockEntry) Context() context.Context {
	return e.ctx
}

func (e *mockEntry) Data() log.Data {
	return e.data
}

func (e *mockEntry) Domain() string {
	return e.domain
}

func (e *mockEntry) Hint() string {
	return e.hint
}

func (e *mockEntry) Level() log.Level {
	return e.level
}

func (e *mockEntry) Message() string {
	return e.message
}

func (e *mockEntry) Owner() any {
	return e.owner
}

func (e *mockEntry) Request() map[string]any {
	return e.request
}

func (e *mockEntry) Response() map[string]any {
	return e.response
}

func (e *mockEntry) Tags() []string {
	return e.tags
}

func (e *mockEntry) Time() time.Time {
	return e.time
}

func (e *mockEntry) Trace() map[string]any {
	return e.stacktrace
}

func (e *mockEntry) User() any {
	return e.user
}

func (e *mockEntry) With() map[string]any {
	return e.with
}
