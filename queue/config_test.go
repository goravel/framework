package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

type ConfigTestSuite struct {
	suite.Suite
	config     queue.Config
	mockConfig *mocksconfig.Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
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
