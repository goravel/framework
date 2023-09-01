package formatter

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type GeneralFormatterTestSuite struct {
	suite.Suite
}

func TestGeneralFormatterTestSuite(t *testing.T) {
	suite.Run(t, new(GeneralFormatterTestSuite))
}

func (s *GeneralFormatterTestSuite) SetupTest() {

}

func (s *GeneralFormatterTestSuite) TestGeneralFormatter() {
	var entry *logrus.Entry
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.timezone").Return("UTC")
	mockConfig.On("GetString", "app.env").Return("test")
	general := NewGeneral(mockConfig)
	beforeEach := func() {
		entry = initMockEntry()
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Error in Marshaling",
			setup: func() {
				entry.Data = logrus.Fields{
					"root": make(chan int),
				}
			},
			assert: func() {
				formatLog, err := general.Format(entry)
				s.NotNil(err)
				s.Nil(formatLog)
			},
		},
		{
			name: "Root key is absent",
			setup: func() {
				entry.Data = logrus.Fields{
					"key": "value",
				}
			},
			assert: func() {
				formatLog, err := general.Format(entry)
				s.NotNil(err)
				s.Nil(formatLog)
			},
		},
		{
			name: "",
			setup: func() {
				entry.Data = logrus.Fields{
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
				formatLog, err := general.Format(entry)
				s.NotNil(err)
				s.Nil(formatLog)
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()
			test.assert()
		})
	}
}

func initMockEntry() *logrus.Entry {
	return &logrus.Entry{
		Level:   logrus.InfoLevel,
		Message: "Test Message",
	}
}
