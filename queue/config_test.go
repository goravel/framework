package queue

import (
	"testing"

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
