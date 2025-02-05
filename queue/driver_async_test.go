package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksqueue "github.com/goravel/framework/mocks/queue"
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
	mockConfig *mocksconfig.Config
	mockQueue  *mocksqueue.Queue
}

func TestDriverAsyncTestSuite(t *testing.T) {
	suite.Run(t, new(DriverAsyncTestSuite))
}

func (s *DriverAsyncTestSuite) SetupTest() {
	testAsyncJob = 0
	testDelayAsyncJob = 0
	testCustomAsyncJob = 0
	testErrorAsyncJob = 0
	testChainAsyncJob = 0
	s.mockQueue = mocksqueue.NewQueue(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.app = NewApplication(s.mockConfig)
	s.app.Register([]queue.Job{&TestAsyncJob{}, &TestDelayAsyncJob{}, &TestCustomAsyncJob{}, &TestErrorAsyncJob{}, &TestChainAsyncJob{}})
}

func (s *DriverAsyncTestSuite) TestDefaultAsyncQueue() {
	s.mockConfig.EXPECT().GetString("queue.default").Return("async").Times(3)
	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(2)
	s.mockConfig.EXPECT().GetString("queue.connections.async.queue", "default").Return("default").Twice()
	s.mockConfig.EXPECT().GetString("queue.connections.async.driver").Return("async").Twice()
	s.mockConfig.EXPECT().GetInt("queue.connections.async.size", 100).Return(10).Twice()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker().Run())
	}(ctx)
	time.Sleep(1 * time.Second)
	s.Nil(s.app.Job(&TestAsyncJob{}, []any{"TestDefaultAsyncQueue", 1}).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testAsyncJob)
}

// func (s *DriverAsyncTestSuite) TestDelayAsyncQueue() {
// 	s.mockConfig.EXPECT().GetString("queue.default").Return("async").Times(3)
// 	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(3)
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.queue", "default").Return("default").Once()
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.driver").Return("async").Twice()
// 	s.mockConfig.EXPECT().GetInt("queue.connections.async.size", 100).Return(10).Twice()

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	go func(ctx context.Context) {
// 		worker := s.app.Worker(queue.Args{
// 			Queue: "delay",
// 		})
// 		s.Nil(worker.Run())
// 	}(ctx)
// 	time.Sleep(1 * time.Second)
// 	s.Nil(s.app.Job(&TestDelayAsyncJob{}, []any{"TestDelayAsyncQueue", 1}).OnQueue("delay").Delay(time.Now().Add(3 * time.Second)).Dispatch())
// 	time.Sleep(2 * time.Second)
// 	s.Equal(0, testDelayAsyncJob)
// 	time.Sleep(3 * time.Second)
// 	s.Equal(1, testDelayAsyncJob)
// }

// func (s *DriverAsyncTestSuite) TestCustomAsyncQueue() {
// 	s.mockConfig.EXPECT().GetString("queue.default").Return("custom").Times(3)
// 	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(3)
// 	s.mockConfig.EXPECT().GetString("queue.connections.custom.queue", "default").Return("default").Once()
// 	s.mockConfig.EXPECT().GetString("queue.connections.custom.driver").Return("async").Times(2)
// 	s.mockConfig.EXPECT().GetInt("queue.connections.custom.size", 100).Return(10).Twice()

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	go func(ctx context.Context) {
// 		worker := s.app.Worker(queue.Args{
// 			Connection: "custom",
// 			Queue:      "custom1",
// 			Concurrent: 2,
// 		})
// 		s.Nil(worker.Run())
// 	}(ctx)
// 	time.Sleep(1 * time.Second)
// 	s.Nil(s.app.Job(&TestCustomAsyncJob{}, []any{"TestCustomAsyncQueue", 1}).OnConnection("custom").OnQueue("custom1").Dispatch())
// 	time.Sleep(2 * time.Second)
// 	s.Equal(1, testCustomAsyncJob)
// }

// func (s *DriverAsyncTestSuite) TestErrorAsyncQueue() {
// 	s.mockConfig.EXPECT().GetString("queue.default").Return("async").Times(3)
// 	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(3)
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.queue", "default").Return("default").Once()
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.driver").Return("async").Once()
// 	s.mockConfig.EXPECT().GetInt("queue.connections.async.size", 100).Return(10).Once()
// 	s.mockConfig.EXPECT().GetString("queue.connections.redis.driver").Return("").Twice()

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	go func(ctx context.Context) {
// 		worker := s.app.Worker(queue.Args{
// 			Queue: "error",
// 		})
// 		s.Nil(worker.Run())
// 	}(ctx)
// 	time.Sleep(1 * time.Second)
// 	s.Error(s.app.Job(&TestErrorAsyncJob{}, []any{"TestErrorAsyncQueue", 1}).OnConnection("redis").OnQueue("error1").Dispatch())
// 	time.Sleep(2 * time.Second)
// 	s.Equal(0, testErrorAsyncJob)
// }

// func (s *DriverAsyncTestSuite) TestChainAsyncQueue() {
// 	s.mockConfig.EXPECT().GetString("queue.default").Return("async").Times(3)
// 	s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Times(3)
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.queue", "default").Return("default").Once()
// 	s.mockConfig.EXPECT().GetString("queue.connections.async.driver").Return("async").Twice()
// 	s.mockConfig.EXPECT().GetInt("queue.connections.async.size", 100).Return(10).Twice()

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	go func(ctx context.Context) {
// 		worker := s.app.Worker(queue.Args{
// 			Queue: "chain",
// 		})
// 		s.Nil(worker.Run())
// 	}(ctx)

// 	time.Sleep(1 * time.Second)
// 	s.Nil(s.app.Chain([]queue.Jobs{
// 		{
// 			Job:  &TestChainAsyncJob{},
// 			Args: []any{"TestChainAsyncJob", 1},
// 		},
// 		{
// 			Job:  &TestAsyncJob{},
// 			Args: []any{"TestAsyncJob", 1},
// 		},
// 	}).OnQueue("chain").Dispatch())

// 	time.Sleep(3 * time.Second)
// 	s.Equal(1, testChainAsyncJob)
// 	s.Equal(1, testAsyncJob)
// }

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
