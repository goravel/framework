package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
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
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
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
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{}, args)

	s.Equal(s.mockConfig, pendingJob.config)
	s.Equal("default", pendingJob.connection)
	s.Equal("default", pendingJob.queue)
	s.NotEmpty(pendingJob.task.Uuid)
	s.Equal(&TestJobOne{}, pendingJob.task.Jobs.Job)
	s.Equal(args, pendingJob.task.Args)
	s.True(pendingJob.delay.IsZero())
}

func (s *PendingJobTestSuite) TestNewPendingJobWithoutArgs() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{})

	s.Equal(s.mockConfig, pendingJob.config)
	s.Equal("default", pendingJob.connection)
	s.Equal("default", pendingJob.queue)
	s.NotEmpty(pendingJob.task.Uuid)
	s.Equal(&TestJobOne{}, pendingJob.task.Jobs.Job)
	s.Empty(pendingJob.task.Args)
	s.True(pendingJob.delay.IsZero())
}

func (s *PendingJobTestSuite) TestNewPendingChainJob() {
	jobs := []queue.Jobs{
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

	pendingChainJob := NewPendingChainJob(s.mockConfig, jobs)

	s.Equal(s.mockConfig, pendingChainJob.config)
	s.Equal("default", pendingChainJob.connection)
	s.Equal("default", pendingChainJob.queue)
	s.NotEmpty(pendingChainJob.task.Uuid)
	s.Equal(jobs[0].Job, pendingChainJob.task.Job)
	s.Equal(jobs[0].Args, pendingChainJob.task.Args)
	s.True(pendingChainJob.delay.IsZero())
	s.Equal(jobs[1].Job, pendingChainJob.task.Chain[0].Job)
	s.Equal(jobs[1].Args, pendingChainJob.task.Chain[0].Args)
	s.Equal(jobs[1].Delay, pendingChainJob.task.Chain[0].Delay)
}

func (s *PendingJobTestSuite) TestDelay() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{})

	delayTime := time.Now().Add(5 * time.Minute)
	pendingJobWithDelay := pendingJob.Delay(delayTime)

	s.Equal(delayTime, pendingJob.delay)
	s.Equal(pendingJob, pendingJobWithDelay)
}

func (s *PendingJobTestSuite) TestOnConnection() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{})

	newConnection := "redis"
	pendingJobWithNewConnection := pendingJob.OnConnection(newConnection)

	s.Equal(newConnection, pendingJob.connection)
	s.Equal(pendingJob, pendingJobWithNewConnection)
}

func (s *PendingJobTestSuite) TestOnQueue() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{})

	newQueue := "high"
	pendingJobWithNewQueue := pendingJob.OnQueue(newQueue)

	s.Equal("high", pendingJob.queue)
	s.Equal(pendingJob, pendingJobWithNewQueue)
}

func (s *PendingJobTestSuite) TestDispatchSync() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobOne{}, []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	})

	err := pendingJob.DispatchSync()

	s.Nil(err)
	s.Equal([]any{"arg1"}, testJobOne)
}

func (s *PendingJobTestSuite) TestDispatchSyncWithError() {
	pendingJob := NewPendingJob(s.mockConfig, &TestJobErr{}, []queue.Arg{
		{
			Type:  "string",
			Value: "arg1",
		},
	})

	err := pendingJob.DispatchSync()

	s.Equal(assert.AnError, err)
}

func (s *PendingJobTestSuite) TestDispatchSyncChain() {
	jobs := []queue.Jobs{
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

	pendingChainJob := NewPendingChainJob(s.mockConfig, jobs)

	err := pendingChainJob.DispatchSync()

	s.Nil(err)
	s.Equal([]any{"arg1"}, testJobOne)
	s.Equal([]any{"arg2"}, testJobTwo)
}

func (s *PendingJobTestSuite) TestDispatchSyncChainWithError() {
	jobs := []queue.Jobs{
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

	pendingChainJob := NewPendingChainJob(s.mockConfig, jobs)

	err := pendingChainJob.DispatchSync()

	s.Equal(assert.AnError, err)
	s.Equal([]any{"arg1"}, testJobOne)
}
