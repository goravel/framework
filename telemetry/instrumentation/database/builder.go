package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	contractsdb "github.com/goravel/framework/contracts/database/db"
)

type tableKeyType struct{}

var tableKey = tableKeyType{}

// ContextWithTable tags ctx with the collection a query targets so the wrapped
// builder can set db.collection.name for structured queries. Raw SQL omits it.
func ContextWithTable(ctx context.Context, table string) context.Context {
	return context.WithValue(ctx, tableKey, table)
}

// WrapBuilder wraps a query builder with telemetry, or returns it unchanged when
// inst is nil (telemetry off).
func WrapBuilder(inner contractsdb.Builder, inst *Instrument) contractsdb.Builder {
	if inst == nil {
		return inner
	}

	return &tracedBuilder{Builder: inner, instrument: inst}
}

// WrapTxBuilder wraps a transaction builder with telemetry, or returns it
// unchanged when inst is nil.
func WrapTxBuilder(inner contractsdb.TxBuilder, inst *Instrument) contractsdb.TxBuilder {
	if inst == nil {
		return inner
	}

	return &tracedTxBuilder{TxBuilder: inner, instrument: inst}
}

type tracedBuilder struct {
	contractsdb.Builder
	instrument *Instrument
}

func (r *tracedBuilder) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tracedExec(ctx, r.instrument, r.Builder, query, args...)
}

func (r *tracedBuilder) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return tracedRead(ctx, r.instrument, r.Builder.GetContext, dest, query, args...)
}

func (r *tracedBuilder) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return tracedRead(ctx, r.instrument, r.Builder.SelectContext, dest, query, args...)
}

func (r *tracedBuilder) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return tracedQueryx(ctx, r.instrument, r.Builder, query, args...)
}

type tracedTxBuilder struct {
	contractsdb.TxBuilder
	instrument *Instrument
}

func (r *tracedTxBuilder) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tracedExec(ctx, r.instrument, r.TxBuilder, query, args...)
}

func (r *tracedTxBuilder) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return tracedRead(ctx, r.instrument, r.TxBuilder.GetContext, dest, query, args...)
}

func (r *tracedTxBuilder) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return tracedRead(ctx, r.instrument, r.TxBuilder.SelectContext, dest, query, args...)
}

func (r *tracedTxBuilder) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return tracedQueryx(ctx, r.instrument, r.TxBuilder, query, args...)
}

func tableFromContext(ctx context.Context) string {
	table, _ := ctx.Value(tableKey).(string)
	return table
}

func tracedExec(ctx context.Context, inst *Instrument, inner contractsdb.CommonBuilder, query string, args ...any) (sql.Result, error) {
	if !inst.active() {
		return inner.ExecContext(ctx, query, args...)
	}

	start := time.Now()
	spanCtx, span := inst.startSpan(ctx, operationName(query))

	result, err := inner.ExecContext(spanCtx, query, args...)

	rows := int64(-1)
	if err == nil {
		if affected, rowsErr := result.RowsAffected(); rowsErr == nil {
			rows = affected
		}
	}

	inst.endSpan(spanCtx, span, start, query, tableFromContext(ctx), rows, err)

	return result, err
}

func tracedRead(ctx context.Context, inst *Instrument, exec func(context.Context, any, string, ...any) error, dest any, query string, args ...any) error {
	if !inst.active() {
		return exec(ctx, dest, query, args...)
	}

	start := time.Now()
	spanCtx, span := inst.startSpan(ctx, operationName(query))

	err := exec(spanCtx, dest, query, args...)

	inst.endSpan(spanCtx, span, start, query, tableFromContext(ctx), -1, err)

	return err
}

func tracedQueryx(ctx context.Context, inst *Instrument, inner contractsdb.CommonBuilder, query string, args ...any) (*sqlx.Rows, error) {
	if !inst.active() {
		return inner.QueryxContext(ctx, query, args...)
	}

	start := time.Now()
	spanCtx, span := inst.startSpan(ctx, operationName(query))

	rows, err := inner.QueryxContext(spanCtx, query, args...)

	inst.endSpan(spanCtx, span, start, query, tableFromContext(ctx), -1, err)

	return rows, err
}
