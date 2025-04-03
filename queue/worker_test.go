package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type WorkerTestSuite struct {
	suite.Suite
	mockConfig *mocksqueue.Config
	mockLog    *mockslog.Log
	mockJob    *mocksqueue.JobRepository
	worker     *Worker
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (s *WorkerTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.mockJob = mocksqueue.NewJobRepository(s.T())

	s.worker = NewWorker(s.mockConfig, 2, "sync", "default", s.mockJob, s.mockLog)
}

func (s *WorkerTestSuite) TestNewWorker() {
	s.Equal(2, s.worker.concurrent)
	s.Equal("sync", s.worker.connection)
	s.Equal("default", s.worker.queue)
	s.Equal(1*time.Second, s.worker.currentDelay)
	s.Equal(32*time.Second, s.worker.maxDelay)
}

func (s *WorkerTestSuite) TestRunWithSyncDriver() {
	s.mockConfig.EXPECT().Driver("sync").Return(queue.DriverSync).Once()

	err := s.worker.Run()
	s.Equal(errors.QueueDriverSyncNotNeedToRun.Args("sync"), err)
}

func (s *WorkerTestSuite) TestRunWithUnknownDriver() {
	s.mockConfig.EXPECT().Driver("sync").Return("unknown").Once()

	err := s.worker.Run()
	s.Equal(errors.QueueDriverNotSupported.Args("unknown"), err)
}

func (s *WorkerTestSuite) TestShutdown() {
	s.worker.isShutdown.Store(false)

	err := s.worker.Shutdown()
	s.NoError(err)
	s.True(s.worker.isShutdown.Load())
}
