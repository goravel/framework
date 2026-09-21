package queue

import (
	"context"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/queue/utils"
)

func (s *WorkerTestSuite) TestQueueNames() {
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
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.worker.queue = tt.queue

			s.Equal(tt.expect, s.worker.queueNames())
		})
	}
}

func (s *WorkerTestSuite) TestPop() {
	s.Run("falls through empty queues in priority order", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound).Once()
		s.mockDriver.EXPECT().Pop("default").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.pop([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("default", actualQueue)
	})

	s.Run("only pops from the first queue holding a job", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.pop([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("high", actualQueue)
	})

	s.Run("skips a nil job returned without an error", func() {
		s.SetupTest()

		reservedJob := mocksqueue.NewReservedJob(s.T())
		s.mockDriver.EXPECT().Pop("high").Return(nil, nil).Once()
		s.mockDriver.EXPECT().Pop("default").Return(reservedJob, nil).Once()

		actualJob, actualQueue, err := s.worker.pop([]string{"high", "default"})

		s.NoError(err)
		s.Same(reservedJob, actualJob)
		s.Equal("default", actualQueue)
	})

	s.Run("handles an empty queue list", func() {
		s.SetupTest()

		actualJob, actualQueue, err := s.worker.pop(nil)

		s.Nil(actualJob)
		s.Empty(actualQueue)
		s.ErrorIs(err, errors.QueueDriverNoJobFound)
	})

	s.Run("stops on driver errors", func() {
		s.SetupTest()

		s.mockDriver.EXPECT().Pop("high").Return(nil, assert.AnError).Once()

		actualJob, actualQueue, err := s.worker.pop([]string{"high", "default"})

		s.Nil(actualJob)
		s.Equal("high", actualQueue)
		s.Equal(assert.AnError, err)
	})

	s.Run("reports the last queue when all queues are empty", func() {
		s.SetupTest()

		s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound).Once()
		s.mockDriver.EXPECT().Pop("default").Return(nil, errors.QueueDriverNoJobFound).Once()

		actualJob, actualQueue, err := s.worker.pop([]string{"high", "default"})

		s.Nil(actualJob)
		s.Equal("default", actualQueue)
		s.ErrorIs(err, errors.QueueDriverNoJobFound)
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

	// Receive only accepts one queue, so a multi-queue worker must not call it.
	receiveCalled := make(chan struct{}, 1)
	mockDriverWithReceive.EXPECT().Receive(mock.Anything, mock.Anything, mock.Anything).
		Run(func(context.Context, string, int) {
			select {
			case receiveCalled <- struct{}{}:
			default:
			}
		}).Return(nil, nil).Maybe()

	poppedFromAllQueues := make(chan struct{})
	s.mockDriver.EXPECT().Pop("high").Return(nil, errors.QueueDriverNoJobFound).Once()
	s.mockDriver.EXPECT().Pop("default").Return(nil, errors.QueueDriverNoJobFound).
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

	select {
	case <-receiveCalled:
		s.Fail("a multi-queue worker must not receive from a single queue")
	default:
	}

	s.NoError(s.worker.Shutdown())
	s.NoError(<-runErrChan)
}

func (s *WorkerTestSuite) TestCallRecordsReservedQueue() {
	s.SetupTest()

	task := contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{Job: &TestJobErr{}},
		UUID:     "test",
	}
	s.mockJob.EXPECT().Call(task.Job.Signature(), make([]any, 0)).Return(assert.AnError).Once()
	s.mockJson.EXPECT().MarshalString(utils.Task{
		Job:  utils.Job{Signature: task.Job.Signature()},
		UUID: "test",
	}).Return("{}", nil).Once()

	// The worker is configured for "default", the job was reserved from "high".
	released, err := s.worker.call(task, nil, "high")

	s.False(released)
	s.Equal(errors.QueueFailedToCallJob, err)
	s.Equal("high", (<-s.worker.failedJobChan).Queue)
}
