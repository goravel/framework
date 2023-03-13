package queue

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/support"
	supporttime "github.com/goravel/framework/support/time"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"
)

var (
	testSyncJob        = 0
	testAsyncJob       = 0
	testDelayAsyncJob  = 0
	testCustomAsyncJob = 0
	testErrorAsyncJob  = 0
	testChainAsyncJob  = 0
	testChainSyncJob   = 0
)

type QueueTestSuite struct {
	suite.Suite
	app           *Application
	redisResource *dockertest.Resource
}

func TestQueueTestSuite(t *testing.T) {
	redisPool, redisResource, err := testingdocker.Redis()
	if err != nil {
		log.Fatalf("Get redis error: %s", err)
	}

	suite.Run(t, &QueueTestSuite{
		app:           NewApplication(),
		redisResource: redisResource,
	})

	if err := redisPool.Purge(redisResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *QueueTestSuite) SetupTest() {
}

func (s *QueueTestSuite) TestWorker() {
	var (
		mockConfig *configmocks.Config
	)

	beforeEach := func() {
		mockConfig = mock.Config()
	}

	tests := []struct {
		description  string
		setup        func()
		args         *queue.Args
		expectWorker queue.Worker
	}{
		{
			description: "success when args is nil",
			setup: func() {
				mockConfig.On("GetString", "queue.default").Return("redis").Once()
				mockConfig.On("GetString", "app.name").Return("app").Once()
				mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("queue").Once()
			},
			expectWorker: &support.Worker{
				Connection: "redis",
				Queue:      "app_queues:queue",
				Concurrent: 1,
			},
		},
		{
			description: "success when args isn't nil",
			setup: func() {
				mockConfig.On("GetString", "app.name").Return("app").Once()
			},
			args: &queue.Args{
				Connection: "redis",
				Queue:      "queue",
				Concurrent: 2,
			},
			expectWorker: &support.Worker{
				Connection: "redis",
				Queue:      "app_queues:queue",
				Concurrent: 2,
			},
		},
	}

	for _, test := range tests {
		beforeEach()
		test.setup()
		worker := s.app.Worker(test.args)
		s.Equal(test.expectWorker, worker, test.description)
		mockConfig.AssertExpectations(s.T())
	}
}

func (s *QueueTestSuite) TestSyncQueue() {
	s.Nil(s.app.Job(&TestSyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestSyncQueue"},
		{Type: "int", Value: 1},
	}).DispatchSync())
	s.Equal(1, testSyncJob)
}

