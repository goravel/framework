package db

import "database/sql"

type DB interface {
	Table(name string) Query
}

type Query interface {
	// Avg(column string) (any, error)
	// Count(dest *int64) error
	// Chunk(size int, callback func(rows []any) error) error
	// CrossJoin(table string, on any, args ...any) Query
	// DoesntExist() (bool, error)
	// Distinct() Query
	// Each(callback func(rows []any) error) error
	// Exists() (bool, error)
	// Find(dest any, conds ...any) error
	First(dest any) error
	Delete() (*Result, error)
	Get(dest any) error
	// GroupBy(column string) Query
	GroupByRaw(query string, args ...any) Query
	// HavingRaw(query any, args ...any) Query
	// Join(table string, on any, args ...any) Query
	// LeftJoin(table string, on any, args ...any) Query
	// Max(column string) (any, error)
	// OrderBy(column string) Query
	// OrderByRaw(query string, args ...any) Query
	// OrWhere(query any, args ...any) Query
	// OrWhereLike()
	// OrWhereNotLike
	Pluck(column string, dest any) error
	Insert(data any) (*Result, error)
	// RightJoin(table string, on any, args ...any) Query
	// Select(dest any, columns ...string) error
	// SelectRaw(query string, args ...any) (any, error)
	Update(data any) (*Result, error)
	// Value(column string, dest any) error
	Where(query any, args ...any) Query
	// WhereAll()
	// WhereAny()
	// WhereLike()
	// WhereNone()
	// WhereNot()
	// WhereNotLike()
	// WhereNull(column string) Query
	// WhereRaw(query string, args ...any) Query
}

type Result struct {
	RowsAffected int64
}

type Builder interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
}
