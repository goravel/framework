package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type TaskTestSuite struct {
	suite.Suite
	mockConfig *mocksqueue.Config
	mockJob    *mocksqueue.Job
}

func TestTaskTestSuite(t *testing.T) {
	suite.Run(t, new(TaskTestSuite))
}

func (s *TaskTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockJob = mocksqueue.NewJob(s.T())
}

func (s *TaskTestSuite) TestNewTask() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()

	// Create a new task
	args := []any{"arg1", "arg2"}
	task := NewTask(s.mockConfig, s.mockJob, args)

	// Assertions
	s.Equal(s.mockConfig, task.config)
	s.Equal("default", task.connection)
	s.Equal("default_queue", task.queue)
	s.False(task.chain)
	s.True(task.delay.IsZero())
	s.Len(task.jobs, 1)
	s.Equal(s.mockJob, task.jobs[0].Job)
	s.Equal(args, task.jobs[0].Args)
}

func (s *TaskTestSuite) TestNewTaskWithoutArgs() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()

	// Create a new task without args
	task := NewTask(s.mockConfig, s.mockJob)

	// Assertions
	s.Equal(s.mockConfig, task.config)
	s.Equal("default", task.connection)
	s.Equal("default_queue", task.queue)
	s.False(task.chain)
	s.True(task.delay.IsZero())
	s.Len(task.jobs, 1)
	s.Equal(s.mockJob, task.jobs[0].Job)
	s.Empty(task.jobs[0].Args)
}

func (s *TaskTestSuite) TestNewChainTask() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()

	// Create jobs for the chain
	jobs := []queue.Jobs{
		{
			Job:  s.mockJob,
			Args: []any{"arg1"},
		},
		{
			Job:  s.mockJob,
			Args: []any{"arg2"},
		},
	}

	// Create a new chain task
	task := NewChainTask(s.mockConfig, jobs)

	// Assertions
	s.Equal(s.mockConfig, task.config)
	s.Equal("default", task.connection)
	s.Equal("default_queue", task.queue)
	s.True(task.chain)
	s.True(task.delay.IsZero())
	s.Equal(jobs, task.jobs)
}

func (s *TaskTestSuite) TestDelay() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()

	// Create a new task
	task := NewTask(s.mockConfig, s.mockJob)

	// Set a delay
	delayTime := time.Now().Add(5 * time.Minute)
	taskWithDelay := task.Delay(delayTime)

	// Assertions
	s.Equal(delayTime, task.delay)
	s.Equal(task, taskWithDelay)
}

func (s *TaskTestSuite) TestOnConnection() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()

	// Create a new task
	task := NewTask(s.mockConfig, s.mockJob)

	// Change connection
	newConnection := "redis"
	taskWithNewConnection := task.OnConnection(newConnection)

	// Assertions
	s.Equal(newConnection, task.connection)
	s.Equal(task, taskWithNewConnection)
}

func (s *TaskTestSuite) TestOnQueue() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()
	s.mockConfig.EXPECT().Queue("default", "high").Return("high_queue").Once()

	// Create a new task
	task := NewTask(s.mockConfig, s.mockJob)

	// Change queue
	newQueue := "high"
	taskWithNewQueue := task.OnQueue(newQueue)

	// Assertions
	s.Equal("high_queue", task.queue)
	s.Equal(task, taskWithNewQueue)
}

func (s *TaskTestSuite) TestDispatchSync() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()
	s.mockJob.EXPECT().Handle([]any{"arg1"}...).Return(nil).Once()

	// Create a new task
	task := NewTask(s.mockConfig, s.mockJob, []any{"arg1"})

	// Dispatch synchronously
	err := task.DispatchSync()

	// Assertions
	s.Nil(err)
}

func (s *TaskTestSuite) TestDispatchSyncWithError() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()
	s.mockJob.EXPECT().Handle([]any{"arg1"}...).Return(assert.AnError).Once()

	// Create a new task
	task := NewTask(s.mockConfig, s.mockJob, []any{"arg1"})

	// Dispatch synchronously
	err := task.DispatchSync()

	// Assertions
	s.Equal(assert.AnError, err)
}

func (s *TaskTestSuite) TestDispatchSyncChain() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()
	s.mockJob.EXPECT().Handle([]any{"arg1"}...).Return(nil).Once()
	s.mockJob.EXPECT().Handle([]any{"arg2"}...).Return(nil).Once()

	// Create jobs for the chain
	jobs := []queue.Jobs{
		{
			Job:  s.mockJob,
			Args: []any{"arg1"},
		},
		{
			Job:  s.mockJob,
			Args: []any{"arg2"},
		},
	}

	// Create a new chain task
	task := NewChainTask(s.mockConfig, jobs)

	// Dispatch synchronously
	err := task.DispatchSync()

	// Assertions
	s.Nil(err)
}

func (s *TaskTestSuite) TestDispatchSyncChainWithError() {
	// Setup expectations
	s.mockConfig.EXPECT().Default().Return("default", "default", 1).Once()
	s.mockConfig.EXPECT().Queue("default", "default").Return("default_queue").Once()
	s.mockJob.EXPECT().Handle([]any{"arg1"}...).Return(nil).Once()
	s.mockJob.EXPECT().Handle([]any{"arg2"}...).Return(assert.AnError).Once()

	// Create jobs for the chain
	jobs := []queue.Jobs{
		{
			Job:  s.mockJob,
			Args: []any{"arg1"},
		},
		{
			Job:  s.mockJob,
			Args: []any{"arg2"},
		},
	}

	// Create a new chain task
	task := NewChainTask(s.mockConfig, jobs)

	// Dispatch synchronously
	err := task.DispatchSync()

	// Assertions
	s.Equal(assert.AnError, err)
}
