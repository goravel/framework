package db

import "database/sql"

type DB interface {
	Table(name string) Query
}

type Query interface {
	First(dest any) error
	Get(dest any) error
	Insert(data any) (*Result, error)
	Where(query any, args ...any) Query
}

type Result struct {
	RowsAffected int64
}

type Builder interface {
	Exec(query string, args ...any) (sql.Result, error)
	Get(dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
}
