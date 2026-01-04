package log

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/log"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

// mockHandler is a test implementation of log.Handler
type mockHandler struct {
	enabledLevel log.Level
	handleErr    error
	handleCalled bool
	lastEntry    log.Entry
}

func (m *mockHandler) Enabled(level log.Level) bool {
	return level >= m.enabledLevel
}

func (m *mockHandler) Handle(entry log.Entry) error {
	m.handleCalled = true
	m.lastEntry = entry
	return m.handleErr
}

func (s *UtilsTestSuite) TestSlogAdapterEnabled() {
	tests := []struct {
		name         string
		enabledLevel log.Level
		testLevel    slog.Level
		expected     bool
	}{
		{
			name:         "debug level enabled for debug",
			enabledLevel: log.LevelDebug,
			testLevel:    slog.LevelDebug,
			expected:     true,
		},
		{
			name:         "debug level enabled for info",
			enabledLevel: log.LevelDebug,
			testLevel:    slog.LevelInfo,
			expected:     true,
		},
		{
			name:         "info level not enabled for debug",
			enabledLevel: log.LevelInfo,
			testLevel:    slog.LevelDebug,
			expected:     false,
		},
		{
			name:         "error level enabled for error",
			enabledLevel: log.LevelError,
			testLevel:    slog.LevelError,
			expected:     true,
		},
		{
			name:         "error level not enabled for warn",
			enabledLevel: log.LevelError,
			testLevel:    slog.LevelWarn,
			expected:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := &mockHandler{enabledLevel: tt.enabledLevel}
			adapter := &slogAdapter{handler: handler}

			result := adapter.Enabled(context.Background(), tt.testLevel)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *UtilsTestSuite) TestSlogAdapterHandle() {
	handler := &mockHandler{enabledLevel: log.LevelDebug}
	adapter := &slogAdapter{handler: handler}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	err := adapter.Handle(context.Background(), record)

	s.Nil(err)
	s.True(handler.handleCalled)
	s.NotNil(handler.lastEntry)
	s.Equal("test message", handler.lastEntry.Message())
}

func (s *UtilsTestSuite) TestSlogAdapterHandleError() {
	expectedErr := errors.New("handle error")
	handler := &mockHandler{
		enabledLevel: log.LevelDebug,
		handleErr:    expectedErr,
	}
	adapter := &slogAdapter{handler: handler}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	err := adapter.Handle(context.Background(), record)

	s.Equal(expectedErr, err)
	s.True(handler.handleCalled)
}

func (s *UtilsTestSuite) TestSlogAdapterWithAttrs() {
	handler := &mockHandler{enabledLevel: log.LevelDebug}
	adapter := &slogAdapter{handler: handler}

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := adapter.WithAttrs(attrs)

	// The adapter should return itself (no attribute handling in current implementation)
	s.Equal(adapter, newHandler)
}

func (s *UtilsTestSuite) TestSlogAdapterWithGroup() {
	handler := &mockHandler{enabledLevel: log.LevelDebug}
	adapter := &slogAdapter{handler: handler}

	newHandler := adapter.WithGroup("testgroup")

	// The adapter should return itself (no group handling in current implementation)
	s.Equal(adapter, newHandler)
}

func (s *UtilsTestSuite) TestHandlerToSlogHandler() {
	handler := &mockHandler{enabledLevel: log.LevelDebug}

	slogHandler := HandlerToSlogHandler(handler)

	s.NotNil(slogHandler)

	// Verify it's our adapter
	adapter, ok := slogHandler.(*slogAdapter)
	s.True(ok)
	s.Equal(handler, adapter.handler)
}

func (s *UtilsTestSuite) TestHandlerToSlogHandlerIntegration() {
	handler := &mockHandler{enabledLevel: log.LevelInfo}

	slogHandler := HandlerToSlogHandler(handler)

	// Test Enabled
	s.False(slogHandler.Enabled(context.Background(), slog.LevelDebug))
	s.True(slogHandler.Enabled(context.Background(), slog.LevelInfo))
	s.True(slogHandler.Enabled(context.Background(), slog.LevelError))

	// Test Handle
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "integration test", 0)
	record.Add("code", "TEST001")

	err := slogHandler.Handle(context.Background(), record)
	s.Nil(err)
	s.True(handler.handleCalled)
	s.Equal("integration test", handler.lastEntry.Message())
	s.Equal("TEST001", handler.lastEntry.Code())
}

// mockHook is a test implementation of log.Hook
type mockHook struct {
	levels     []log.Level
	fireCalled bool
	lastEntry  log.Entry
	fireErr    error
}

func (m *mockHook) Levels() []log.Level {
	return m.levels
}

func (m *mockHook) Fire(entry log.Entry) error {
	m.fireCalled = true
	m.lastEntry = entry
	return m.fireErr
}

func (s *UtilsTestSuite) TestHookAdapterEnabled() {
	tests := []struct {
		name      string
		levels    []log.Level
		testLevel log.Level
		expected  bool
	}{
		{
			name:      "enabled for matching level",
			levels:    []log.Level{log.LevelDebug, log.LevelInfo, log.LevelError},
			testLevel: log.LevelInfo,
			expected:  true,
		},
		{
			name:      "not enabled for non-matching level",
			levels:    []log.Level{log.LevelError, log.LevelFatal},
			testLevel: log.LevelInfo,
			expected:  false,
		},
		{
			name:      "enabled for first level",
			levels:    []log.Level{log.LevelDebug, log.LevelInfo},
			testLevel: log.LevelDebug,
			expected:  true,
		},
		{
			name:      "enabled for last level",
			levels:    []log.Level{log.LevelWarning, log.LevelError},
			testLevel: log.LevelError,
			expected:  true,
		},
		{
			name:      "not enabled for empty levels",
			levels:    []log.Level{},
			testLevel: log.LevelInfo,
			expected:  false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			hook := &mockHook{levels: tt.levels}
			adapter := &hookAdapter{hook: hook}

			result := adapter.Enabled(tt.testLevel)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *UtilsTestSuite) TestHookAdapterHandle() {
	hook := &mockHook{levels: []log.Level{log.LevelInfo}}
	adapter := &hookAdapter{hook: hook}

	entry := &mockEntry{
		level:   log.LevelInfo,
		message: "test message",
	}

	err := adapter.Handle(entry)

	s.Nil(err)
	s.True(hook.fireCalled)
	s.Equal(entry, hook.lastEntry)
}

func (s *UtilsTestSuite) TestHookAdapterHandleError() {
	expectedErr := errors.New("fire error")
	hook := &mockHook{
		levels:  []log.Level{log.LevelInfo},
		fireErr: expectedErr,
	}
	adapter := &hookAdapter{hook: hook}

	entry := &mockEntry{
		level:   log.LevelInfo,
		message: "test message",
	}

	err := adapter.Handle(entry)

	s.Equal(expectedErr, err)
	s.True(hook.fireCalled)
}

func (s *UtilsTestSuite) TestHookToHandler() {
	hook := &mockHook{levels: []log.Level{log.LevelInfo, log.LevelError}}

	handler := HookToHandler(hook)

	s.NotNil(handler)

	// Verify it's our adapter
	adapter, ok := handler.(*hookAdapter)
	s.True(ok)
	s.Equal(hook, adapter.hook)

	// Test Enabled
	s.False(handler.Enabled(log.LevelDebug))
	s.True(handler.Enabled(log.LevelInfo))
	s.True(handler.Enabled(log.LevelError))
	s.False(handler.Enabled(log.LevelFatal))
}

// mockEntry is a test implementation of log.Entry
type mockEntry struct {
	level    log.Level
	message  string
	code     string
	ctx      context.Context
	domain   string
	hint     string
	owner    any
	request  map[string]any
	response map[string]any
	tags     []string
	user     any
	with     map[string]any
	trace    map[string]any
	data     log.Data
	t        time.Time
}

func (m *mockEntry) Code() string             { return m.code }
func (m *mockEntry) Context() context.Context { return m.ctx }
func (m *mockEntry) Data() log.Data           { return m.data }
func (m *mockEntry) Domain() string           { return m.domain }
func (m *mockEntry) Hint() string             { return m.hint }
func (m *mockEntry) Level() log.Level         { return m.level }
func (m *mockEntry) Message() string          { return m.message }
func (m *mockEntry) Owner() any               { return m.owner }
func (m *mockEntry) Request() map[string]any  { return m.request }
func (m *mockEntry) Response() map[string]any { return m.response }
func (m *mockEntry) Tags() []string           { return m.tags }
func (m *mockEntry) Time() time.Time          { return m.t }
func (m *mockEntry) Trace() map[string]any    { return m.trace }
func (m *mockEntry) User() any                { return m.user }
func (m *mockEntry) With() map[string]any     { return m.with }
