package queue

import (
	"time"

	"github.com/goravel/framework/contracts/config"
)

type Config interface {
	config.Config
	Debug() bool
	DefaultConnection() string
	// DefaultQueue returns the first valid queue name configured for the default connection.
	// This is the queue used when dispatching a job without an explicit queue.
	DefaultQueue() string
	DefaultConcurrent() int
	Driver(connection string) string
	FailedDatabase() string
	FailedTable() string
	Via(connection string) any
	// Timeout returns the blocking receive timeout for the given connection,
	// resolved from queue.connections.<connection>.timeout (an integer number
	// of seconds or a duration string like "5s"), falling back to the default
	// when missing, non-positive or unparsable.
	Timeout(connection string) time.Duration
}
