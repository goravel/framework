package queue

import "github.com/goravel/framework/contracts/database/db"

type Config interface {
	Debug() bool
	DefaultConnection() string
	Driver(connection string) string
	FailedJobsQuery() db.Query
	Queue(connection, queue string) string
	Size(connection string) int
	Via(connection string) any
}
