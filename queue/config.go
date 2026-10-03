package queue

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/queue"
)

// defaultReceiveTimeout is the fallback timeout for a single blocking Receive
// call made by the worker. It can be overridden per connection via
// `queue.connections.<connection>.timeout`.
const defaultReceiveTimeout = queue.DefaultReceiveTimeout

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
	defaultQueue := splitQueueNames(configuredQueue(config, defaultConnection))[0]
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

// Timeout resolves the blocking receive timeout for the given connection from
// `queue.connections.<connection>.timeout`, falling back to the default when
// the value is missing, non-positive or unparsable.
func (r *Config) Timeout(connection string) time.Duration {
	value := r.Get(fmt.Sprintf("queue.connections.%s.timeout", connection))

	switch timeout := value.(type) {
	case time.Duration:
		return durationToTimeout(timeout)
	case string:
		if duration, err := time.ParseDuration(timeout); err == nil {
			return durationToTimeout(duration)
		}
		if seconds, err := strconv.ParseFloat(timeout, 64); err == nil {
			return floatToTimeout(seconds)
		}
		return defaultReceiveTimeout
	case json.Number:
		if seconds, err := timeout.Float64(); err == nil {
			return floatToTimeout(seconds)
		}
		return defaultReceiveTimeout
	default:
		if timeout, ok := numericTimeout(value); ok {
			return timeout
		}
		return defaultReceiveTimeout
	}
}

// numericTimeout normalizes any Go integer or float kind into a timeout. It
// reports false when the value is not numeric so the caller can apply the
// default. Values are widened to int64/uint64/float64 before the positivity
// and overflow guards run, so the conversion is 32-bit safe.
func numericTimeout(value any) (time.Duration, bool) {
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return secondsToTimeout(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintToTimeout(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return floatToTimeout(v.Float()), true
	}

	return 0, false
}

// secondsToTimeout converts a configured number of seconds into a duration,
// falling back to the default when the value is non-positive or too large to be
// represented as a duration without overflowing.
func secondsToTimeout(seconds int64) time.Duration {
	if seconds <= 0 || seconds > math.MaxInt64/int64(time.Second) {
		return defaultReceiveTimeout
	}
	return time.Duration(seconds) * time.Second
}

// uintToTimeout converts a configured number of unsigned seconds into a
// duration, falling back to the default when the value is zero or too large to
// be represented as a duration without overflowing.
func uintToTimeout(seconds uint64) time.Duration {
	if seconds == 0 || seconds > uint64(math.MaxInt64)/uint64(time.Second) {
		return defaultReceiveTimeout
	}
	return time.Duration(seconds) * time.Second
}

// floatToTimeout converts a configured number of seconds into a duration,
// falling back to the default when the value is non-positive, NaN or too large
// to be represented as a duration without overflowing.
func floatToTimeout(seconds float64) time.Duration {
	nanos := seconds * float64(time.Second)
	if math.IsNaN(seconds) || seconds <= 0 || nanos >= float64(math.MaxInt64) {
		return defaultReceiveTimeout
	}
	return durationToTimeout(time.Duration(nanos))
}

// durationToTimeout returns the duration as-is when positive, otherwise the
// default.
func durationToTimeout(duration time.Duration) time.Duration {
	if duration <= 0 {
		return defaultReceiveTimeout
	}
	return duration
}

func configuredQueue(config contractsconfig.Config, connection string) string {
	return config.GetString(fmt.Sprintf("queue.connections.%s.queue", connection), "default")
}

// splitQueueNames normalizes a comma-separated worker queue list. It always
// returns at least one name, falls back to "default", and preserves order and
// duplicate names to match Laravel's semantics.
func splitQueueNames(queue string) []string {
	queueNames := make([]string, 0, strings.Count(queue, ",")+1)
	for _, queueName := range strings.Split(queue, ",") {
		queueName = strings.TrimSpace(queueName)
		if queueName != "" {
			queueNames = append(queueNames, queueName)
		}
	}

	if len(queueNames) == 0 {
		return []string{"default"}
	}

	return queueNames
}
