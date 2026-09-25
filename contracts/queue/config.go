package queue

import (
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
}
