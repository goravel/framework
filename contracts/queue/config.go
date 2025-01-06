package queue

import "github.com/goravel/framework/contracts/database/orm"

type Config interface {
	Debug() bool
	DefaultConnection() string
	Driver(connection string) string
	FailedJobsQuery() orm.Query
	Queue(connection, queue string) string
	Redis(queueConnection string) (dsn string, database int, queue string) // TODO: Will be removed in v1.17
	Size(connection string) int
	Via(connection string) any
}
