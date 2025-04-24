package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
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

	s.worker = NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, "sync", "default", 2)
}

func (s *WorkerTestSuite) TestNewWorker() {
	s.Equal(2, s.worker.concurrent)
	s.Equal("sync", s.worker.connection)
	s.Equal("default", s.worker.queue)
	s.Equal(1*time.Second, s.worker.currentDelay)
	s.Equal(32*time.Second, s.worker.maxDelay)
}

func (s *WorkerTestSuite) Test_run() {
	now := time.Now()
	carbon.SetTestNow(carbon.FromStdTime(now))
	connection := "redis"
	queue := "default"
	queueKey := fmt.Sprintf("%s_%s:%s_%s", "goravel", "queues", connection, queue)
	errorTask := contractsqueue.Task{
		Jobs: contractsqueue.Jobs{
			Job:   &TestJobErr{},
			Delay: time.Now().Add(1 * time.Hour),
		},
		UUID:  "test",
		Chain: []contractsqueue.Jobs{},
	}
	successTask := contractsqueue.Task{
		Jobs: contractsqueue.Jobs{
			Job:   &TestJobOne{},
			Args:  testArgs,
			Delay: time.Now().Add(1 * time.Hour),
		},
		UUID:  "test",
		Chain: []contractsqueue.Jobs{},
	}

	s.Run("no job found", func() {
		s.mockConfig.EXPECT().QueueKey(connection, queue).Return(queueKey).Once()

		mockDriver := mocksqueue.NewDriver(s.T())
		mockDriver.EXPECT().Pop(queueKey).Return(contractsqueue.Task{}, errors.QueueDriverNoJobFound).Once()

		worker := NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, connection, queue, 1)

		go func() {
			err := worker.run(mockDriver)
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(worker.Shutdown())
	})

	s.Run("failed to pop job", func() {
		s.mockConfig.EXPECT().QueueKey(connection, queue).Return(queueKey).Once()

		mockDriver := mocksqueue.NewDriver(s.T())
		mockDriver.EXPECT().Pop(queueKey).Return(contractsqueue.Task{}, assert.AnError).Once()

		s.mockLog.EXPECT().Error(errors.QueueDriverFailedToPop.Args(queueKey, assert.AnError)).Once()

		worker := NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, connection, queue, 1)

		go func() {
			err := worker.run(mockDriver)
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(worker.Shutdown())
	})

	s.Run("failed to call job", func() {
		mockFailedJobsQuery := mocksdb.NewQuery(s.T())

		s.mockConfig.EXPECT().QueueKey(connection, queue).Return(queueKey).Once()

		mockDriver := mocksqueue.NewDriver(s.T())
		mockDriver.EXPECT().Pop(queueKey).Return(errorTask, nil).Once()

		s.mockJob.EXPECT().Call(errorTask.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()

		s.mockConfig.EXPECT().FailedJobsQuery().Return(mockFailedJobsQuery).Once()
		mockFailedJobsQuery.EXPECT().Insert(&FailedJob{
			UUID:       errorTask.UUID,
			Connection: connection,
			Queue:      queue,
			Payload:    "{\"signature\":\"test_job_err\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
			Exception:  assert.AnError.Error(),
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		}).Return(nil, nil).Once()

		mockDriver.EXPECT().Pop(queueKey).Return(contractsqueue.Task{}, errors.QueueDriverNoJobFound).Once()

		worker := NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, connection, queue, 1)

		go func() {
			err := worker.run(mockDriver)
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(worker.Shutdown())
	})

	s.Run("failed to insert failed job", func() {
		failedJob := &FailedJob{
			UUID:       errorTask.UUID,
			Connection: connection,
			Queue:      queue,
			Payload:    "{\"signature\":\"test_job_err\",\"args\":null,\"delay\":null,\"uuid\":\"test\",\"chain\":[]}",
			Exception:  assert.AnError.Error(),
			FailedAt:   carbon.NewDateTime(carbon.Now()),
		}

		mockFailedJobsQuery := mocksdb.NewQuery(s.T())

		s.mockConfig.EXPECT().QueueKey(connection, queue).Return(queueKey).Once()

		mockDriver := mocksqueue.NewDriver(s.T())
		mockDriver.EXPECT().Pop(queueKey).Return(errorTask, nil).Once()

		s.mockJob.EXPECT().Call(errorTask.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()

		s.mockConfig.EXPECT().FailedJobsQuery().Return(mockFailedJobsQuery).Once()
		mockFailedJobsQuery.EXPECT().Insert(failedJob).Return(nil, assert.AnError).Once()

		s.mockLog.EXPECT().Error(errors.QueueFailedToSaveFailedJob.Args(assert.AnError, failedJob)).Once()

		mockDriver.EXPECT().Pop(queueKey).Return(contractsqueue.Task{}, errors.QueueDriverNoJobFound).Once()

		worker := NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, connection, queue, 1)

		go func() {
			err := worker.run(mockDriver)
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(worker.Shutdown())
	})

	s.Run("success", func() {
		s.mockConfig.EXPECT().QueueKey(connection, queue).Return(queueKey).Once()

		mockDriver := mocksqueue.NewDriver(s.T())
		mockDriver.EXPECT().Pop(queueKey).Return(successTask, nil).Once()

		s.mockJob.EXPECT().Call(successTask.Job.Signature(), ConvertArgs(testArgs)).Return(nil).Once()

		mockDriver.EXPECT().Pop(queueKey).Return(contractsqueue.Task{}, errors.QueueDriverNoJobFound).Once()

		worker := NewWorker(s.mockConfig, s.mockJob, json.New(), s.mockLog, connection, queue, 1)

		go func() {
			err := worker.run(mockDriver)
			s.NoError(err)
		}()

		time.Sleep(500 * time.Millisecond)

		s.NoError(worker.Shutdown())
	})
}

func (s *WorkerTestSuite) TestRunWithSyncDriver() {
	s.mockConfig.EXPECT().Driver("sync").Return(contractsqueue.DriverSync).Once()

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
