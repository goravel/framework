package db

import (
	"time"
)

// QueryExecuted is the event fired after every SQL query executed by the
// database, mirroring Laravel's Illuminate\Database\Events\QueryExecuted.
// Listeners are registered through Listen on the DB facade or the package
// level, and receive one event per executed query.
type QueryExecuted struct {
	// Connection is the name of the connection the query was executed on.
	Connection string
	// Sql is the SQL query with placeholders.
	Sql string
	// Bindings are the values bound to the placeholders.
	Bindings []any
	// RawSql is the SQL query with the bindings interpolated.
	RawSql string
	// Time is how long the query took to execute.
	Time time.Duration
	// Error is the error returned by the query, nil when the query succeeded.
	Error error
}
