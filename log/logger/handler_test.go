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
			result := handler.Enabled(context.Background(), tt.recordLevel)
			s.Equal(tt.expectEnabled, result)
		})
	}
}

func (s *FileHandlerTestSuite) TestHandle() {
	s.mockConfig.EXPECT().GetString("app.env").Return("test").Maybe()

	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)

	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "test message",
	}

	err := handler.Handle(context.Background(), record)
	s.Nil(err)
	s.Contains(s.buffer.String(), "test.info: test message")
}

func (s *FileHandlerTestSuite) TestHandleWithRootAttrs() {
	s.mockConfig.EXPECT().GetString("app.env").Return("test").Maybe()

	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)

	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "test message",
	}
	record.AddAttrs(slog.Group("root",
		slog.String("code", "test_code"),
		slog.String("hint", "test_hint"),
	))

	err := handler.Handle(context.Background(), record)
	s.Nil(err)
	output := s.buffer.String()
	s.Contains(output, "test.info: test message")
	s.Contains(output, "[Code] test_code")
	s.Contains(output, "[Hint] test_hint")
}

func (s *FileHandlerTestSuite) TestWithAttrs() {
	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)
	newHandler := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})

	s.NotNil(newHandler)
	s.IsType(&FileHandler{}, newHandler)
	fileHandler := newHandler.(*FileHandler)
	s.Len(fileHandler.attrs, 1)
	s.Equal("key", fileHandler.attrs[0].Key)
}

func (s *FileHandlerTestSuite) TestWithGroup() {
	handler := NewFileHandler(s.buffer, s.mockConfig, s.json, log.LevelDebug)
	newHandler := handler.WithGroup("testgroup")

	s.NotNil(newHandler)
	s.IsType(&FileHandler{}, newHandler)
	fileHandler := newHandler.(*FileHandler)
	s.Len(fileHandler.groups, 1)
	s.Equal("testgroup", fileHandler.groups[0])
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
			result := levelToString(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractValue(t *testing.T) {
	tests := []struct {
		name     string
		value    slog.Value
		expected any
	}{
		{
			name:     "string value",
			value:    slog.StringValue("test"),
			expected: "test",
		},
		{
			name:     "int64 value",
			value:    slog.Int64Value(42),
			expected: int64(42),
		},
		{
			name:     "float64 value",
			value:    slog.Float64Value(3.14),
			expected: 3.14,
		},
		{
			name:     "bool value",
			value:    slog.BoolValue(true),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractGroupData(t *testing.T) {
	// Test with group value
	groupValue := slog.GroupValue(
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	)
	result := extractGroupData(groupValue)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, int64(42), result["key2"])

	// Test with any value containing map
	mapValue := slog.AnyValue(map[string]any{
		"key1": "value1",
		"key2": 42,
	})
	result = extractGroupData(mapValue)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, 42, result["key2"])
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
