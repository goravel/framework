package queue

import (
	"time"

	"github.com/goravel/framework/contracts/config"
)

// DefaultReceiveTimeout is the fallback timeout for a single blocking receive
// call made by the queue worker when the connection's configured timeout is
// missing, non-positive or unparsable.
const DefaultReceiveTimeout = 5 * time.Second

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
	// resolved from `queue.connections.<connection>.timeout` (a number of
	// seconds, integer or fractional, or a duration string like "5s"), falling
	// back to DefaultReceiveTimeout when missing, non-positive or unparsable.
	Timeout(connection string) time.Duration
}
