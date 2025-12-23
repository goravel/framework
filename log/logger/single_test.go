package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/file"
)

type SingleTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	json       foundation.Json
}

func TestSingleTestSuite(t *testing.T) {
	suite.Run(t, new(SingleTestSuite))
}

func (s *SingleTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.json = json.New()
}

func (s *SingleTestSuite) TearDownTest() {
	// Cleanup test files
	_ = file.Remove("storage")
}

func (s *SingleTestSuite) TestNewSingle() {
	single := NewSingle(s.mockConfig, s.json)
	s.NotNil(single)
	s.Equal(s.mockConfig, single.config)
	s.Equal(s.json, single.json)
}

func (s *SingleTestSuite) TestHandle_Success() {
	s.mockConfig.EXPECT().GetString("logging.channels.single.path").Return("storage/logs/test.log").Once()
	s.mockConfig.EXPECT().GetString("logging.channels.single.level").Return("debug").Once()

	single := NewSingle(s.mockConfig, s.json)
	handler, err := single.Handle("logging.channels.single")

	s.Nil(err)
	s.NotNil(handler)
}

func (s *SingleTestSuite) TestHandle_EmptyPath() {
	s.mockConfig.EXPECT().GetString("logging.channels.single.path").Return("").Once()

	single := NewSingle(s.mockConfig, s.json)
	handler, err := single.Handle("logging.channels.single")

	s.Nil(handler)
	s.Equal(errors.LogEmptyLogFilePath, err)
}

func (s *SingleTestSuite) TestHandle_DifferentLevels() {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warning level", "warning"},
		{"error level", "error"},
		{"fatal level", "fatal"},
		{"panic level", "panic"},
		{"invalid level defaults to debug", "invalid"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := mocksconfig.NewConfig(s.T())
			mockConfig.EXPECT().GetString("logging.channels.single.path").Return("storage/logs/test.log").Once()
			mockConfig.EXPECT().GetString("logging.channels.single.level").Return(tt.level).Once()

			single := NewSingle(mockConfig, s.json)
			handler, err := single.Handle("logging.channels.single")

			s.Nil(err)
			s.NotNil(handler)
		})
	}
}

func TestGetLevelFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected log.Level
	}{
		{"debug", "debug", log.LevelDebug},
		{"info", "info", log.LevelInfo},
		{"warning", "warning", log.LevelWarning},
		{"warn", "warn", log.LevelWarning},
		{"error", "error", log.LevelError},
		{"fatal", "fatal", log.LevelFatal},
		{"panic", "panic", log.LevelPanic},
		{"invalid defaults to debug", "invalid", log.LevelDebug},
		{"empty defaults to debug", "", log.LevelDebug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevelFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
