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

	// Verify it implements slog.Handler
	_, ok := slogHandler.(slog.Handler)
	s.True(ok)

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