func (s *QueueTestSuite) TestDefaultAsyncQueue() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "queue.default").Return("redis").Times(3)
	mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(cast.ToInt(s.redisResource.GetPort("6379/tcp"))).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	mockQueue, _ := mock.Queue()
	mockQueue.On("GetJobs").Return([]queue.Job{&TestAsyncJob{}}).Once()

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

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

	mockConfig.AssertExpectations(s.T())
	mockQueue.AssertExpectations(s.T())
	mockEvent.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestDelayAsyncQueue() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "queue.default").Return("redis").Times(5)
	mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(cast.ToInt(s.redisResource.GetPort("6379/tcp"))).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	mockQueue, _ := mock.Queue()
	mockQueue.On("GetJobs").Return([]queue.Job{&TestDelayAsyncJob{}}).Once()

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

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
	}).OnQueue("delay").Delay(supporttime.Now().Add(3 * time.Second)).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testDelayAsyncJob)
	time.Sleep(3 * time.Second)
	s.Equal(1, testDelayAsyncJob)

	mockConfig.AssertExpectations(s.T())
	mockQueue.AssertExpectations(s.T())
	mockEvent.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestCustomAsyncQueue() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	mockConfig.On("GetString", "queue.connections.custom.queue", "default").Return("default").Twice()
	mockConfig.On("GetString", "queue.connections.custom.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.custom.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(cast.ToInt(s.redisResource.GetPort("6379/tcp"))).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	mockQueue, _ := mock.Queue()
	mockQueue.On("GetJobs").Return([]queue.Job{&TestCustomAsyncJob{}}).Once()

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

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

	mockConfig.AssertExpectations(s.T())
	mockQueue.AssertExpectations(s.T())
	mockEvent.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestErrorAsyncQueue() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(cast.ToInt(s.redisResource.GetPort("6379/tcp"))).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	mockQueue, _ := mock.Queue()
	mockQueue.On("GetJobs").Return([]queue.Job{&TestErrorAsyncJob{}}).Once()

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

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
	s.Nil(s.app.Job(&TestErrorAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestErrorAsyncQueue"},
		{Type: "int", Value: 1},
	}).OnConnection("redis").OnQueue("error1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testErrorAsyncJob)

	mockConfig.AssertExpectations(s.T())
	mockQueue.AssertExpectations(s.T())
	mockEvent.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestChainAsyncQueue() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "queue.default").Return("redis").Times(5)
	mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	mockConfig.On("GetInt", "database.redis.default.port").Return(cast.ToInt(s.redisResource.GetPort("6379/tcp"))).Twice()
	mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()

	mockQueue, _ := mock.Queue()
	mockQueue.On("GetJobs").Return([]queue.Job{&TestChainAsyncJob{}, &TestChainSyncJob{}}).Once()

	mockEvent, _ := mock.Event()
	mockEvent.On("GetEvents").Return(map[event.Event][]event.Listener{}).Once()

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
				{Type: "string", Value: "TestChainAsyncQueue"},
				{Type: "int", Value: 1},
			},
		},
		{
			Job: &TestChainSyncJob{},
			Args: []queue.Arg{
				{Type: "string", Value: "TestChainSyncQueue"},
				{Type: "int", Value: 1},
			},
		},
	}).OnQueue("chain").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testChainAsyncJob)
	s.Equal(1, testChainSyncJob)

	mockConfig.AssertExpectations(s.T())
	mockQueue.AssertExpectations(s.T())
	mockEvent.AssertExpectations(s.T())
}

type TestAsyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestAsyncJob) Signature() string {
	return "test_async_job"
}

//Handle Execute the job.
func (receiver *TestAsyncJob) Handle(args ...any) error {
	testAsyncJob++

	return nil
}

type TestDelayAsyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestDelayAsyncJob) Signature() string {
	return "test_delay_async_job"
}

//Handle Execute the job.
func (receiver *TestDelayAsyncJob) Handle(args ...any) error {
	testDelayAsyncJob++

	return nil
}

type TestSyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestSyncJob) Signature() string {
	return "test_sync_job"
}

//Handle Execute the job.
func (receiver *TestSyncJob) Handle(args ...any) error {
	testSyncJob++

	return nil
}

type TestCustomAsyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestCustomAsyncJob) Signature() string {
	return "test_async_job"
}

//Handle Execute the job.
func (receiver *TestCustomAsyncJob) Handle(args ...any) error {
	testCustomAsyncJob++

	return nil
}

type TestErrorAsyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestErrorAsyncJob) Signature() string {
	return "test_async_job"
}

//Handle Execute the job.
func (receiver *TestErrorAsyncJob) Handle(args ...any) error {
	testErrorAsyncJob++

	return nil
}

type TestChainAsyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestChainAsyncJob) Signature() string {
	return "test_async_job"
}

//Handle Execute the job.
func (receiver *TestChainAsyncJob) Handle(args ...any) error {
	testChainAsyncJob++

	return nil
}

type TestChainSyncJob struct {
}

//Signature The name and signature of the job.
func (receiver *TestChainSyncJob) Signature() string {
	return "test_sync_job"
}

//Handle Execute the job.
func (receiver *TestChainSyncJob) Handle(args ...any) error {
	testChainSyncJob++

	return nil
}
