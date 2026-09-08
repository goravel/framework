package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

type ConfigTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	config     *Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	s.mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	s.config = NewConfig(s.mockConfig)
}

func (s *ConfigTestSuite) TestDebug() {
	s.True(s.config.Debug())
}

func (s *ConfigTestSuite) TestDefaultConnection() {
	s.Equal("redis", s.config.DefaultConnection())
}

func (s *ConfigTestSuite) TestDefaultQueue() {
	s.Equal("default", s.config.DefaultQueue())
}

func (s *ConfigTestSuite) TestDefaultConcurrent() {
	s.Equal(2, s.config.DefaultConcurrent())
}

func (s *ConfigTestSuite) TestDriver() {
	s.mockConfig.EXPECT().GetString("queue.connections.sync.driver").Return("sync").Once()
	s.Equal("sync", s.config.Driver("sync"))
}

func (s *ConfigTestSuite) TestFailedDatabase() {
	s.Equal("mysql", s.config.FailedDatabase())
}

func (s *ConfigTestSuite) TestFailedTable() {
	s.Equal("failed_jobs", s.config.FailedTable())
}

func (s *ConfigTestSuite) TestVia() {
	s.mockConfig.EXPECT().Get("queue.connections.sync.via").Return("sync").Once()
	s.Equal("sync", s.config.Via("sync"))
}

func (s *ConfigTestSuite) TestTimeout() {
	tests := []struct {
		name     string
		value    any
		expected time.Duration
	}{
		{"int seconds", 10, 10 * time.Second},
		{"int64 seconds", int64(8), 8 * time.Second},
		{"float64 seconds", float64(7), 7 * time.Second},
		{"duration string", "5s", 5 * time.Second},
		{"duration string with ms", "1500ms", 1500 * time.Millisecond},
		{"zero falls back to default", 0, defaultReceiveTimeout * time.Second},
		{"negative falls back to default", -3, defaultReceiveTimeout * time.Second},
		{"negative duration string falls back to default", "-5s", defaultReceiveTimeout * time.Second},
		{"garbage string falls back to default", "not-a-duration", defaultReceiveTimeout * time.Second},
		{"missing falls back to default", nil, defaultReceiveTimeout * time.Second},
		{"unsupported type falls back to default", true, defaultReceiveTimeout * time.Second},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.mockConfig.EXPECT().Get("queue.connections.redis.timeout").Return(test.value).Once()
			s.Equal(test.expected, s.config.Timeout("redis"))
		})
	}
}
