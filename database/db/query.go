package db

import (
	"context"
	databasesql "database/sql"
	"reflect"
	"sort"

	sq "github.com/Masterminds/squirrel"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/str"
)

type Query struct {
	builder    db.Builder
	conditions Conditions
	ctx        context.Context
	driver     driver.Driver
	logger     logger.Logger
}

func NewQuery(ctx context.Context, driver driver.Driver, builder db.Builder, logger logger.Logger, table string) *Query {
	return &Query{
		builder: builder,
		conditions: Conditions{
			table: table,
		},
		driver: driver,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *Query) Delete() (*db.Result, error) {
	sql, args, err := r.buildDelete()
	if err != nil {
		return nil, err
	}

	result, err := r.builder.Exec(sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	r.trace(sql, args, rowsAffected, nil)

	return &db.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (r *Query) First(dest any) error {
	sql, args, err := r.buildSelect()
	if err != nil {
		return err
	}

	err = r.builder.Get(dest, sql, args...)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			r.trace(sql, args, 0, nil)
			return nil
		}

		r.trace(sql, args, -1, err)

		return err
	}

	r.trace(sql, args, 1, nil)

	return nil
}

func (r *Query) Get(dest any) error {
	sql, args, err := r.buildSelect()
	if err != nil {
		return err
	}

	err = r.builder.Select(dest, sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)
		return err
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Ptr {
		destValue = destValue.Elem()
	}

	rowsAffected := int64(-1)
	if destValue.Kind() == reflect.Slice {
		rowsAffected = int64(destValue.Len())
	}

	r.trace(sql, args, rowsAffected, nil)

	return nil
}

func (r *Query) Insert(data any) (*db.Result, error) {
	mapData, err := convertToSliceMap(data)
	if err != nil {
		return nil, err
	}
	if len(mapData) == 0 {
		return &db.Result{
			RowsAffected: 0,
		}, nil
	}

	sql, args, err := r.buildInsert(mapData)
	if err != nil {
		return nil, err
	}

	result, err := r.builder.Exec(sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	r.trace(sql, args, rowsAffected, nil)

	return &db.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (r *Query) OrWhere(query any, args ...any) db.Query {
	q := NewQuery(r.ctx, r.driver, r.builder, r.logger, r.conditions.table)
	q.conditions = r.conditions
	q.conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
		or:    true,
	})

	return q
}

func (r *Query) Update(data any) (*db.Result, error) {
	mapData, err := convertToMap(data)
	if err != nil {
		return nil, err
	}

	sql, args, err := r.buildUpdate(mapData)
	if err != nil {
		return nil, err
	}

	result, err := r.builder.Exec(sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.trace(sql, args, -1, err)
		return nil, err
	}

	r.trace(sql, args, rowsAffected, nil)

	return &db.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (r *Query) Where(query any, args ...any) db.Query {
	q := NewQuery(r.ctx, r.driver, r.builder, r.logger, r.conditions.table)
	q.conditions = r.conditions
	q.conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
	})

	return q
}

func (r *Query) buildDelete() (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Delete(r.conditions.table)
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	for _, where := range r.conditions.where {
		query, args := r.buildWhere(where)
		builder = builder.Where(query, args...)
	}

	return builder.ToSql()
}

func (r *Query) buildInsert(data []map[string]any) (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Insert(r.conditions.table)
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	first := data[0]
	cols := make([]string, 0, len(first))
	for col := range first {
		cols = append(cols, col)
	}
	sort.Strings(cols)
	builder = builder.Columns(cols...)

	for _, row := range data {
		vals := make([]any, 0, len(first))
		for _, col := range cols {
			vals = append(vals, row[col])
		}
		builder = builder.Values(vals...)
	}

	return builder.ToSql()
}

func (r *Query) buildSelect() (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Select("*")
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	builder = builder.From(r.conditions.table)

	for _, where := range r.conditions.where {
		query, args := r.buildWhere(where)
		builder = builder.Where(query, args...)
	}

	return builder.ToSql()
}

func (r *Query) buildUpdate(data map[string]any) (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Update(r.conditions.table)
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	for _, where := range r.conditions.where {
		query, args := r.buildWhere(where)
		builder = builder.Where(query, args...)
	}

	builder = builder.SetMap(data)

	return builder.ToSql()
}

func (r *Query) buildWhere(where Where) (any, []any) {
	query, ok := where.query.(string)
	if ok {
		if !str.Of(query).Trim().Contains(" ", "?") {
			if len(where.args) > 1 {
				return sq.Eq{query: where.args}, nil
			} else if len(where.args) == 1 {
				return sq.Eq{query: where.args[0]}, nil
			}
		}
	}

	// if where.or {
	// 	return sq.Or{where.query}, where.args
	// }

	return where.query, where.args
}

func (r *Query) placeholderFormat() database.PlaceholderFormat {
	if r.driver.Config().PlaceholderFormat != nil {
		return r.driver.Config().PlaceholderFormat
	}

	return nil
}

func (r *Query) trace(sql string, args []any, rowsAffected int64, err error) {
	r.logger.Trace(r.ctx, carbon.Now(), r.driver.Explain(sql, args...), rowsAffected, err)
}
