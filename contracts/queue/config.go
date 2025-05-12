package queue

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
)

type Config interface {
	Config() config.Config
	Debug() bool
	DefaultConnection() string
	DefaultQueue() string
	DefaultConcurrent() int
	Driver(connection string) string
	FailedJobsQuery() db.Query
	QueueKey(connection, queue string) string
	Via(connection string) any
}
