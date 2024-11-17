package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/queue"
	configmock "github.com/goravel/framework/mocks/config"
	logmock "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
	testingdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

var (
	testSyncJob                = 0
	testAsyncJob               = 0
	testAsyncJobOfDisableDebug = 0
	testDelayAsyncJob          = 0
	testCustomAsyncJob         = 0
	testErrorAsyncJob          = 0
	testChainAsyncJob          = 0
	testChainSyncJob           = 0
	testChainAsyncJobError     = 0
	testChainSyncJobError      = 0
)

type QueueTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *configmock.Config
	mockLog    *logmock.Log
	port       int
}

func TestQueueTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	redisDocker := testingdocker.NewRedis()
	assert.Nil(t, redisDocker.Build())

	suite.Run(t, &QueueTestSuite{
		port: redisDocker.Config().Port,
	})

	assert.Nil(t, redisDocker.Stop())
}

func (s *QueueTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.mockLog = &logmock.Log{}
	s.app = NewApplication(s.mockConfig, s.mockLog)
}

func (s *QueueTestSuite) TestSyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Once()
	s.Nil(s.app.Job(&TestSyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestSyncQueue"},
		{Type: "int", Value: 1},
	}).DispatchSync())
	s.Equal(1, testSyncJob)
}

func (s *QueueTestSuite) TestDefaultAsyncQueue_EnableDebug() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(true).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.mockLog.On("Infof", "Launching a worker with the following settings:").Once()
	s.mockLog.On("Infof", "- Broker: %s", "://").Once()
	s.mockLog.On("Infof", "- DefaultQueue: %s", "goravel_queues:debug").Once()
	s.mockLog.On("Infof", "- ResultBackend: %s", "://").Once()
	s.mockLog.On("Info", "[*] Waiting for messages. To exit press CTRL+C").Once()
	s.mockLog.On("Debugf", "Received new message: %s", mock.Anything).Once()
	s.mockLog.On("Debugf", "Processed task %s. Results = %s", mock.Anything, mock.Anything).Once()
	s.app.jobs = []queue.Job{&TestAsyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
			Queue: "debug",
		}).Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestAsyncJob{}, []queue.Arg{
		{Type: "string", Value: "TestDefaultAsyncQueue_EnableDebug"},
		{Type: "int", Value: 1},
	}).OnQueue("debug").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestDefaultAsyncQueue_DisableDebug() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(3)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.app.jobs = []queue.Job{&TestAsyncJobOfDisableDebug{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker().Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestAsyncJobOfDisableDebug{}, []queue.Arg{
		{Type: "string", Value: "TestDefaultAsyncQueue_DisableDebug"},
		{Type: "int", Value: 1},
	}).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testAsyncJobOfDisableDebug)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestDelayAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(2)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.app.jobs = []queue.Job{&TestDelayAsyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
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
	}).OnQueue("delay").Delay(carbon.Now().AddSeconds(3).StdTime()).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testDelayAsyncJob)
	time.Sleep(3 * time.Second)
	s.Equal(1, testDelayAsyncJob)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestCustomAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.custom.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.custom.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.custom.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.app.jobs = []queue.Job{&TestCustomAsyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
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
}

func (s *QueueTestSuite) TestErrorAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Twice()
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.app.jobs = []queue.Job{&TestErrorAsyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
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

	s.mockConfig.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestChainAsyncQueue() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(2)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.app.jobs = []queue.Job{&TestChainAsyncJob{}, &TestChainSyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
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

	s.mockConfig.AssertExpectations(s.T())
}

func (s *QueueTestSuite) TestChainAsyncQueue_Error() {
	s.mockConfig.On("GetString", "queue.default").Return("redis").Times(2)
	s.mockConfig.On("GetString", "app.name").Return("goravel").Times(4)
	s.mockConfig.On("GetBool", "app.debug").Return(false).Times(2)
	s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Twice()
	s.mockConfig.On("GetString", "queue.connections.redis.driver").Return("redis").Times(3)
	s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Twice()
	s.mockConfig.On("GetString", "database.redis.default.host").Return("localhost").Twice()
	s.mockConfig.On("GetString", "database.redis.default.password").Return("").Twice()
	s.mockConfig.On("GetInt", "database.redis.default.port").Return(s.port).Twice()
	s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Twice()
	s.mockLog.On("Errorf", "Failed processing task %s. Error = %v", mock.Anything, errors.New("error")).Once()
	s.app.jobs = []queue.Job{&TestChainAsyncJob{}, &TestChainSyncJob{}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker(queue.Args{
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
				{Type: "bool", Value: true},
			},
		},
		{
			Job:  &TestChainSyncJob{},
			Args: []queue.Arg{},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainAsyncJobError)
	s.Equal(0, testChainSyncJobError)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
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

type TestAsyncJobOfDisableDebug struct {
}

// Signature The name and signature of the job.
func (receiver *TestAsyncJobOfDisableDebug) Signature() string {
	return "test_async_job_of_disable_debug"
}

// Handle Execute the job.
func (receiver *TestAsyncJobOfDisableDebug) Handle(args ...any) error {
	testAsyncJobOfDisableDebug++

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

type TestSyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestSyncJob) Signature() string {
	return "test_sync_job"
}

// Handle Execute the job.
func (receiver *TestSyncJob) Handle(args ...any) error {
	testSyncJob++

	return nil
}

type TestCustomAsyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestCustomAsyncJob) Signature() string {
	return "test_async_job"
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
	return "test_async_job"
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
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestChainAsyncJob) Handle(args ...any) error {
	if len(args) > 0 && cast.ToBool(args[0]) {
		testChainAsyncJobError++

		return errors.New("error")
	}

	testChainAsyncJob++

	return nil
}

type TestChainSyncJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestChainSyncJob) Signature() string {
	return "test_sync_job"
}

// Handle Execute the job.
func (receiver *TestChainSyncJob) Handle(args ...any) error {
	testChainSyncJob++

	return nil
}
