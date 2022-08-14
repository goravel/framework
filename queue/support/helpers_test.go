package support

import (
	"os"
	"testing"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/events"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/facades"
	testing2 "github.com/goravel/framework/support/testing"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "default", queue)
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
	_, err := jobs2Tasks([]queue.Job{
		&TestJob{},
	})

	assert.Nil(t, err, "success")

	_, err = jobs2Tasks([]queue.Job{
		&TestJob{},
		&TestJobDuplicate{},
	})

	assert.NotNil(t, err, "Signature duplicate")

	_, err = jobs2Tasks([]queue.Job{
		&TestJobEmpty{},
	})

	assert.NotNil(t, err, "Signature empty")
}

type TestEvent struct {
}

func (receiver *TestEvent) Signature() string {
	return "TestName"
}

func (receiver *TestEvent) Handle(args []events.Arg) ([]events.Arg, error) {
	return args, nil
}

type TestListener struct {
}

func (receiver *TestListener) Signature() string {
	return "TestName"
}

func (receiver *TestListener) Queue(args ...interface{}) events.Queue {
	return events.Queue{
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

func (receiver *TestListenerDuplicate) Queue(args ...interface{}) events.Queue {
	return events.Queue{
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

func (receiver *TestListenerEmpty) Queue(args ...interface{}) events.Queue {
	return events.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerEmpty) Handle(args ...interface{}) error {
	return nil
}

func TestEvents2Tasks(t *testing.T) {
	_, err := events2Tasks(map[events.Event][]events.Listener{
		&TestEvent{}: {
			&TestListener{},
		},
	})

	assert.Nil(t, err)

	_, err = events2Tasks(map[events.Event][]events.Listener{
		&TestEvent{}: {
			&TestListener{},
			&TestListenerDuplicate{},
		},
	})

	assert.NotNil(t, err)

	_, err = events2Tasks(map[events.Event][]events.Listener{
		&TestEvent{}: {
			&TestListenerEmpty{},
		},
	})

	assert.NotNil(t, err)
}

func initConfig() {
	testing2.CreateEnv()
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

	os.Remove(".env")
}
