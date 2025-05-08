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

func (s *ConfigTestSuite) TestDefault() {
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()

	connection, queue, concurrent := s.config.Default()
	s.Equal("redis", connection)
	s.Equal("default", queue)
	s.Equal(2, concurrent)
}

func (s *ConfigTestSuite) TestDriver() {
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

func (s *ConfigTestSuite) TestQueueKey() {
	s.mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	s.Equal("myapp_queues:redis_custom", s.config.QueueKey("redis", "custom"))
}

func (s *ConfigTestSuite) TestVia() {
	s.mockConfig.EXPECT().Get("queue.connections.sync.via").Return("sync").Once()
	s.Equal("sync", s.config.Via("sync"))
}
