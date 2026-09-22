package queue

import (
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/models"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
)

func (s *WorkerTestSuite) TestSplitQueueNames() {
	tests := []struct {
		name   string
		queue  string
		expect []string
	}{
		{
			name:   "single queue",
			queue:  "default",
			expect: []string{"default"},
		},
		{
			name:   "comma separated queues",
			queue:  "high,default",
			expect: []string{"high", "default"},
		},
		{
			name:   "trimmed and empty queue names",
			queue:  " high, , default,",
			expect: []string{"high", "default"},
		},
		{
			name:   "whitespace-only queue falls back to default",
			queue:  " , ",
			expect: []string{"default"},
		},
		{
			name:   "empty queue falls back to default",
			queue:  "",
			expect: []string{"default"},
		},
		{
			name:   "duplicate queues are preserved",
			queue:  "high,high,default",
			expect: []string{"high", "high", "default"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expect, splitQueueNames(tt.queue))
		})
	}
}

func (s *WorkerTestSuite) TestPopNextJob() {
	s.Run("falls through empty queues in priority order", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound.Args("high")).Once()
		s.mockDriver.EXPECT().Pop("default").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.popNextJob([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("default", actualQueue)
	})

	s.Run("only pops from the first queue holding a job", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.popNextJob([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("high", actualQueue)
	})

	s.Run("skips a nil job returned without an error", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(nil, nil).Once()
		s.mockDriver.EXPECT().Pop("default").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.popNextJob([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("default", actualQueue)
	})

	s.Run("stops on driver errors", func() {
		s.SetupTest()

		s.mockDriver.EXPECT().Pop("high").Return(nil, assert.AnError).Once()

		actualJob, actualQueue, err := s.worker.popNextJob([]string{"high", "default"})

		s.Nil(actualJob)
		s.Equal("high", actualQueue)
		s.Equal(assert.AnError, err)
	})

	s.Run("reports the last queue when all queues are empty", func() {
		s.SetupTest()

		lastErr := errors.QueueDriverNoJobFound.Args("default")
		s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound.Args("high")).Once()
		s.mockDriver.EXPECT().Pop("default").Return(nil, lastErr).Once()

		actualJob, actualQueue, err := s.worker.popNextJob([]string{"high", "default"})

		s.Nil(actualJob)
		s.Equal("default", actualQueue)
		s.Same(lastErr, err)
	})
}

func (s *WorkerTestSuite) TestRunWithMultipleQueuesPops() {
	s.SetupTest()
	s.worker.queue = "high,default"

	mockDriverWithReceive := mocksqueue.NewDriverWithReceive(s.T())
	s.worker.driver = &receiveDriver{
		driver:   s.mockDriver,
		receiver: mockDriverWithReceive,
	}

	poppedFromAllQueues := make(chan struct{})
	s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound.Args("high")).Once()
	s.mockDriver.EXPECT().Pop("default").Return(nil, errors.QueueDriverNoJobFound.Args("default")).
		Run(func(string) {
			close(poppedFromAllQueues)
		}).Once()

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- s.worker.run()
	}()

	select {
	case <-poppedFromAllQueues:
	case <-time.After(5 * time.Second):
		s.Fail("the worker did not pop from every configured queue")
	}

	s.NoError(s.worker.Shutdown())
	select {
	case err := <-runErrChan:
		s.NoError(err)
	case <-time.After(5 * time.Second):
		s.Fail("the worker did not stop after shutdown")
	}
	mockDriverWithReceive.AssertNotCalled(s.T(), "Receive", mock.Anything, mock.Anything, mock.Anything)
}

func (s *WorkerTestSuite) TestRunRecordsReservedQueue() {
	s.SetupTest()
	carbon.SetTestNow(carbon.FromStdTime(time.Now()))
	defer carbon.ClearTestNow()

	s.worker.queue = "high,default"

	task := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{Job: &TestJobErr{}},
		UUID:     "test",
	}
	reservedJob := mocksqueue.NewReservedJob(s.T())
	lastPollCompleted := make(chan struct{})
	s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound.Args("high")).Twice()
	s.mockDriver.EXPECT().Pop("default").Return(reservedJob, nil).Once()
	s.mockDriver.EXPECT().Pop("default").Return(nil, errors.QueueDriverNoJobFound.Args("default")).
		Run(func(string) {
			close(lastPollCompleted)
		}).Once()
	reservedJob.EXPECT().Task().Return(task).Once()
	reservedJob.EXPECT().Attempts().Return(1).Once()
	reservedJob.EXPECT().Delete().Return(nil).Once()

	s.mockJob.EXPECT().Call(task.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()
	s.mockJson.EXPECT().MarshalString(utils.Task{
		Job:  utils.Job{Signature: task.Job.Signature()},
		UUID: "test",
	}).Return("{}", nil).Once()

	failedJob := &models.FailedJob{
		UUID:       "test",
		Connection: "sync",
		Queue:      "default",
		Payload:    "{}",
		Exception:  assert.AnError.Error(),
		FailedAt:   carbon.NewDateTime(carbon.Now()),
	}
	s.mockConfig.EXPECT().FailedDatabase().Return("mysql").Once()
	s.mockConfig.EXPECT().FailedTable().Return("failed_jobs").Once()
	s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
	mockQuery := mocksdb.NewQuery(s.T())
	s.mockDB.EXPECT().Table("failed_jobs").Return(mockQuery).Once()
	failedJobSaved := make(chan struct{})
	mockQuery.EXPECT().Insert(failedJob).Run(func(any) {
		close(failedJobSaved)
	}).Return(nil, nil).Once()

	runErrChan := make(chan error, 1)
	go func() {
		runErrChan <- s.worker.run()
	}()

	for _, completed := range []chan struct{}{lastPollCompleted, failedJobSaved} {
		select {
		case <-completed:
		case <-time.After(5 * time.Second):
			s.Fail("the worker did not finish processing the failed job")
		}
	}

	s.NoError(s.worker.Shutdown())
	select {
	case err := <-runErrChan:
		s.NoError(err)
	case <-time.After(5 * time.Second):
		s.Fail("the worker did not stop after shutdown")
	}
}
