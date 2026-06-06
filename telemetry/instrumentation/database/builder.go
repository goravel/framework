package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	contractsdb "github.com/goravel/framework/contracts/database/db"
)

type tracedCommonBuilder struct {
	inner  contractsdb.CommonBuilder
	system string
}

func WrapBuilder(inner contractsdb.CommonBuilder, driverName string) contractsdb.CommonBuilder {
	return &tracedCommonBuilder{inner: inner, system: driverName}
}

type tracedBuilder struct {
	*tracedCommonBuilder
	inner contractsdb.Builder
}

func (r *tracedBuilder) Beginx() (*sqlx.Tx, error) { return r.inner.Beginx() }

func WrapBuilderFull(inner contractsdb.Builder, driverName string) contractsdb.Builder {
	return &tracedBuilder{tracedCommonBuilder: &tracedCommonBuilder{inner: inner, system: driverName}, inner: inner}
}

type tracedTxBuilder struct {
	*tracedCommonBuilder
	inner contractsdb.TxBuilder
}

func (r *tracedTxBuilder) Commit() error   { return r.inner.Commit() }
func (r *tracedTxBuilder) Rollback() error { return r.inner.Rollback() }

func WrapTxBuilder(inner contractsdb.TxBuilder, driverName string) contractsdb.TxBuilder {
	return &tracedTxBuilder{tracedCommonBuilder: &tracedCommonBuilder{inner: inner, system: driverName}, inner: inner}
}

func (r *tracedCommonBuilder) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	spanCtx, span, ok := startSpan(ctx, operationName(query))
	if !ok {
		return r.inner.ExecContext(ctx, query, args...)
	}

	start := time.Now()
	result, err := r.inner.ExecContext(spanCtx, query, args...)

	rows := int64(-1)
	if err == nil && result != nil {
		if affected, affErr := result.RowsAffected(); affErr == nil {
			rows = affected
		}
	}
	endSpan(spanCtx, span, start, dbSystem(r.system), query, "", rows, err)

	return result, err
}

func (r *tracedCommonBuilder) Explain(sql string, args ...any) string {
	return r.inner.Explain(sql, args...)
}

func (r *tracedCommonBuilder) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	spanCtx, span, ok := startSpan(ctx, operationName(query))
	if !ok {
		return r.inner.GetContext(ctx, dest, query, args...)
	}

	start := time.Now()
	err := r.inner.GetContext(spanCtx, dest, query, args...)
	endSpan(spanCtx, span, start, dbSystem(r.system), query, "", -1, err)

	return err
}

func (r *tracedCommonBuilder) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	spanCtx, span, ok := startSpan(ctx, operationName(query))
	if !ok {
		return r.inner.QueryxContext(ctx, query, args...)
	}

	start := time.Now()
	rows, err := r.inner.QueryxContext(spanCtx, query, args...)
	endSpan(spanCtx, span, start, dbSystem(r.system), query, "", -1, err)

	return rows, err
}

func (r *tracedCommonBuilder) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	spanCtx, span, ok := startSpan(ctx, operationName(query))
	if !ok {
		return r.inner.SelectContext(ctx, dest, query, args...)
	}

	start := time.Now()
	err := r.inner.SelectContext(spanCtx, dest, query, args...)
	endSpan(spanCtx, span, start, dbSystem(r.system), query, "", -1, err)

	return err
}
