package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdb "github.com/goravel/framework/mocks/database/db"
)

type ConfigTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	mockDB     *mocksdb.DB
	config     *Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockDB = mocksdb.NewDB(s.T())
	s.config = NewConfig(s.mockConfig, s.mockDB)
}

func (s *ConfigTestSuite) TestDebug() {
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.True(s.config.Debug())

	s.mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
	s.False(s.config.Debug())
}

func (s *ConfigTestSuite) TestDefaultConnection() {
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.Equal("redis", s.config.DefaultConnection())
}

func (s *ConfigTestSuite) TestDriver() {
	// Test with empty connection (should use default)
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.driver").Return("redis").Once()
	s.Equal("redis", s.config.Driver(""))

	// Test with specific connection
	s.mockConfig.EXPECT().GetString("queue.connections.sync.driver").Return("sync").Once()
	s.Equal("sync", s.config.Driver("sync"))
}

func (s *ConfigTestSuite) TestFailedJobsQuery() {
	mockQuery := mocksdb.NewQuery(s.T())

	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()
	s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
	s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()

	result := s.config.FailedJobsQuery()
	s.Equal(mockQuery, result)
}

func (s *ConfigTestSuite) TestQueue() {
	// Test with default app name
	s.mockConfig.EXPECT().GetString("app.name").Return("").Once()
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.Equal("goravel_queues:default", s.config.Queue("", ""))

	// Test with custom app name
	s.mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.Equal("myapp_queues:default", s.config.Queue("redis", ""))

	// Test with custom queue
	s.mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	s.Equal("myapp_queues:custom", s.config.Queue("redis", "custom"))
}

func (s *ConfigTestSuite) TestSize() {
	// Test with empty connection (should use default)
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.size", 100).Return(200).Once()
	s.Equal(200, s.config.Size(""))

	// Test with specific connection
	s.mockConfig.EXPECT().GetInt("queue.connections.sync.size", 100).Return(50).Once()
	s.Equal(50, s.config.Size("sync"))
}

func (s *ConfigTestSuite) TestVia() {
	// Test with empty connection (should use default)
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().Get("queue.connections.redis.via").Return("redis").Once()
	s.Equal("redis", s.config.Via(""))

	// Test with specific connection
	s.mockConfig.EXPECT().Get("queue.connections.sync.via").Return("sync").Once()
	s.Equal("sync", s.config.Via("sync"))
}
