// TODO: Will be removed in v1.17

package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
)

type MachineryTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	mockLog    *mockslog.Log
	machinery  *Machinery
}

func TestMachineryTestSuite(t *testing.T) {
	suite.Run(t, new(MachineryTestSuite))
}

func (s *MachineryTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockLog = mockslog.NewLog(s.T())
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
			name:       "redis",
			connection: "redis",
			setup: func() {
				s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Once()
				s.mockConfig.On("GetString", "database.redis.default.host").Return("127.0.0.1").Once()
				s.mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
				s.mockConfig.On("GetInt", "database.redis.default.port").Return(6379).Once()
				s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
				s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
				s.mockConfig.On("GetString", "app.name").Return("goravel").Once()
				s.mockConfig.On("GetBool", "app.debug").Return(true).Once()
			},
			expectServer: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.machinery = NewMachinery(test.connection, NewConfig(s.mockConfig), s.mockLog)
			test.setup()
			server := s.machinery.server(test.queue)
			s.Equal(test.expectServer, server != nil)
		})
	}
}
