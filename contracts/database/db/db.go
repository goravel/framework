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
	// dump
	// dumpRawSql
	// Each(callback func(rows []any) error) error
	// Exists() (bool, error)
	// Find(dest any, conds ...any) error
	First(dest any) error
	// decrement
	Delete() (*Result, error)
	Get(dest any) error
	// GroupBy(column string) Query
	// GroupByRaw(query string, args ...any) Query
	// having
	// HavingRaw(query any, args ...any) Query
	// increment
	// inRandomOrder
	Insert(data any) (*Result, error)
	// incrementEach
	// insertGetId
	// Join(table string, on any, args ...any) Query
	// latest
	// LeftJoin(table string, on any, args ...any) Query
	// limit
	// lockForUpdate
	// Max(column string) (any, error)
	// offset
	// OrderBy(column string) Query
	// orderByDesc
	// OrderByRaw(query string, args ...any) Query
	// OrWhere(query any, args ...any) Query
	// OrWhereLike()
	// OrWhereNotLike
	// Pluck(column string, dest any) error
	// RightJoin(table string, on any, args ...any) Query
	// Select(dest any, columns ...string) error
	// SelectRaw(query string, args ...any) (any, error)
	// sharedLock
	// skip
	// take
	Update(data any) (*Result, error)
	// updateOrInsert
	// Value(column string, dest any) error
	// when
	Where(query any, args ...any) Query
	// WhereAll()
	// WhereAny()
	// whereBetween
	// whereColumn
	// whereExists
	// WhereLike()
	// WhereIn()
	// WhereNone()
	// WhereNot()
	// whereNotBetween
	// whereNotIn
	// WhereNotLike()
	// whereNotNull
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
