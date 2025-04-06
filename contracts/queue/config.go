package queue

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
)

type Config interface {
	Config() config.Config
	Debug() bool
	Default() (connection, queue string, concurrent int)
	Driver(connection string) string
	FailedJobsQuery() db.Query
	Queue(connection, queue string) string
	Via(connection string) any
}
