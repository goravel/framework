package support

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/event"
	queuecontract "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/testing/file"
	"github.com/goravel/framework/testing/mock"
)

func TestGetServer(t *testing.T) {
	initConfig()
	server, err := GetServer("sync", "")
	assert.Nil(t, server)
	assert.Nil(t, err)

	server, err = GetServer("redis", "")
	assert.Nil(t, err)
	assert.NotNil(t, server)

	server, err = GetServer("custom", "")
	assert.Nil(t, server)
	assert.NotNil(t, err)
}

func TestGetQueueName(t *testing.T) {
	var (
		mockConfig *configmocks.Config
	)

	beforeEach := func() {
		mockConfig = mock.Config()
	}

	tests := []struct {
		description     string
		setup           func()
		connection      string
		queue           string
		expectQueueName string
	}{
		{
			description: "success when connection and queue are empty",
			setup: func() {
				mockConfig.On("GetString", "app.name").Return("").Once()
				mockConfig.On("GetString", "queue.default").Return("redis").Once()
				mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("queue").Once()
			},
			expectQueueName: "goravel_queues:queue",
		},
		{
			description: "success when connection and queue aren't empty",
			setup: func() {
				mockConfig.On("GetString", "app.name").Return("app").Once()

			},
			connection:      "redis",
			queue:           "queue",
			expectQueueName: "app_queues:queue",
		},
	}

	for _, test := range tests {
		beforeEach()
		test.setup()
		queueName := GetQueueName(test.connection, test.queue)
		assert.Equal(t, test.expectQueueName, queueName, test.description)
	}
}

func TestGetDriver(t *testing.T) {
	initConfig()
	assert.Equal(t, "sync", getDriver("sync"))
	assert.Equal(t, "redis", getDriver("redis"))
}

func TestGetRedisServer(t *testing.T) {
	initConfig()
	assert.NotNil(t, getRedisServer("default", ""))
}

func TestGetRedisConfig(t *testing.T) {
	initConfig()
	redisConfig, database, queue := getRedisConfig("redis")
	assert.Equal(t, "127.0.0.1:6379", redisConfig)
	assert.Equal(t, 0, database)
	assert.Equal(t, "goravel_queues:default", queue)
}

type TestJob struct {
}

func (receiver *TestJob) Signature() string {
	return "TestName"
}

func (receiver *TestJob) Handle(args ...interface{}) error {
	return nil
}

type TestJobDuplicate struct {
}

func (receiver *TestJobDuplicate) Signature() string {
	return "TestName"
}

func (receiver *TestJobDuplicate) Handle(args ...interface{}) error {
	return nil
}

type TestJobEmpty struct {
}

func (receiver *TestJobEmpty) Signature() string {
	return ""
}

func (receiver *TestJobEmpty) Handle(args ...interface{}) error {
	return nil
}

func TestJobs2Tasks(t *testing.T) {
	_, err := jobs2Tasks([]queuecontract.Job{
		&TestJob{},
	})

	assert.Nil(t, err, "success")

	_, err = jobs2Tasks([]queuecontract.Job{
		&TestJob{},
		&TestJobDuplicate{},
	})

	assert.NotNil(t, err, "Signature duplicate")

	_, err = jobs2Tasks([]queuecontract.Job{
		&TestJobEmpty{},
	})

	assert.NotNil(t, err, "Signature empty")
}

type TestEvent struct {
}

func (receiver *TestEvent) Signature() string {
	return "TestName"
}

func (receiver *TestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestListener struct {
}

func (receiver *TestListener) Signature() string {
	return "TestName"
}

func (receiver *TestListener) Queue(args ...interface{}) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListener) Handle(args ...interface{}) error {
	return nil
}

type TestListenerDuplicate struct {
}

func (receiver *TestListenerDuplicate) Signature() string {
	return "TestName"
}

func (receiver *TestListenerDuplicate) Queue(args ...interface{}) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerDuplicate) Handle(args ...interface{}) error {
	return nil
}

type TestListenerEmpty struct {
}

func (receiver *TestListenerEmpty) Signature() string {
	return ""
}

func (receiver *TestListenerEmpty) Queue(args ...interface{}) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerEmpty) Handle(args ...interface{}) error {
	return nil
}

func TestEvents2Tasks(t *testing.T) {
	_, err := eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListener{},
		},
	})
	assert.Nil(t, err)

	_, err = eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListener{},
			&TestListenerDuplicate{},
		},
	})
	assert.Nil(t, err)

	_, err = eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListenerEmpty{},
		},
	})

	assert.NotNil(t, err)
}

func initConfig() {
	_ = file.CreateEnv()
	configServiceProvider := config.ServiceProvider{}
	configServiceProvider.Register()

	facadesConfig := facades.Config
	facadesConfig.Add("queue", map[string]interface{}{
		"default": facadesConfig.Env("QUEUE_CONNECTION", "redis"),
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"redis": map[string]interface{}{
				"driver":      "redis",
				"connection":  "default",
				"queue":       "default",
				"retry_after": 90,
			},
		},
	})

	facadesConfig.Add("database", map[string]interface{}{
		"redis": map[string]interface{}{
			"default": map[string]interface{}{
				"host":     facadesConfig.Env("REDIS_HOST", "127.0.0.1"),
				"password": facadesConfig.Env("REDIS_PASSWORD", ""),
				"port":     facadesConfig.Env("REDIS_PORT", 6379),
				"database": facadesConfig.Env("REDIS_DB", 0),
			},
		},
	})

	_ = os.Remove(".env")
}
