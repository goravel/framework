package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	configmock "github.com/goravel/framework/mocks/config"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	queuemock "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
)

var (
	testAsyncJob       = 0
	testDelayAsyncJob  = 0
	testCustomAsyncJob = 0
	testErrorAsyncJob  = 0
	testChainAsyncJob  = 0
)

type DriverAsyncTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *configmock.Config
	mockQueue  *queuemock.Queue
}

func TestDriverAsyncTestSuite(t *testing.T) {
	suite.Run(t, &DriverAsyncTestSuite{})
}

func (s *DriverAsyncTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.mockQueue = &queuemock.Queue{}
	s.app = NewApplication(s.mockConfig)

	JobRegistry = new(sync.Map)
	testAsyncJob = 0
	testDelayAsyncJob = 0
	testCustomAsyncJob = 0
	testErrorAsyncJob = 0
	testChainAsyncJob = 0

	mockOrm := &ormmock.Orm{}
	mockQuery := &ormmock.Query{}
	mockOrm.On("Connection", "database").Return(mockOrm)
	mockOrm.On("Query").Return(mockQuery)
	mockQuery.On("Table", "failed_jobs").Return(mockQuery)

	OrmFacade = mockOrm

	s.Nil(s.app.Register([]queue.Job{&TestAsyncJob{}, &TestDelayAsyncJob{}, &TestCustomAsyncJob{}, &TestErrorAsyncJob{}, &TestChainAsyncJob{}}))
}

func (s *DriverAsyncTestSuite) TestDefaultAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("async").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(2)
	s.mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.async.driver").Return("async").Times(3)
	s.mockConfig.On("GetString", "queue.failed.database").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(nil).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestDefaultAsyncQueue"},
		{Type: "int", Value: 1},
	}).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverAsyncTestSuite) TestDelayAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("async").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.async.driver").Return("async").Times(3)
	s.mockConfig.On("GetString", "queue.failed.database").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(&queue.Args{
			Queue: "delay",
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestDelayAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestDelayAsyncQueue"},
		{Type: "int", Value: 1},
	}).OnQueue("delay").Delay(carbon.Now().AddSeconds(3)).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testDelayAsyncJob)
	time.Sleep(3 * time.Second)
	s.Equal(1, testDelayAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverAsyncTestSuite) TestCustomAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("custom").Times(7)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.custom.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.custom.driver").Return("async").Times(4)
	s.mockConfig.On("GetString", "queue.failed.database").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(&queue.Args{
			Connection: "custom",
			Queue:      "custom1",
			Concurrent: 2,
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestCustomAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestCustomAsyncQueue"},
		{Type: "int", Value: 1},
	}).OnConnection("custom").OnQueue("custom1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testCustomAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverAsyncTestSuite) TestErrorAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("async").Times(7)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.async.driver").Return("async").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("").Once()
	s.mockConfig.On("GetString", "queue.failed.database").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(&queue.Args{
			Queue: "error",
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Error(s.app.Job(&TestErrorAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestErrorAsyncQueue"},
		{Type: "int", Value: 1},
	}).OnConnection("redis").OnQueue("error1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testErrorAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverAsyncTestSuite) TestChainAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("async").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.async.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.async.driver").Return("async").Times(3)
	s.mockConfig.On("GetString", "queue.failed.database").Return("database").Once()
	s.mockConfig.On("GetString", "queue.failed.table").Return("failed_jobs").Once()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(&queue.Args{
			Queue: "chain",
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)

	time.Sleep(2 * time.Second)
	s.Nil(s.app.Chain([]queue.Jobs{
		{
			Job: &TestChainAsyncJob{},
			Args: []queue.Arg{
				{Type: "string", Value: "TestChainAsyncJob"},
				{Type: "int", Value: 1},
			},
		},
		{
			Job: &TestAsyncJob{},
			Args: []queue.Arg{
				{Type: "string", Value: "TestAsyncJob"},
				{Type: "int", Value: 1},
			},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainAsyncJob)
	s.Equal(1, testAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
}

type TestAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestAsyncJob) Signature() string {
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestAsyncJob) Handle(args ...any) error {
	testAsyncJob++

	return nil
}

type TestDelayAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestDelayAsyncJob) Signature() string {
	return "test_delay_async_job"
}

// Handle Execute the job.
func (receiver *TestDelayAsyncJob) Handle(args ...any) error {
	testDelayAsyncJob++

	return nil
}

type TestCustomAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestCustomAsyncJob) Signature() string {
	return "test_custom_async_job"
}

// Handle Execute the job.
func (receiver *TestCustomAsyncJob) Handle(args ...any) error {
	testCustomAsyncJob++

	return nil
}

type TestErrorAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestErrorAsyncJob) Signature() string {
	return "test_error_async_job"
}

// Handle Execute the job.
func (receiver *TestErrorAsyncJob) Handle(args ...any) error {
	testErrorAsyncJob++

	return nil
}

type TestChainAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestChainAsyncJob) Signature() string {
	return "test_chain_async_job"
}

// Handle Execute the job.
func (receiver *TestChainAsyncJob) Handle(args ...any) error {
	testChainAsyncJob++

	return nil
}
