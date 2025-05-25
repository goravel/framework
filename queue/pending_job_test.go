package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

type PendingJobTestSuite struct {
	suite.Suite
	mockConfig *mocksqueue.Config
}

func TestPendingJobTestSuite(t *testing.T) {
	suite.Run(t, new(PendingJobTestSuite))
}

func (s *PendingJobTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockConfig.EXPECT().DefaultConnection().Return("default").Once()
	s.mockConfig.EXPECT().DefaultQueue().Return("default").Once()
}

func (s *PendingJobTestSuite) TestDelay() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{})
	s.NoError(err)

	delayTime := time.Now().Add(5 * time.Minute)
	pendingJobWithDelay := pendingJob.Delay(delayTime)

	s.Equal(delayTime, pendingJob.delay)
	s.Equal(pendingJob, pendingJobWithDelay)
}

func (s *PendingJobTestSuite) TestDispatchWithSyncDriver() {
	s.mockConfig.EXPECT().Driver("default").Return(queue.DriverSync).Once()

	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	})
	s.NoError(err)

	err = pendingJob.Dispatch()

	s.NoError(err)
	s.Equal([]any{"arg1"}, testJobOne)
}

func (s *PendingJobTestSuite) TestDispatchWithCustomDriver() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	}
	mockDriver := mocksqueue.NewDriver(s.T())

	s.mockConfig.EXPECT().Driver("default").Return(queue.DriverCustom).Once()
	s.mockConfig.EXPECT().Via("default").Return(mockDriver).Once()
	mockDriver.EXPECT().Push(mock.MatchedBy(func(task queue.Task) bool {
		return s.IsType(&TestJobOne{}, task.Job) && s.ElementsMatch(task.Args, args) && task.Delay.IsZero()
	}), "default:default").Return(nil).Once()

	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, args)
	s.NoError(err)

	err = pendingJob.Dispatch()

	s.NoError(err)
}

func (s *PendingJobTestSuite) TestDispatchWithCustomDriverWithConnectionAndQueue() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	}
	mockDriver := mocksqueue.NewDriver(s.T())

	s.mockConfig.EXPECT().Driver("kafka").Return(queue.DriverCustom).Once()
	s.mockConfig.EXPECT().Via("kafka").Return(mockDriver).Once()
	mockDriver.EXPECT().Push(mock.MatchedBy(func(task queue.Task) bool {
		return s.IsType(&TestJobOne{}, task.Job) && s.ElementsMatch(task.Args, args) && task.Delay.IsZero()
	}), "kafka:high").Return(nil).Once()

	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, args)
	s.NoError(err)

	err = pendingJob.OnConnection("kafka").OnQueue("high").Dispatch()

	s.NoError(err)
}

func (s *PendingJobTestSuite) TestDispatchWithCustomDriverAndDelay() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	}
	mockDriver := mocksqueue.NewDriver(s.T())
	now := time.Now()

	s.mockConfig.EXPECT().Driver("default").Return(queue.DriverCustom).Once()
	s.mockConfig.EXPECT().Via("default").Return(mockDriver).Once()
	mockDriver.EXPECT().Push(mock.MatchedBy(func(task queue.Task) bool {
		return s.IsType(&TestJobOne{}, task.Job) && s.ElementsMatch(task.Args, args) && task.Delay.Equal(now.Add(1*time.Second))
	}), "default:default").Return(nil).Once()

	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, args)
	s.NoError(err)

	err = pendingJob.Delay(now.Add(1 * time.Second)).Dispatch()

	s.NoError(err)
}

func (s *PendingJobTestSuite) TestDispatchChainWithCustomDriverAndDelay() {
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	}
	mockDriver := mocksqueue.NewDriver(s.T())

	now := time.Now()
	carbon.SetTestNow(carbon.FromStdTime(now))

	s.mockConfig.EXPECT().Driver("default").Return(queue.DriverCustom).Once()
	s.mockConfig.EXPECT().Via("default").Return(mockDriver).Once()
	mockDriver.EXPECT().Push(mock.MatchedBy(func(task queue.Task) bool {
		return s.IsType(&TestJobOne{}, task.Job) &&
			s.ElementsMatch(task.Args, args) &&
			task.Delay.Equal(now.Add(2*time.Second)) &&
			len(task.Chain) == 1 &&
			s.IsType(&TestJobTwo{}, task.Chain[0].Job) &&
			s.ElementsMatch(task.Chain[0].Args, args) &&
			task.Chain[0].Delay.Equal(now.Add(2*time.Second))
	}), "default:default").Return(nil).Once()

	pendingJob, err := NewPendingChainJob(s.mockConfig, nil, nil, nil, []queue.ChainJob{
		{
			Job:   &TestJobOne{},
			Args:  args,
			Delay: now.Add(1 * time.Second),
		},
		{
			Job:   &TestJobTwo{},
			Args:  args,
			Delay: now.Add(2 * time.Second),
		},
	})
	s.NoError(err)

	err = pendingJob.Delay(now.Add(1 * time.Second)).Dispatch()

	s.NoError(err)
}

