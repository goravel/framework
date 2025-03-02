package db

import (
	"context"
	"database/sql"
)

type DB interface {
	// BeginTransaction() Query
	Connection(name string) DB
	Table(name string) Query
	// Transaction(txFunc func(tx Query) error) error
	WithContext(ctx context.Context) DB
}

type Query interface {
	// Avg(column string) (any, error)
	// commit
	// Count(dest *int64) error
	// Chunk(size int, callback func(rows []any) error) error
	// CrossJoin(table string, on any, args ...any) Query
	// DoesntExist() (bool, error)
	// Distinct() Query
	// dump
	// dumpRawSql
	// Each(callback func(rows []any) error) error
	// Exists() (bool, error)
	Find(dest any, conds ...any) error
	First(dest any) error
	// firstOrFail
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
	OrderBy(column string) Query
	OrderByDesc(column string) Query
	OrderByRaw(raw string) Query
	OrWhere(query any, args ...any) Query
	OrWhereBetween(column string, args []any) Query
	OrWhereColumn(column1 string, column2 ...string) Query
	OrWhereIn(column string, args []any) Query
	OrWhereLike(column string, value string) Query
	OrWhereNot(query any, args ...any) Query
	OrWhereNotBetween(column string, args []any) Query
	OrWhereNotIn(column string, args []any) Query
	OrWhereNotLike(column string, value string) Query
	OrWhereNotNull(column string) Query
	OrWhereNull(column string) Query
	OrWhereRaw(raw string, args []any) Query
	// Pluck(column string, dest any) error
	// rollBack
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
	WhereBetween(column string, args []any) Query
	WhereColumn(column1 string, column2 ...string) Query
	WhereExists(func() Query) Query
	WhereIn(column string, args []any) Query
	WhereLike(column string, value string) Query
	WhereNot(query any, args ...any) Query
	WhereNotBetween(column string, args []any) Query
	WhereNotIn(column string, args []any) Query
	WhereNotLike(column string, value string) Query
	WhereNotNull(column string) Query
	WhereNull(column string) Query
	WhereRaw(raw string, args []any) Query
}

type Result struct {
	RowsAffected int64
}

type Builder interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
}
