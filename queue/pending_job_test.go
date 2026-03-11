package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type PendingJobTestSuite struct {
	suite.Suite
	mockDriverCreator *mocksqueue.DriverCreator
	pendingJob        *PendingJob
}

func TestPendingJobTestSuite(t *testing.T) {
	suite.Run(t, new(PendingJobTestSuite))
}

func (s *PendingJobTestSuite) SetupTest() {
	s.mockDriverCreator = mocksqueue.NewDriverCreator(s.T())
	s.pendingJob = &PendingJob{
		connection:    "default",
		driverCreator: s.mockDriverCreator,
		queue:         "default",
		task: contractsqueue.Task{
			UUID: "test",
			ChainJob: contractsqueue.ChainJob{
				Job: &TestJobOne{},
			},
		},
	}
}

func (s *PendingJobTestSuite) Test_Delay() {
	delayTime := time.Now().Add(5 * time.Minute)
	pendingJobWithDelay := s.pendingJob.Delay(delayTime)

	s.Equal(delayTime, s.pendingJob.delay)
	s.Equal(s.pendingJob, pendingJobWithDelay)
}

func (s *PendingJobTestSuite) Test_Dispatch() {
	s.Run("happy path", func() {
		s.SetupTest()

		mockDriver := mocksqueue.NewDriver(s.T())
		s.mockDriverCreator.EXPECT().Create("default").Return(mockDriver, nil).Once()
		mockDriver.EXPECT().Push(s.pendingJob.task, s.pendingJob.queue).Return(nil).Once()

		err := s.pendingJob.Dispatch()

		s.NoError(err)
	})

	s.Run("happy path with custom connection and queue", func() {
		s.SetupTest()

		mockDriver := mocksqueue.NewDriver(s.T())
		s.mockDriverCreator.EXPECT().Create("kafka").Return(mockDriver, nil).Once()
		mockDriver.EXPECT().Push(s.pendingJob.task, "high").Return(nil).Once()

		err := s.pendingJob.OnConnection("kafka").OnQueue("high").Dispatch()

		s.NoError(err)
	})

	s.Run("failed to create driver", func() {
		s.SetupTest()

		s.mockDriverCreator.EXPECT().Create("default").Return(nil, assert.AnError).Once()

		err := s.pendingJob.Dispatch()

		s.Equal(assert.AnError, err)
	})
}

func (s *PendingJobTestSuite) TestDispatchSync() {
	s.Run("happy path", func() {
		err := s.pendingJob.DispatchSync()

		s.NoError(err)
	})
}

func (s *PendingJobTestSuite) TestNewPendingChainJob() {
	jobs := []contractsqueue.ChainJob{
		{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{
					Type:  "string",
					Value: "arg1",
				},
			},
		},
		{
			Job: &TestJobOne{},
			Args: []contractsqueue.Arg{
				{
					Type:  "string",
					Value: "arg2",
				},
			},
			Delay: time.Now().Add(1 * time.Minute),
		},
	}

	mockConfig := mocksqueue.NewConfig(s.T())
	mockConfig.EXPECT().DefaultConnection().Return("default").Once()
	mockConfig.EXPECT().DefaultQueue().Return("default").Once()

	pendingChainJob := NewPendingChainJob(mockConfig, nil, nil, nil, nil, jobs, nil)

	s.Equal("default", pendingChainJob.connection)
	s.Equal("default", pendingChainJob.queue)
	s.NotEmpty(pendingChainJob.task.UUID)
	s.Equal(jobs[0].Job, pendingChainJob.task.Job)
	s.Equal(jobs[0].Args, pendingChainJob.task.Args)
	s.True(pendingChainJob.delay.IsZero())
	s.Equal(jobs[1].Job, pendingChainJob.task.Chain[0].Job)
	s.Equal(jobs[1].Args, pendingChainJob.task.Chain[0].Args)
	s.Equal(jobs[1].Delay, pendingChainJob.task.Chain[0].Delay)
}

func (s *PendingJobTestSuite) TestNewPendingJob() {
	args := []contractsqueue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
		{
			Type:  "string",
			Value: "arg2",
		},
	}

	s.Run("with args", func() {
		mockConfig := mocksqueue.NewConfig(s.T())
		mockConfig.EXPECT().DefaultConnection().Return("default").Once()
		mockConfig.EXPECT().DefaultQueue().Return("default").Once()

		pendingJob := NewPendingJob(mockConfig, nil, nil, nil, nil, &TestJobOne{}, nil, args)

		s.Equal("default", pendingJob.connection)
		s.Equal("default", pendingJob.queue)
		s.NotEmpty(pendingJob.task.UUID)
		s.Equal(&TestJobOne{}, pendingJob.task.Job)
		s.Equal(args, pendingJob.task.Args)
		s.True(pendingJob.delay.IsZero())
	})

	s.Run("without args", func() {
		mockConfig := mocksqueue.NewConfig(s.T())
		mockConfig.EXPECT().DefaultConnection().Return("default").Once()
		mockConfig.EXPECT().DefaultQueue().Return("default").Once()

		pendingJob := NewPendingJob(mockConfig, nil, nil, nil, nil, &TestJobOne{}, nil)

		s.Equal("default", pendingJob.connection)
		s.Equal("default", pendingJob.queue)
		s.NotEmpty(pendingJob.task.UUID)
		s.Equal(&TestJobOne{}, pendingJob.task.Job)
		s.Empty(pendingJob.task.Args)
		s.True(pendingJob.delay.IsZero())
	})
}

func (s *PendingJobTestSuite) TestOnConnection() {
	pendingJobWithNewConnection := s.pendingJob.OnConnection("redis")

	s.Equal("redis", s.pendingJob.connection)
	s.Equal(s.pendingJob, pendingJobWithNewConnection)
}

func (s *PendingJobTestSuite) TestOnQueue() {
	newQueue := "high"
	pendingJobWithNewQueue := s.pendingJob.OnQueue(newQueue)

	s.Equal("high", s.pendingJob.queue)
	s.Equal(s.pendingJob, pendingJobWithNewQueue)
}