func (s *PendingJobTestSuite) TestDispatchSync() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	})
	s.NoError(err)

	err = pendingJob.DispatchSync()

	s.NoError(err)
	s.Equal([]any{"arg1"}, testJobOne)
}

func (s *PendingJobTestSuite) TestDispatchSyncChain() {
	jobs := []queue.ChainJob{
		{
			Job: &TestJobOne{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg1",
				},
			},
		},
		{
			Job: &TestJobTwo{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg2",
				},
			},
		},
	}

	pendingChainJob, err := NewPendingChainJob(s.mockConfig, nil, nil, nil, jobs)
	s.NoError(err)

	err = pendingChainJob.DispatchSync()

	s.Nil(err)
	s.Equal([]any{"arg1"}, testJobOne)
	s.Equal([]any{"arg2"}, testJobTwo)
}

func (s *PendingJobTestSuite) TestDispatchSyncChainWithError() {
	jobs := []queue.ChainJob{
		{
			Job: &TestJobOne{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg1",
				},
			},
		},
		{
			Job: &TestJobErr{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg2",
				},
			},
		},
	}

	pendingChainJob, err := NewPendingChainJob(s.mockConfig, nil, nil, nil, jobs)
	s.NoError(err)

	err = pendingChainJob.DispatchSync()

	s.Equal(assert.AnError, err)
	s.Equal([]any{"arg1"}, testJobOne)
}

func (s *PendingJobTestSuite) TestDispatchSyncWithError() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobErr{}, []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	})
	s.NoError(err)

	err = pendingJob.DispatchSync()

	s.Equal(assert.AnError, err)
}

func (s *PendingJobTestSuite) TestNewPendingChainJob() {
	jobs := []queue.ChainJob{
		{
			Job: &TestJobOne{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg1",
				},
			},
		},
		{
			Job: &TestJobOne{},
			Args: []queue.Arg{
				{
					Type:  "string",
					Value: "arg2",
				},
			},
			Delay: time.Now().Add(1 * time.Minute),
		},
	}

	pendingChainJob, err := NewPendingChainJob(s.mockConfig, nil, nil, nil, jobs)
	s.NoError(err)

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
	args := []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
		{
			Type:  "string",
			Value: "arg2",
		},
	}
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{}, args)
	s.NoError(err)

	s.Equal("default", pendingJob.connection)
	s.Equal("default", pendingJob.queue)
	s.NotEmpty(pendingJob.task.UUID)
	s.Equal(&TestJobOne{}, pendingJob.task.ChainJob.Job)
	s.Equal(args, pendingJob.task.Args)
	s.True(pendingJob.delay.IsZero())
}

func (s *PendingJobTestSuite) TestNewPendingJobWithoutArgs() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{})
	s.NoError(err)

	s.Equal("default", pendingJob.connection)
	s.Equal("default", pendingJob.queue)
	s.NotEmpty(pendingJob.task.UUID)
	s.Equal(&TestJobOne{}, pendingJob.task.ChainJob.Job)
	s.Empty(pendingJob.task.Args)
	s.True(pendingJob.delay.IsZero())
}

func (s *PendingJobTestSuite) TestOnConnection() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{})
	s.NoError(err)

	newConnection := "redis"
	pendingJobWithNewConnection := pendingJob.OnConnection(newConnection)

	s.Equal(newConnection, pendingJob.connection)
	s.Equal(pendingJob, pendingJobWithNewConnection)
}

func (s *PendingJobTestSuite) TestOnQueue() {
	pendingJob, err := NewPendingJob(s.mockConfig, nil, nil, nil, &TestJobOne{})
	s.NoError(err)

	newQueue := "high"
	pendingJobWithNewQueue := pendingJob.OnQueue(newQueue)

	s.Equal("high", pendingJob.queue)
	s.Equal(pendingJob, pendingJobWithNewQueue)
}
