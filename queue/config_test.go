package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
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
				s.mockConfig.On("GetString", "queue.default").Return("redis").Once()
				s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("queue").Once()
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
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Once()

	redisConfig, database, queue := s.config.Redis("redis")

	s.Equal("127.0.0.1:6379", redisConfig)
	s.Equal(0, database)
	s.Equal("goravel_queues:default", queue)
}
