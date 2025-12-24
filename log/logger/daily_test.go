package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/file"
)

type DailyTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	json       foundation.Json
}

func TestDailyTestSuite(t *testing.T) {
	suite.Run(t, new(DailyTestSuite))
}

func (s *DailyTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.json = json.New()
}

func (s *DailyTestSuite) TearDownTest() {
	// Cleanup test files
	_ = file.Remove("storage")
}

func (s *DailyTestSuite) TestNewDaily() {
	daily := NewDaily(s.mockConfig, s.json)
	s.NotNil(daily)
	s.Equal(s.mockConfig, daily.config)
	s.Equal(s.json, daily.json)
}

func (s *DailyTestSuite) TestHandle_Success() {
	s.mockConfig.EXPECT().GetString("logging.channels.daily.path").Return("storage/logs/daily.log").Once()
	s.mockConfig.EXPECT().GetString("logging.channels.daily.level").Return("debug").Once()
	s.mockConfig.EXPECT().GetString("logging.channels.daily.formatter").Return("").Once()
	s.mockConfig.EXPECT().GetInt("logging.channels.daily.days").Return(7).Once()

	daily := NewDaily(s.mockConfig, s.json)
	handler, err := daily.Handle("logging.channels.daily")

	s.Nil(err)
	s.NotNil(handler)
}

func (s *DailyTestSuite) TestHandle_EmptyPath() {
	s.mockConfig.EXPECT().GetString("logging.channels.daily.path").Return("").Once()

	daily := NewDaily(s.mockConfig, s.json)
	handler, err := daily.Handle("logging.channels.daily")

	s.Nil(handler)
	s.Equal(errors.LogEmptyLogFilePath, err)
}

func (s *DailyTestSuite) TestHandle_DifferentLevels() {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warning level", "warning"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := mocksconfig.NewConfig(s.T())
			mockConfig.EXPECT().GetString("logging.channels.daily.path").Return("storage/logs/daily.log").Once()
			mockConfig.EXPECT().GetString("logging.channels.daily.level").Return(tt.level).Once()
			mockConfig.EXPECT().GetString("logging.channels.daily.formatter").Return("").Once()
			mockConfig.EXPECT().GetInt("logging.channels.daily.days").Return(7).Once()

			daily := NewDaily(mockConfig, s.json)
			handler, err := daily.Handle("logging.channels.daily")

			s.Nil(err)
			s.NotNil(handler)
		})
	}
}

func (s *DailyTestSuite) TestHandle_DifferentDays() {
	tests := []struct {
		name string
		days int
	}{
		{"1 day", 1},
		{"7 days", 7},
		{"30 days", 30},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := mocksconfig.NewConfig(s.T())
			mockConfig.EXPECT().GetString("logging.channels.daily.path").Return("storage/logs/daily.log").Once()
			mockConfig.EXPECT().GetString("logging.channels.daily.level").Return("debug").Once()
			mockConfig.EXPECT().GetString("logging.channels.daily.formatter").Return("").Once()
			mockConfig.EXPECT().GetInt("logging.channels.daily.days").Return(tt.days).Once()

			daily := NewDaily(mockConfig, s.json)
			handler, err := daily.Handle("logging.channels.daily")

			s.Nil(err)
			s.NotNil(handler)
		})
	}
}

func TestNewRotatingFileHandler(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	j := json.New()
	buffer := new(bytes.Buffer)

	handler := NewRotatingFileHandler(buffer, mockConfig, j, nil)

	if handler == nil {
		t.Error("Expected handler to be not nil")
	}
}
