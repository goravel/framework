package database

import (
	"context"
	"database/sql"
)

type DB interface {
	Connection(name string) DB
	Query() Sqlx
}

type Sqlx interface {
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	//sqlx.QueryerContext
}
