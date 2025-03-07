package db

import (
	"context"
	"database/sql"
)

type DB interface {
	// BeginTransaction Begin a transaction.
	BeginTransaction() (DB, error)
	// Commit Commit the transaction.
	Commit() error
	// Connection Get a database connection by name.
	Connection(name string) DB
	// Rollback Rollback the transaction.
	Rollback() error
	// Table Get a table instance.
	Table(name string) Query
	// Transaction Execute a transaction.
	Transaction(txFunc func(tx DB) error) error
	// WithContext Set the context for the query.
	WithContext(ctx context.Context) DB
}

type Query interface {
	// Count Retrieve the "count" result of the query.
	Count() (int64, error)
	// Chunk Execute a callback over a given chunk size.
	// Chunk(size int, callback func(dest []any) error) error
	// CrossJoin specifying CROSS JOIN conditions for the query.
	CrossJoin(query string, args ...any) Query
	// DoesntExist Determine if no rows exist for the current query.
	DoesntExist() (bool, error)
	// Distinct Force the query to only return distinct results.
	Distinct() Query
	// Delete records from the database.
	Delete() (*Result, error)
	// Each(callback func(rows []any) error) error
	// Exists Determine if any rows exist for the current query.
	Exists() (bool, error)
	// Find Execute a query for a single record by ID.
	Find(dest any, conds ...any) error
	// First finds record that match given conditions.
	First(dest any) error
	// FirstOr finds the first record that matches the given conditions or execute the callback and return its result if no record is found.
	FirstOr(dest any, callback func() error) error
	// FirstOrFail finds the first record that matches the given conditions or throws an error.
	FirstOrFail(dest any) error
	// Decrement the given column's values by the given amounts.
	Decrement(column string, value ...uint64) error
	// Get Retrieve all rows from the database.
	Get(dest any) error
	// GroupBy specifies the group method on the query.
	GroupBy(column ...string) Query
	// Having specifying HAVING conditions for the query.
	Having(query any, args ...any) Query
	// Increment a column's value by a given amount.
	Increment(column string, value ...uint64) error
	// inRandomOrder
	// Insert a new record into the database.
	Insert(data any) (*Result, error)
	// InsertGetId returns the ID of the inserted row, only supported by MySQL and Sqlite
	InsertGetId(data any) (int64, error)
	// Join specifying JOIN conditions for the query.
	Join(query string, args ...any) Query
	// Latest Retrieve the latest record from the database.
	Latest(dest any, column ...string) error
	// LeftJoin specifying LEFT JOIN conditions for the query.
	LeftJoin(query string, args ...any) Query
	// Limit Add a limit to the query.
	Limit(limit uint64) Query
	// LockForUpdate Add a lock for update to the query.
	LockForUpdate() Query
	// Offset Add an "offset" clause to the query.
	Offset(offset uint64) Query
	// OrderBy Add an "order by" clause to the query.
	OrderBy(column string) Query
	// OrderByDesc Add a descending "order by" clause to the query.
	OrderByDesc(column string) Query
	// OrderByRaw Add a raw "order by" clause to the query.
	OrderByRaw(raw string) Query
	// OrWhere add an "or where" clause to the query.
	OrWhere(query any, args ...any) Query
	// OrWhereBetween adds an "or where column between x and y" clause to the query.
	OrWhereBetween(column string, x, y any) Query
	// OrWhereColumn adds an "or where column" clause to the query.
	OrWhereColumn(column1 string, column2 ...string) Query
	// OrWhereIn adds an "or where column in" clause to the query.
	OrWhereIn(column string, args []any) Query
	// OrWhereLike adds an "or where column like" clause to the query.
	OrWhereLike(column string, value string) Query
	// OrWhereNot adds an "or where not" clause to the query.
	OrWhereNot(query any, args ...any) Query
	// OrWhereNotBetween adds an "or where column not between x and y" clause to the query.
	OrWhereNotBetween(column string, x, y any) Query
	// OrWhereNotIn adds an "or where column not in" clause to the query.
	OrWhereNotIn(column string, args []any) Query
	// OrWhereNotLike adds an "or where column not like" clause to the query.
	OrWhereNotLike(column string, value string) Query
	// OrWhereNotNull adds an "or where column is not null" clause to the query.
	OrWhereNotNull(column string) Query
	// OrWhereNull adds an "or where column is null" clause to the query.
	OrWhereNull(column string) Query
	// OrWhereRaw adds a raw "or where" clause to the query.
	OrWhereRaw(raw string, args []any) Query
	// Pluck Get a collection instance containing the values of a given column.
	Pluck(column string, dest any) error
	// RightJoin specifying RIGHT JOIN conditions for the query.
	RightJoin(query string, args ...any) Query
	// Select Set the columns to be selected.
	Select(columns ...string) Query
	// SharedLock Add a shared lock to the query.
	SharedLock() Query
	// ToSql Get the SQL representation of the query.
	ToSql() ToSql
	// ToRawSql Get the raw SQL representation of the query with embedded bindings.
	ToRawSql() ToSql
	// Update records in the database.
	Update(column any, value ...any) (*Result, error)
	// Value Get a single column's value from the first result of a query.
	Value(column string, dest any) error
	// When executes the callback if the condition is true.
	When(condition bool, callback func(query Query) Query) Query
	// Where Add a basic where clause to the query.
	Where(query any, args ...any) Query
	// WhereBetween Add a where between statement to the query.
	WhereBetween(column string, x, y any) Query
	// WhereColumn Add a "where" clause comparing two columns to the query.
	WhereColumn(column1 string, column2 ...string) Query
	// WhereExists Add an exists clause to the query.
	WhereExists(func() Query) Query
	// WhereIn Add a "where in" clause to the query.
	WhereIn(column string, args []any) Query
	// WhereLike Add a "where like" clause to the query.
	WhereLike(column string, value string) Query
	// WhereNot Add a basic "where not" clause to the query.
	WhereNot(query any, args ...any) Query
	// WhereNotBetween Add a where not between statement to the query.
	WhereNotBetween(column string, x, y any) Query
	// WhereNotIn Add a "where not in" clause to the query.
	WhereNotIn(column string, args []any) Query
	// WhereNotLike Add a "where not like" clause to the query.
	WhereNotLike(column string, value string) Query
	// WhereNotNull Add a "where not null" clause to the query.
	WhereNotNull(column string) Query
	// WhereNull Add a "where null" clause to the query.
	WhereNull(column string) Query
	// WhereRaw Add a raw where clause to the query.
	WhereRaw(raw string, args []any) Query
}

type Result struct {
	RowsAffected int64
}

type Builder interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	Query(query string, args ...any) (*sql.Rows, error)
	Select(dest any, query string, args ...any) error
}

type ToSql interface {
	Count() string
	Delete() string
	First() string
	Get() string
	Insert(data any) string
	Pluck(column string, dest any) string
	Update(column any, value ...any) string
}
