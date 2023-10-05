package formatter

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type GeneralTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	entry      *logrus.Entry
}

func TestGeneralTestSuite(t *testing.T) {
	suite.Run(t, new(GeneralTestSuite))
}

func (s *GeneralTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.entry = &logrus.Entry{
		Level:   logrus.InfoLevel,
		Message: "Test Message",
	}
}

func (s *GeneralTestSuite) TestFormat() {
	s.mockConfig.On("GetString", "app.timezone").Return("UTC")
	s.mockConfig.On("GetString", "app.env").Return("test")

	general := NewGeneral(s.mockConfig)
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Error in Marshaling",
			setup: func() {
				s.entry.Data = logrus.Fields{
					"root": make(chan int),
				}
			},
			assert: func() {
				formatLog, err := general.Format(s.entry)
				s.NotNil(err)
				s.Nil(formatLog)
			},
		},
		{
			name: "Data is not empty",
			setup: func() {
				s.entry.Data = logrus.Fields{
					"root": map[string]any{
						"code":   "200",
						"domain": "example.com",
						"owner":  "owner",
						"user":   "user1",
					},
				}
			},
			assert: func() {
				formatLog, err := general.Format(s.entry)
				s.Nil(err)
				s.Contains(string(formatLog), "code: \"200\"")
				s.Contains(string(formatLog), "domain: \"example.com\"")
				s.Contains(string(formatLog), "owner: \"owner\"")
				s.Contains(string(formatLog), "user: \"user1\"")
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			test.assert()
		})
		s.mockConfig.AssertExpectations(s.T())
	}
}

func TestFormatData(t *testing.T) {
	var data logrus.Fields
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Data is empty",
			setup: func() {
				data = logrus.Fields{}
			},
			assert: func() {
				formattedData, err := formatData(data)
				assert.Nil(t, err)
				assert.Empty(t, formattedData)
			},
		},
		{
			name: "Root key is absent",
			setup: func() {
				data = logrus.Fields{
					"key": "value",
				}
			},
			assert: func() {
				formattedData, err := formatData(data)
				assert.NotNil(t, err)
				assert.Empty(t, formattedData)
			},
		},
		{
			name: "Invalid data type",
			setup: func() {
				data = logrus.Fields{
					"root": map[string]any{
						"code":     "123",
						"context":  "sample",
						"domain":   "example.com",
						"hint":     make(chan int), // Invalid data type that will cause an error during value extraction
						"owner":    "owner",
						"request":  map[string]any{"method": "GET", "uri": "http://localhost"},
						"response": map[string]any{"status": 200},
						"tags":     []string{"tag1", "tag2"},
						"user":     "user1",
					},
				}
			},
			assert: func() {
				formattedData, err := formatData(data)
				assert.NotNil(t, err)
				assert.Empty(t, formattedData)
			},
		},
		{
			name: "Data is not empty",
			setup: func() {
				data = logrus.Fields{
					"root": map[string]any{
						"code":   "200",
						"domain": "example.com",
						"owner":  "owner",
						"user":   "user1",
					},
				}
			},
			assert: func() {
				formattedData, err := formatData(data)
				assert.Nil(t, err)
				assert.Contains(t, formattedData, "code: \"200\"")
				assert.Contains(t, formattedData, "domain: \"example.com\"")
				assert.Contains(t, formattedData, "owner: \"owner\"")
				assert.Contains(t, formattedData, "user: \"user1\"")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			test.assert()
		})
	}
}

func TestDeleteKey(t *testing.T) {
	tests := []struct {
		name   string
		data   logrus.Fields
		key    string
		assert func()
	}{
		{
			name: "Key is not present in data",
			data: logrus.Fields{
				"key": "value",
			},
			key: "notPresent",
			assert: func() {
				removedData := deleteKey(logrus.Fields{
					"key": "value",
				}, "notPresent")
				assert.Equal(t, logrus.Fields{
					"key": "value",
				}, removedData)
			},
		},
		{
			name: "Key is present in data",
			data: logrus.Fields{
				"key": "value",
			},
			key: "key",
			assert: func() {
				removedData := deleteKey(logrus.Fields{
					"key": "value",
				}, "key")
				assert.Equal(t, logrus.Fields{}, removedData)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.assert()
		})
	}
}

func TestFormatStackTraces(t *testing.T) {
	var stackTraces any
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
				traces, err := formatStackTraces(stackTraces)
				assert.Nil(t, err)
				assert.Equal(t, "trace:\n", traces)
			},
		},
		{
			name: "StackTraces is not nil",
			setup: func() {
				stackTraces = map[string]any{
					"root": map[string]any{
						"message": "error bad request", // root cause
						"stack": []string{
							"main.main:/dummy/examples/logging/example.go:143", // original calling method
							"main.ProcessResource:/dummy/examples/logging/example.go:71",
							"main.(*Request).Validate:/dummy/examples/logging/example.go:29", // location of Wrap call
							"main.(*Request).Validate:/dummy/examples/logging/example.go:28", // location of the root
						},
					},
					"wrap": []map[string]any{
						{
							"message": "received a request with no ID",                                  // additional context
							"stack":   "main.(*Request).Validate:/dummy/examples/logging/example.go:29", // location of Wrap call
						},
					},
				}
			},
			assert: func() {
				traces, err := formatStackTraces(stackTraces)
				assert.Nil(t, err)
				stackTraces := []string{
					"main.main:/dummy/examples/logging/example.go:143",
					"main.ProcessResource:/dummy/examples/logging/example.go:71",
					"main.(*Request).Validate:/dummy/examples/logging/example.go:29",
					"main.(*Request).Validate:/dummy/examples/logging/example.go:28",
				}
				formattedStackTraces := "trace:\n\t" + strings.Join(stackTraces, "\n\t") + "\n"

				assert.Equal(t, formattedStackTraces, traces)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			test.assert()
		})
	}
}
