package queue

import (
	"fmt"
	"time"

	contractsconfig "github.com/goravel/framework/contracts/config"
)

// defaultReceiveTimeout is the fallback timeout, in seconds, for a single
// blocking Receive call made by the worker. It can be overridden per
// connection via queue.connections.<connection>.timeout.
const defaultReceiveTimeout = 5

type Config struct {
	contractsconfig.Config

	appName           string
	defaultConnection string
	defaultQueue      string
	failedDatabase    string
	failedTable       string
	defaultConcurrent int
	debug             bool
}

func NewConfig(config contractsconfig.Config) *Config {
	defaultConnection := config.GetString("queue.default")
	defaultQueue := config.GetString(fmt.Sprintf("queue.connections.%s.queue", defaultConnection), "default")
	defaultConcurrent := max(config.GetInt(fmt.Sprintf("queue.connections.%s.concurrent", defaultConnection), 1), 1)

	c := &Config{
		Config: config,

		appName:           config.GetString("app.name", "goravel"),
		debug:             config.GetBool("app.debug"),
		defaultConnection: defaultConnection,
		defaultQueue:      defaultQueue,
		defaultConcurrent: defaultConcurrent,
		failedDatabase:    config.GetString("queue.failed.database"),
		failedTable:       config.GetString("queue.failed.table"),
	}

	return c
}

func (r *Config) Debug() bool {
	return r.debug
}

func (r *Config) DefaultConnection() string {
	return r.defaultConnection
}

func (r *Config) DefaultQueue() string {
	return r.defaultQueue
}

func (r *Config) DefaultConcurrent() int {
	return r.defaultConcurrent
}

func (r *Config) Driver(connection string) string {
	return r.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func (r *Config) FailedDatabase() string {
	return r.failedDatabase
}

func (r *Config) FailedTable() string {
	return r.failedTable
}

func (r *Config) Via(connection string) any {
	return r.Get(fmt.Sprintf("queue.connections.%s.via", connection))
}

func (r *Config) Timeout(connection string) time.Duration {
	switch timeout := r.Get(fmt.Sprintf("queue.connections.%s.timeout", connection)).(type) {
	case int:
		return secondsToTimeout(timeout)
	case int64:
		return secondsToTimeout(int(timeout))
	case float64:
		return secondsToTimeout(int(timeout))
	case string:
		duration, err := time.ParseDuration(timeout)
		if err != nil {
			return defaultReceiveTimeout * time.Second
		}
		return durationToTimeout(duration)
	default:
		return defaultReceiveTimeout * time.Second
	}
}

// secondsToTimeout converts a configured number of seconds into a duration,
// falling back to the default when the value is missing or non-positive.
func secondsToTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultReceiveTimeout * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// durationToTimeout returns the duration as-is when positive, otherwise the
// default.
func durationToTimeout(duration time.Duration) time.Duration {
	if duration <= 0 {
		return defaultReceiveTimeout * time.Second
	}
	return duration
}
