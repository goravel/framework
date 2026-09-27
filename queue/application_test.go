package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type ApplicationTestSuite struct {
	suite.Suite
	mockConfig *mocksqueue.Config
}

func TestApplicationTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (s *ApplicationTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
}

func (s *ApplicationTestSuite) TestWorkerUsesConfiguredQueueListForDefaultConnection() {
	s.mockConfig.EXPECT().DefaultConnection().Return("sync").Once()
	s.mockConfig.EXPECT().DefaultConcurrent().Return(2).Once()
	s.mockConfig.EXPECT().GetString("queue.connections.sync.queue", "default").Return("high,default").Once()
	s.mockConfig.EXPECT().Driver("sync").Return(contractsqueue.DriverSync).Once()
	s.mockConfig.EXPECT().Debug().Return(false).Once()

	worker := NewApplication(s.mockConfig, nil, nil, nil, nil, nil).Worker().(*Worker)

	s.Equal("high,default", worker.queue)
	s.Equal(2, worker.concurrent)
}

func (s *ApplicationTestSuite) TestWorkerUsesConfiguredQueueListForSelectedConnection() {
	s.mockConfig.EXPECT().DefaultConnection().Return("sync").Once()
	s.mockConfig.EXPECT().DefaultConcurrent().Return(1).Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("high,default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(3).Once()
	s.mockConfig.EXPECT().Driver("redis").Return(contractsqueue.DriverSync).Once()
	s.mockConfig.EXPECT().Debug().Return(false).Once()

	worker := NewApplication(s.mockConfig, nil, nil, nil, nil, nil).Worker(contractsqueue.Args{
		Connection: "redis",
	}).(*Worker)

	s.Equal("high,default", worker.queue)
	s.Equal(3, worker.concurrent)
}
