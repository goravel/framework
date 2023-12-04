package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/mocks/config"
	ormmock "github.com/goravel/framework/mocks/database/orm"
)

type ConfigTestSuite struct {
	suite.Suite
	config     *Config
	mockConfig *configmock.Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.config = NewConfig(s.mockConfig)
}

func (s *ConfigTestSuite) TestQueue() {
	tests := []struct {
		name            string
		setup           func()
		connection      string
		queue           string
		expectQueueName string
	}{
		{
			name: "success when connection and queue are empty",
			setup: func() {
				s.mockConfig.On("GetString", "app.name").Return("").Once()
				s.mockConfig.On("GetString", "queue.default").Return("async").Once()
				s.mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("queue").Once()
			},
			expectQueueName: "goravel_queues:queue",
		},
		{
			name: "success when connection and queue aren't empty",
			setup: func() {
				s.mockConfig.On("GetString", "app.name").Return("app").Once()
			},
			connection:      "redis",
			queue:           "queue",
			expectQueueName: "app_queues:queue",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			queueName := s.config.Queue(test.connection, test.queue)
			s.Equal(test.expectQueueName, queueName)
			s.mockConfig.AssertExpectations(s.T())
		})
	}
}

func (s *ConfigTestSuite) TestRedis() {
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Once()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("127.0.0.1").Once()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(6379).Once()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()

	redisClient := s.config.Redis("redis")

	s.NotNil(redisClient)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ConfigTestSuite) TestDatabase() {
	mockOrm := &ormmock.Orm{}
	mockQuery := &ormmock.Query{}
	mockOrm.On("Connection", "database").Return(mockOrm)
	mockOrm.On("Query").Return(mockQuery)
	mockQuery.On("Table", "jobs").Return(mockQuery)

	OrmFacade = mockOrm

	s.mockConfig.On("GetString", "queue.connections.database.connection").Return("database").Once()
	s.mockConfig.On("GetString", "queue.connections.database.table").Return("jobs").Once()

	orm := s.config.Database("database")

	s.NotNil(orm)
	s.mockConfig.AssertExpectations(s.T())
}

func (s *ConfigTestSuite) TestFailedJobsDatabase() {
	mockOrm := &ormmock.Orm{}
	mockQuery := &ormmock.Query{}
	mockOrm.On("Connection", "database").Return(mockOrm)
	mockOrm.On("Query").Return(mockQuery)
	mockQuery.On("Table", "failed_jobs").Return(mockQuery)

	OrmFacade = mockOrm

	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	orm := s.config.FailedJobsDatabase()

	s.NotNil(orm)
	s.mockConfig.AssertExpectations(s.T())
}
