package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DB interface {
	Connection(name string) DB
	Query() Sqlx
	Transaction(ctx context.Context, txFunc func(tx *sqlx.Tx) error) error
}

type Sqlx interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
	Transaction
}

type Transaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}
