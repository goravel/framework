package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type MachineryTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	machinery  *Machinery
}

func TestMachineryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineryTestSuite))
}

func (s *MachineryTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.machinery = NewMachinery(NewConfig(s.mockConfig))
}

func (s *MachineryTestSuite) TestServer() {
	tests := []struct {
		name         string
		connection   string
		queue        string
		setup        func()
		expectServer bool
		expectErr    bool
	}{
		{
			name:       "sync",
			connection: "sync",
			setup: func() {
				s.mockConfig.On("GetString", "queue.connections.sync.driver").Return("sync").Once()
			},
		},
		{
			name:       "redis",
			connection: "redis",
			setup: func() {
				s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Once()
				s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Once()
				s.mockConfig.On("GetString", "database.redis.default.host").Return("127.0.0.1").Once()
				s.mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
				s.mockConfig.On("GetInt", "database.redis.default.port").Return(6379).Once()
				s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
				s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
				s.mockConfig.On("GetString", "app.name").Return("goravel").Once()
			},
			expectServer: true,
		},
		{
			name:       "error",
			connection: "custom",
			setup: func() {
				s.mockConfig.On("GetString", "queue.connections.custom.driver").Return("custom").Once()

			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			server, err := s.machinery.Server(test.connection, test.queue)
			s.Equal(test.expectServer, server != nil)
			s.Equal(test.expectErr, err != nil)
			s.mockConfig.AssertExpectations(s.T())
		})
	}
}
