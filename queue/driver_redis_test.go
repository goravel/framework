package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	configmock "github.com/goravel/framework/mocks/config"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	queuemock "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/support/carbon"
	testingdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

var (
	testRedisJob       = 0
	testDelayRedisJob  = 0
	testCustomRedisJob = 0
	testErrorRedisJob  = 0
	testChainRedisJob  = 0
)

type DriverRedisTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *configmock.Config
	mockQueue  *queuemock.Queue
	port       int
}

func TestDriverRedisTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	redisDocker := testingdocker.NewRedis()
	assert.Nil(t, redisDocker.Build())

	suite.Run(t, &DriverRedisTestSuite{
		port: redisDocker.Config().Port,
	})

	assert.Nil(t, redisDocker.Stop())
}

func (s *DriverRedisTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.mockQueue = &queuemock.Queue{}
	s.app = NewApplication(s.mockConfig)

	JobRegistry = new(sync.Map)
	testRedisJob = 0
	testDelayRedisJob = 0
	testCustomRedisJob = 0
	testErrorRedisJob = 0
	testChainRedisJob = 0

	mockOrm := &ormmock.Orm{}
	mockQuery := &ormmock.Query{}
	mockOrm.On("Connection", "database").Return(mockOrm)
	mockOrm.On("Query").Return(mockQuery)
	mockQuery.On("Table", "failed_jobs").Return(mockQuery)

	OrmFacade = mockOrm

	s.Nil(s.app.Register([]queue.Job{&TestRedisJob{}, &TestDelayRedisJob{}, &TestCustomRedisJob{}, &TestErrorRedisJob{}, &TestChainRedisJob{}}))
}

func (s *DriverRedisTestSuite) TestDefaultRedisQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Times(2)
	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
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
	s.Nil(s.app.Job(&TestRedisJob{}, []queue.Arg{
		{Type: "string", Value: "TestDefaultRedisQueue"},
		{Type: "int", Value: 1},
	}).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testRedisJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverRedisTestSuite) TestDelayRedisQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Times(2)
	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
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
	s.Nil(s.app.Job(&TestDelayRedisJob{}, []queue.Arg{
		{Type: "string", Value: "TestDelayRedisQueue"},
		{Type: "int", Value: 1},
	}).OnQueue("delay").Delay(carbon.Now().AddSeconds(3)).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testDelayRedisJob)
	time.Sleep(3 * time.Second)
	s.Equal(1, testDelayRedisJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverRedisTestSuite) TestCustomRedisQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("custom").Times(7)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.custom.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.custom.driver").Return("redis").Times(4)
	s.mockConfig.On("GetString", "queue.connections.custom.connection").Return("default").Times(3)
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Times(3)
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Times(3)
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Times(3)
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Times(3)
	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
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
	s.Nil(s.app.Job(&TestCustomRedisJob{}, []queue.Arg{
		{Type: "string", Value: "TestCustomRedisQueue"},
		{Type: "int", Value: 1},
	}).OnConnection("custom").OnQueue("custom1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testCustomRedisJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverRedisTestSuite) TestErrorRedisQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(7)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(4)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Times(3)
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Times(3)
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Times(3)
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Times(3)
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Times(3)
	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
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
	s.Nil(s.app.Job(&TestErrorRedisJob{}, []queue.Arg{
		{Type: "string", Value: "TestErrorRedisQueue"},
		{Type: "int", Value: 1},
	}).OnConnection("redis").OnQueue("error1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testErrorRedisJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockQueue.AssertExpectations(s.T())
}

func (s *DriverRedisTestSuite) TestChainRedisQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(6)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Times(2)
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Times(2)
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Times(2)
	s.mockConfig.On("GetString", "queue.failed.connection").Return("database").Once()
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
			Job: &TestChainRedisJob{},
			Args: []queue.Arg{
				{Type: "string", Value: "TestChainRedisJob"},
				{Type: "int", Value: 1},
			},
		},
		{
			Job: &TestRedisJob{},
			Args: []queue.Arg{
				{Type: "string", Value: "TestRedisJob"},
				{Type: "int", Value: 1},
			},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainRedisJob)
	s.Equal(1, testRedisJob)

	s.mockConfig.AssertExpectations(s.T())
}

type TestRedisJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestRedisJob) Signature() string {
	return "test_redis_job"
}

// Handle Execute the job.
func (receiver *TestRedisJob) Handle(args ...any) error {
	testRedisJob++

	return nil
}

type TestDelayRedisJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestDelayRedisJob) Signature() string {
	return "test_delay_redis_job"
}

// Handle Execute the job.
func (receiver *TestDelayRedisJob) Handle(args ...any) error {
	testDelayRedisJob++

	return nil
}

type TestCustomRedisJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestCustomRedisJob) Signature() string {
	return "test_custom_redis_job"
}

// Handle Execute the job.
func (receiver *TestCustomRedisJob) Handle(args ...any) error {
	testCustomRedisJob++

	return nil
}

type TestErrorRedisJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestErrorRedisJob) Signature() string {
	return "test_error_redis_job"
}

// Handle Execute the job.
func (receiver *TestErrorRedisJob) Handle(args ...any) error {
	testErrorRedisJob++

	return nil
}

type TestChainRedisJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestChainRedisJob) Signature() string {
	return "test_chain_redis_job"
}

// Handle Execute the job.
func (receiver *TestChainRedisJob) Handle(args ...any) error {
	testChainRedisJob++

	return nil
}
