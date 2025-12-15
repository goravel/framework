package formatter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
)

type GeneralTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	record     slog.Record
	json       foundation.Json
}

func TestGeneralTestSuite(t *testing.T) {
	suite.Run(t, new(GeneralTestSuite))
}

func (s *GeneralTestSuite) SetupTest() {
	s.mockConfig = configmock.NewConfig(s.T())
	s.record = slog.NewRecord(time.Now().In(time.UTC), slog.LevelInfo, "Test Message", 0)
	s.json = json.New()
}

func (s *GeneralTestSuite) TestFormat() {
	s.mockConfig.EXPECT().GetString("app.env").Return("test").Twice()

	general := NewGeneral(s.mockConfig, s.json)
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Error in Marshaling",
			setup: func() {
				s.record = slog.NewRecord(time.Now().In(time.UTC), slog.LevelInfo, "Test Message", 0)
				s.record.AddAttrs(slog.Any("root", make(chan int)))
			},
			assert: func() {
				formatLog, err := general.Format(context.Background(), s.record)
				s.NotNil(err)
				s.Nil(formatLog)
			},
		},
		{
			name: "Data is not empty",
			setup: func() {
				s.record = slog.NewRecord(time.Now().In(time.UTC), slog.LevelInfo, "Test Message", 0)
				s.record.AddAttrs(slog.Any("root", map[string]any{
					"code":   "200",
					"domain": "example.com",
					"owner":  "owner",
					"user":   "user1",
				}))
			},
			assert: func() {
				formatLog, err := general.Format(context.Background(), s.record)
				s.Nil(err)
				s.Contains(string(formatLog), fmt.Sprintf("[%s] test.info: Test Message", s.record.Time.In(time.UTC).Format("2006-01-02 15:04:05.999")))
				s.Contains(string(formatLog), "[Code] 200")
				s.Contains(string(formatLog), "[Domain] example.com")
				s.Contains(string(formatLog), "[Owner] owner")
				s.Contains(string(formatLog), "[User] user1")
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			test.assert()
		})
	}
}

func (s *GeneralTestSuite) TestFormatData() {
	var data map[string]any
	general := NewGeneral(s.mockConfig, s.json)
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Data is empty",
			setup: func() {
				data = map[string]any{}
			},
			assert: func() {
				formattedData, err := general.formatData(data)
				s.Nil(err)
				s.Empty(formattedData)
			},
		},
		{
			name: "Root key is absent",
			setup: func() {
				data = map[string]any{
					"key": "value",
				}
			},
			assert: func() {
				formattedData, err := general.formatData(data)
				s.NotNil(err)
				s.Empty(formattedData)
			},
		},
		{
			name: "Data is not empty",
			setup: func() {
				data = map[string]any{
					"root": map[string]any{
						"code":   "200",
						"domain": "example.com",
						"owner":  "owner",
						"user":   "user1",
					},
				}
			},
			assert: func() {
				formattedData, err := general.formatData(data)
				s.Nil(err)
				s.Contains(formattedData, "[Code] 200")
				s.Contains(formattedData, "[Domain] example.com")
				s.Contains(formattedData, "[Owner] owner")
				s.Contains(formattedData, "[User] user1")
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			test.assert()
		})
	}
}

func (s *GeneralTestSuite) TestFormatStackTraces() {
	var stackTraces any
	general := NewGeneral(s.mockConfig, s.json)
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "StackTraces is nil",
			setup: func() {
				stackTraces = nil
			},
			assert: func() {
				traces, err := general.formatStackTraces(stackTraces)
				s.Nil(err)
				s.Equal("[Trace]\n", traces)
			},
		},
		{
			name: "StackTraces is not nil",
			setup: func() {
				stackTraces = map[string]any{
					"root": map[string]any{
						"message": "error bad request", // root cause
						"stack": []string{
							"/dummy/examples/logging/example.go:143 [main.main]", // original calling method
							"/dummy/examples/logging/example.go:71 [main.ProcessResource]",
							"/dummy/examples/logging/example.go:29 [main.(*Request).Validate]", // location of Wrap call
							"/dummy/examples/logging/example.go:28 [main.(*Request).Validate]", // location of the root
						},
					},
					"wrap": []map[string]any{
						{
							"message": "received a request with no ID",                                    // additional context
							"stack":   "/dummy/examples/logging/example.go:29 [main.(*Request).Validate]", // location of Wrap call
						},
					},
				}
			},
			assert: func() {
				traces, err := general.formatStackTraces(stackTraces)
				s.Nil(err)
				stackTraces := []string{
					"/dummy/examples/logging/example.go:143 [main.main]",
					"/dummy/examples/logging/example.go:71 [main.ProcessResource]",
					"/dummy/examples/logging/example.go:29 [main.(*Request).Validate]",
					"/dummy/examples/logging/example.go:28 [main.(*Request).Validate]",
				}
				formattedStackTraces := "[Trace]\n" + strings.Join(stackTraces, "\n") + "\n"

				s.Equal(formattedStackTraces, traces)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			test.assert()
		})
	}
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
		{
			name:     "Edge case: Colon at the end",
			input:    "file.go:",
			expected: "file.go:\n",
		},
		{
			name:     "Edge case: Colon at the beginning",
			input:    ":file.go",
			expected: ":file.go\n",
		},
		{
			name:     "Edge case: Multiple colons with no method",
			input:    "/path/to/file.go:100:200",
			expected: "100:200 [/path/to/file.go]\n",
		},
		{
			name:     "Valid stack trace with nested method and line",
			input:    "pkg.subpkg.functionName:/path/to/file.go:55",
			expected: "/path/to/file.go:55 [pkg.subpkg.functionName]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStackTrace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeleteKey(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		key    string
		assert func()
	}{
		{
			name: "Key is not present in data",
			data: map[string]any{
				"key": "value",
			},
			key: "notPresent",
			assert: func() {
				removedData := deleteKey(map[string]any{
					"key": "value",
				}, "notPresent")
				assert.Equal(t, map[string]any{
					"key": "value",
				}, removedData)
			},
		},
		{
			name: "Key is present in data",
			data: map[string]any{
				"key": "value",
			},
			key: "key",
			assert: func() {
				removedData := deleteKey(map[string]any{
					"key": "value",
				}, "key")
				assert.Equal(t, map[string]any{}, removedData)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.assert()
		})
	}
}
