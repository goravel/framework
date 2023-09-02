package formatter

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type GeneralFormatterTestSuite struct {
	suite.Suite
	mockConfig  *configmock.Config
	entry       *logrus.Entry
	stackTraces any
}

func TestGeneralFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(GeneralFormatterTestSuite))
}

func (s *GeneralFormatterTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.entry = &logrus.Entry{
		Level:   logrus.InfoLevel,
		Message: "Test Message",
	}

	s.mockConfig.On("GetString", "app.timezone").Return("UTC")
	s.mockConfig.On("GetString", "app.env").Return("test")
}

func (s *GeneralFormatterTestSuite) TestGeneral_Format() {
	general := NewGeneral(s.mockConfig)
	formatLog, err := general.Format(s.entry)
	s.Nil(err)
	s.NotNil(formatLog)
}

func (s *GeneralFormatterTestSuite) TestFormatData() {
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Data is empty",
			setup: func() {
				s.entry.Data = logrus.Fields{}
			},
			assert: func() {
				formattedData, err := formatData(s.entry.Data)
				s.Nil(err)
				s.Empty(formattedData)
			},
		},
		{
			name: "Root key is absent",
			setup: func() {
				s.entry.Data = logrus.Fields{
					"key": "value",
				}
			},
			assert: func() {
				formattedData, err := formatData(s.entry.Data)
				s.NotNil(err)
				s.Empty(formattedData)
			},
		},
		{
			name: "Invalid data type",
			setup: func() {
				s.entry.Data = logrus.Fields{
					"root": map[string]interface{}{
						"code":     "123",
						"context":  "sample",
						"domain":   "example.com",
						"hint":     make(chan int), // Invalid data type that will cause an error during value extraction
						"owner":    "owner",
						"request":  map[string]interface{}{"method": "GET", "uri": "http://localhost"},
						"response": map[string]interface{}{"status": 200},
						"tags":     []string{"tag1", "tag2"},
						"user":     "user1",
					},
				}
			},
			assert: func() {
				formattedData, err := formatData(s.entry.Data)
				s.NotNil(err)
				s.Empty(formattedData)
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

func (s *GeneralFormatterTestSuite) TestDeleteKey() {
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
				s.Equal(logrus.Fields{
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
				s.Equal(logrus.Fields{}, removedData)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.assert()
		})
	}
}

func (s *GeneralFormatterTestSuite) TestFormatStackTraces() {
	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "StackTraces is nil",
			setup: func() {
				s.stackTraces = nil
			},
			assert: func() {
				traces, err := formatStackTraces(s.stackTraces)
				s.Nil(err)
				s.Equal("trace:\n", traces)
			},
		},
		{
			name: "StackTraces is not nil",
			setup: func() {
				s.stackTraces = map[string]interface{}{
					"root": map[string]interface{}{
						"message": "error bad request", // root cause
						"stack": []string{
							"main.main:/dummy/examples/logging/example.go:143", // original calling method
							"main.ProcessResource:/dummy/examples/logging/example.go:71",
							"main.(*Request).Validate:/dummy/examples/logging/example.go:29", // location of Wrap call
							"main.(*Request).Validate:/dummy/examples/logging/example.go:28", // location of the root
						},
					},
					"wrap": []map[string]interface{}{
						{
							"message": "received a request with no ID",                                  // additional context
							"stack":   "main.(*Request).Validate:/dummy/examples/logging/example.go:29", // location of Wrap call
						},
					},
				}
			},
			assert: func() {
				traces, err := formatStackTraces(s.stackTraces)
				s.Nil(err)
				stackTraces := []string{
					"main.main:/dummy/examples/logging/example.go:143",
					"main.ProcessResource:/dummy/examples/logging/example.go:71",
					"main.(*Request).Validate:/dummy/examples/logging/example.go:29",
					"main.(*Request).Validate:/dummy/examples/logging/example.go:28",
				}
				formattedStackTraces := "trace:\n\t" + strings.Join(stackTraces, "\n\t") + "\n"

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
