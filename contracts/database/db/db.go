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
	// commit
	Count() (int64, error)
	// Chunk(size int, callback func(rows []any) error) error
	// CrossJoin(table string, on any, args ...any) Query
	// DoesntExist() (bool, error)
	// Distinct() Query
	// dump
	// dumpRawSql
	Delete() (*Result, error)
	// Each(callback func(rows []any) error) error
	Exists() (bool, error)
	Find(dest any, conds ...any) error
	First(dest any) error
	// FirstOr
	FirstOrFail(dest any) error
	// decrement
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
	// Limit(limit uint64) Query
	// lockForUpdate
	// offset
	OrderBy(column string) Query
	OrderByDesc(column string) Query
	OrderByRaw(raw string) Query
	OrWhere(query any, args ...any) Query
	OrWhereBetween(column string, x, y any) Query
	OrWhereColumn(column1 string, column2 ...string) Query
	OrWhereIn(column string, args []any) Query
	OrWhereLike(column string, value string) Query
	OrWhereNot(query any, args ...any) Query
	OrWhereNotBetween(column string, x, y any) Query
	OrWhereNotIn(column string, args []any) Query
	OrWhereNotLike(column string, value string) Query
	OrWhereNotNull(column string) Query
	OrWhereNull(column string) Query
	OrWhereRaw(raw string, args []any) Query
	Pluck(column string, dest any) error
	// rollBack
	// RightJoin(table string, on any, args ...any) Query
	Select(columns ...string) Query
	// sharedLock
	// skip
	// take
	// ToSql
	// ToRawSql
	Update(column any, value ...any) (*Result, error)
	// updateOrInsert
	// Value(column string, dest any) error
	// when
	Where(query any, args ...any) Query
	WhereBetween(column string, x, y any) Query
	WhereColumn(column1 string, column2 ...string) Query
	WhereExists(func() Query) Query
	WhereIn(column string, args []any) Query
	WhereLike(column string, value string) Query
	WhereNot(query any, args ...any) Query
	WhereNotBetween(column string, x, y any) Query
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
	// Query(query string, args ...any) (*sql.Rows, error)
	Select(dest any, query string, args ...any) error
}
