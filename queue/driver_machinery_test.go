// TODO: Will be removed in v1.17

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
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
	testingdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

var (
	testMachineryJob               = 0
	testMachineryJobOfDisableDebug = 0
	testDelayMachineryJob          = 0
	testCustomMachineryJob         = 0
	testErrorMachineryJob          = 0
	testChainMachineryJob          = 0
	testChainMachineryJobError     = 0
)

type MachineryTestSuite struct {
	suite.Suite
	app        *Application
	mockConfig *mocksconfig.Config
	mockLog    *mockslog.Log
	machinery  *Machinery
	port       int
}

func TestMachineryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	redisDocker := testingdocker.NewRedis()
	assert.Nil(t, redisDocker.Build())

	suite.Run(t, &MachineryTestSuite{
		port: redisDocker.Config().Port,
	})

	assert.Nil(t, redisDocker.Shutdown())
}

func (s *MachineryTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockLog = mockslog.NewLog(s.T())
	s.app = NewApplication(s.mockConfig)
}

func (s *MachineryTestSuite) TestServer() {
	tests := []struct {
		name         string
		connection   string
		queue        string
		setup        func()
		expectServer bool
		expectErr    bool
	}{
		{
			name:       "redis",
			connection: "redis",
			setup: func() {
				s.mockConfig.On("GetString", "queue.connections.redis.connection").Return("default").Once()
				s.mockConfig.On("GetString", "database.redis.default.host").Return("127.0.0.1").Once()
				s.mockConfig.On("GetString", "database.redis.default.password").Return("").Once()
				s.mockConfig.On("GetInt", "database.redis.default.port").Return(6379).Once()
				s.mockConfig.On("GetInt", "database.redis.default.database").Return(0).Once()
				s.mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("default").Once()
				s.mockConfig.On("GetString", "app.name").Return("goravel").Once()
				s.mockConfig.On("GetBool", "app.debug").Return(true).Once()
			},
			expectServer: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.machinery = NewMachinery(test.connection, NewConfig(s.mockConfig), s.mockLog)
			test.setup()
			server := s.machinery.server(test.queue)
			s.Equal(test.expectServer, server != nil)
		})
	}
}

func (s *MachineryTestSuite) TestDefaultAsyncQueue_EnableDebug() {
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
	s.app.Register([]queue.Job{&TestMachineryJob{}})

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
	s.Nil(s.app.Job(&TestMachineryJob{}, []any{
		"TestDefaultAsyncQueue_EnableDebug",
		1,
	}).OnQueue("debug").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testMachineryJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestDefaultAsyncQueue_DisableDebug() {
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
	s.app.Register([]queue.Job{&TestMachineryJobOfDisableDebug{}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		s.Nil(s.app.Worker().Run())

		for range ctx.Done() {
			return
		}
	}(ctx)
	time.Sleep(2 * time.Second)
	s.Nil(s.app.Job(&TestMachineryJobOfDisableDebug{}, []any{
		"TestDefaultAsyncQueue_DisableDebug",
		1,
	}).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testMachineryJobOfDisableDebug)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestDelayAsyncQueue() {
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
	s.app.Register([]queue.Job{&TestDelayMachineryJob{}})

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
	s.Nil(s.app.Job(&TestDelayMachineryJob{}, []any{
		"TestDelayAsyncQueue",
		1,
	}).OnQueue("delay").Delay(carbon.Now().AddSeconds(3).StdTime()).Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testDelayMachineryJob)
	time.Sleep(3 * time.Second)
	s.Equal(1, testDelayMachineryJob)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestCustomAsyncQueue() {
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
	s.app.Register([]queue.Job{&TestCustomMachineryJob{}})

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
	s.Nil(s.app.Job(&TestCustomMachineryJob{}, []any{
		"TestCustomAsyncQueue",
		1,
	}).OnConnection("custom").OnQueue("custom1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(1, testCustomMachineryJob)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestErrorAsyncQueue() {
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
	s.app.Register([]queue.Job{&TestErrorMachineryJob{}})

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
	s.Nil(s.app.Job(&TestErrorMachineryJob{}, []any{
		"TestErrorAsyncQueue",
		1,
	}).OnConnection("redis").OnQueue("error1").Dispatch())
	time.Sleep(2 * time.Second)
	s.Equal(0, testErrorMachineryJob)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestChainAsyncQueue() {
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
	s.app.Register([]queue.Job{&TestChainMachineryJob{}, &TestChainSyncJob{}})

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
			Job: &TestChainMachineryJob{},
			Args: []any{
				"TestChainAsyncQueue",
				1,
			},
		},
		{
			Job: &TestChainSyncJob{},
			Args: []any{
				"TestChainSyncQueue",
				1,
			},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainMachineryJob)
	s.Equal(1, testChainSyncJob)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *MachineryTestSuite) TestChainAsyncQueue_Error() {
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
	s.app.Register([]queue.Job{&TestChainMachineryJob{}, &TestChainSyncJob{}})

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
			Job:  &TestChainMachineryJob{},
			Args: []any{true},
		},
		{
			Job:  &TestChainSyncJob{},
			Args: []any{},
		},
	}).OnQueue("chain").Dispatch())

	time.Sleep(2 * time.Second)
	s.Equal(1, testChainMachineryJobError)
	s.Equal(0, testChainSyncJob)

	s.mockConfig.AssertExpectations(s.T())
	s.mockLog.AssertExpectations(s.T())
}

type TestMachineryJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestMachineryJob) Signature() string {
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestMachineryJob) Handle(args ...any) error {
	testMachineryJob++

	return nil
}

type TestMachineryJobOfDisableDebug struct {
}

// Signature The name and signature of the job.
func (receiver *TestMachineryJobOfDisableDebug) Signature() string {
	return "test_async_job_of_disable_debug"
}

// Handle Execute the job.
func (receiver *TestMachineryJobOfDisableDebug) Handle(args ...any) error {
	testMachineryJobOfDisableDebug++

	return nil
}

type TestDelayMachineryJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestDelayMachineryJob) Signature() string {
	return "test_delay_async_job"
}

// Handle Execute the job.
func (receiver *TestDelayMachineryJob) Handle(args ...any) error {
	testDelayMachineryJob++

	return nil
}

type TestCustomMachineryJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestCustomMachineryJob) Signature() string {
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestCustomMachineryJob) Handle(args ...any) error {
	testCustomMachineryJob++

	return nil
}

type TestErrorMachineryJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestErrorMachineryJob) Signature() string {
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestErrorMachineryJob) Handle(args ...any) error {
	testErrorMachineryJob++

	return nil
}

type TestChainMachineryJob struct {
}

// Signature The name and signature of the job.
func (receiver *TestChainMachineryJob) Signature() string {
	return "test_async_job"
}

// Handle Execute the job.
func (receiver *TestChainMachineryJob) Handle(args ...any) error {
	if len(args) > 0 && cast.ToBool(args[0]) {
		testChainMachineryJobError++

		return errors.New("error")
	}

	testChainMachineryJob++

	return nil
}
