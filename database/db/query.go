package db

import (
	"context"
	databasesql "database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/framework/support/str"
)

type Query struct {
	builder    db.Builder
	conditions Conditions
	ctx        context.Context
	err        error
	driver     driver.Driver
	logger     logger.Logger
	txLogs     *[]TxLog
}

func NewQuery(ctx context.Context, driver driver.Driver, builder db.Builder, logger logger.Logger, table string, txLogs *[]TxLog) *Query {
	return &Query{
		builder: builder,
		conditions: Conditions{
			Table: table,
		},
		ctx:    ctx,
		driver: driver,
		logger: logger,
		txLogs: txLogs,
	}
}

func (r *Query) Count() (int64, error) {
	r.conditions.Selects = []string{"COUNT(*)"}

	sql, args, err := r.buildSelect()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.builder.Get(&count, sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)

		return 0, err
	}

	r.trace(sql, args, -1, nil)

	return count, nil
}

// func (r *Query) Chunk(size int, callback func(dest []any) error) error {
// 	sql, args, err := r.buildSelect()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (r *Query) Decrement(column string, value ...uint64) error {
	v := uint64(1)
	if len(value) > 0 {
		v = value[0]
	}

	_, err := r.Update(column, sq.Expr(fmt.Sprintf("%s - ?", column), v))
	if err != nil {
		return err
	}

	return nil
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

func (r *Query) Distinct() db.Query {
	q := r.clone()
	q.conditions.Distinct = convert.Pointer(true)

	return q
}

func (r *Query) DoesntExist() (bool, error) {
	count, err := r.Count()
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (r *Query) Exists() (bool, error) {
	count, err := r.Count()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Query) Find(dest any, conds ...any) error {
	var q db.Query
	if len(conds) > 2 {
		return errors.DatabaseInvalidArgumentNumber.Args(len(conds), "1 or 2")
	} else if len(conds) == 1 {
		q = r.Where("id", conds...)
	} else if len(conds) == 2 {
		q = r.Where(conds[0], conds[1])
	} else {
		q = r.clone()
	}

	destValue := reflect.Indirect(reflect.ValueOf(dest))
	if destValue.Kind() == reflect.Slice {
		return q.Get(dest)
	}

	return q.First(dest)
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

func (r *Query) FirstOr(dest any, callback func() error) error {
	sql, args, err := r.buildSelect()
	if err != nil {
		return err
	}

	err = r.builder.Get(dest, sql, args...)
	if err != nil {
		if errors.Is(err, databasesql.ErrNoRows) {
			r.trace(sql, args, 0, nil)

			return callback()
		}

		r.trace(sql, args, -1, err)

		return err
	}

	r.trace(sql, args, 1, nil)

	return nil
}

func (r *Query) FirstOrFail(dest any) error {
	sql, args, err := r.buildSelect()
	if err != nil {
		return err
	}

	err = r.builder.Get(dest, sql, args...)
	if err != nil {
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

func (r *Query) Increment(column string, value ...uint64) error {
	v := uint64(1)
	if len(value) > 0 {
		v = value[0]
	}

	_, err := r.Update(column, sq.Expr(fmt.Sprintf("%s + ?", column), v))
	if err != nil {
		return err
	}

	return nil
}

func (r *Query) Insert(data any) (*db.Result, error) {
	mapData, err := convertToSliceMap(data)
	if err != nil {
		return nil, err
	}
	if len(mapData) == 0 {
		return nil, errors.DatabaseDataIsEmpty
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

func (r *Query) InsertGetId(data any) (int64, error) {
	mapData, err := convertToMap(data)
	if err != nil {
		return 0, err
	}
	if len(mapData) == 0 {
		return 0, errors.DatabaseUnsupportedType.Args("nil", "struct, map[string]any").SetModule("DB")
	}

	sql, args, err := r.buildInsert([]map[string]any{mapData})
	if err != nil {
		return 0, err
	}

	result, err := r.builder.Exec(sql, args...)
	if err != nil {
		r.trace(sql, args, -1, err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.trace(sql, args, -1, err)
		return 0, err
	}

	r.trace(sql, args, id, nil)

	return id, nil
}

// func (r *Query) Limit(limit uint64) db.Query {
// 	q := r.clone()
// 	q.conditions.Limit = &limit

// 	return q
// }

func (r *Query) Latest(dest any, column ...string) error {
	col := "created_at"
	if len(column) > 0 {
		col = column[0]
	}

	return r.OrderByDesc(col).First(dest)
}

func (r *Query) OrderBy(column string) db.Query {
	q := r.clone()
	q.conditions.OrderBy = append(q.conditions.OrderBy, column+" ASC")

	return q
}

func (r *Query) OrderByDesc(column string) db.Query {
	q := r.clone()
	q.conditions.OrderBy = append(q.conditions.OrderBy, column+" DESC")

	return q
}

func (r *Query) OrderByRaw(raw string) db.Query {
	q := r.clone()
	q.conditions.OrderBy = append(q.conditions.OrderBy, raw)

	return q
}

func (r *Query) OrWhere(query any, args ...any) db.Query {
	q := r.clone()
	q.conditions.Where = append(q.conditions.Where, Where{
		query: query,
		args:  args,
		or:    true,
	})

	return q
}

func (r *Query) OrWhereBetween(column string, x, y any) db.Query {
	return r.OrWhere(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", column), x, y))
}

func (r *Query) OrWhereColumn(column1 string, column2 ...string) db.Query {
	if len(column2) == 0 || len(column2) > 2 {
		r.err = errors.DatabaseInvalidArgumentNumber.Args(len(column2), "1 or 2")
		return r
	}

	if len(column2) == 1 {
		return r.OrWhere(sq.Expr(fmt.Sprintf("%s = %s", column1, column2[0])))
	}

	return r.OrWhere(sq.Expr(fmt.Sprintf("%s %s %s", column1, column2[0], column2[1])))
}

func (r *Query) OrWhereIn(column string, args []any) db.Query {
	return r.OrWhere(column, args)
}

func (r *Query) OrWhereLike(column string, value string) db.Query {
	return r.OrWhere(sq.Like{column: value})
}

func (r *Query) OrWhereNot(query any, args ...any) db.Query {
	query, args, err := r.buildWhere(Where{
		query: query,
		args:  args,
	})
	if err != nil {
		r.err = err
		return r
	}

	sqlizer, err := r.toSqlizer(query, args)
	if err != nil {
		r.err = err
		return r
	}

	sql, args, err := sqlizer.ToSql()
	if err != nil {
		r.err = err
		return r
	}

	return r.OrWhere(sq.Expr(fmt.Sprintf("NOT (%s)", sql), args...))
}

func (r *Query) OrWhereNotBetween(column string, x, y any) db.Query {
	return r.OrWhere(sq.Expr(fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), x, y))
}

func (r *Query) OrWhereNotIn(column string, args []any) db.Query {
	return r.OrWhere(sq.NotEq{column: args})
}

func (r *Query) OrWhereNotLike(column string, value string) db.Query {
	return r.OrWhere(sq.NotLike{column: value})
}

func (r *Query) OrWhereNotNull(column string) db.Query {
	return r.OrWhere(sq.NotEq{column: nil})
}

func (r *Query) OrWhereNull(column string) db.Query {
	return r.OrWhere(sq.Eq{column: nil})
}

func (r *Query) OrWhereRaw(raw string, args []any) db.Query {
	return r.OrWhere(sq.Expr(raw, args...))
}

func (r *Query) Pluck(column string, dest any) error {
	r.conditions.Selects = []string{column}

	return r.Get(dest)
}

func (r *Query) Select(columns ...string) db.Query {
	q := r.clone()
	q.conditions.Selects = append(q.conditions.Selects, columns...)

	return q
}

func (r *Query) ToSql() db.ToSql {
	q := r.clone()
	return NewToSql(q, false)
}

func (r *Query) ToRawSql() db.ToSql {
	q := r.clone()
	return NewToSql(q, true)
}

func (r *Query) Update(column any, value ...any) (*db.Result, error) {
	columnStr, ok := column.(string)
	if ok {
		if len(value) != 1 {
			return nil, errors.DatabaseInvalidArgumentNumber.Args(len(value), "1")
		}

		return r.Update(map[string]any{columnStr: value[0]})
	}

	mapData, err := convertToMap(column)
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

// func (r *Query) Value(column string, dest any) error {
// 	r.conditions.Selects = []string{column}
// 	r.conditions.Limit = convert.Pointer(uint64(1))

// 	sql, args, err := r.buildSelect()
// 	if err != nil {
// 		return err
// 	}

// 	err = r.builder.Get(dest, sql, args...)
// 	if err != nil {
// 		if errors.Is(err, databasesql.ErrNoRows) {
// 			r.trace(sql, args, 0, nil)
// 			return nil
// 		}

// 		r.trace(sql, args, -1, err)

// 		return err
// 	}

// 	r.trace(sql, args, -1, nil)

// 	return nil
// }

func (r *Query) When(condition bool, callback func(query db.Query) db.Query) db.Query {
	if condition {
		return callback(r)
	}

	return r
}

func (r *Query) Where(query any, args ...any) db.Query {
	q := r.clone()
	q.conditions.Where = append(q.conditions.Where, Where{
		query: query,
		args:  args,
	})

	return q
}

func (r *Query) WhereBetween(column string, x, y any) db.Query {
	return r.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", column), x, y))
}

func (r *Query) WhereColumn(column1 string, column2 ...string) db.Query {
	if len(column2) == 0 || len(column2) > 2 {
		r.err = errors.DatabaseInvalidArgumentNumber.Args(len(column2), "1 or 2")
		return r
	}

	if len(column2) == 1 {
		return r.Where(sq.Expr(fmt.Sprintf("%s = %s", column1, column2[0])))
	}

	return r.Where(sq.Expr(fmt.Sprintf("%s %s %s", column1, column2[0], column2[1])))
}

func (r *Query) WhereExists(query func() db.Query) db.Query {
	subQuery := query()
	sql, args, err := subQuery.(*Query).buildSelect()
	if err != nil {
		r.err = err
		return r
	}

	sql = r.driver.Explain(sql, args...)

	return r.Where(sq.Expr(fmt.Sprintf("EXISTS (%s)", sql)))
}

func (r *Query) WhereIn(column string, args []any) db.Query {
	return r.Where(column, args)
}

func (r *Query) WhereLike(column string, value string) db.Query {
	return r.Where(sq.Like{column: value})
}

func (r *Query) WhereNot(query any, args ...any) db.Query {
	query, args, err := r.buildWhere(Where{
		query: query,
		args:  args,
	})
	if err != nil {
		r.err = err
		return r
	}

	sqlizer, err := r.toSqlizer(query, args)
	if err != nil {
		r.err = err
		return r
	}

	sql, args, err := sqlizer.ToSql()
	if err != nil {
		r.err = err
		return r
	}

	return r.Where(sq.Expr(fmt.Sprintf("NOT (%s)", sql), args...))
}

func (r *Query) WhereNotBetween(column string, x, y any) db.Query {
	return r.Where(sq.Expr(fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), x, y))
}

func (r *Query) WhereNotIn(column string, args []any) db.Query {
	return r.Where(sq.NotEq{column: args})
}

func (r *Query) WhereNotLike(column string, value string) db.Query {
	return r.Where(sq.NotLike{column: value})
}

func (r *Query) WhereNotNull(column string) db.Query {
	return r.Where(sq.NotEq{column: nil})
}

func (r *Query) WhereNull(column string) db.Query {
	return r.Where(sq.Eq{column: nil})
}

func (r *Query) WhereRaw(raw string, args []any) db.Query {
	return r.Where(sq.Expr(raw, args...))
}

func (r *Query) buildDelete() (sql string, args []any, err error) {
	if r.err != nil {
		return "", nil, r.err
	}

	if r.conditions.Table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Delete(r.conditions.Table)
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	sqlizer, err := r.buildWheres(r.conditions.Where)
	if err != nil {
		return "", nil, err
	}

	return builder.Where(sqlizer).ToSql()
}

func (r *Query) buildInsert(data []map[string]any) (sql string, args []any, err error) {
	if r.err != nil {
		return "", nil, r.err
	}

	if r.conditions.Table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Insert(r.conditions.Table)
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
	if r.err != nil {
		return "", nil, r.err
	}

	if r.conditions.Table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	selects := "*"
	if len(r.conditions.Selects) > 0 {
		selects = strings.Join(r.conditions.Selects, ", ")
	}

	builder := sq.Select(selects)

	if r.conditions.Distinct != nil && *r.conditions.Distinct {
		builder = builder.Distinct()
	}

	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	builder = builder.From(r.conditions.Table)
	sqlizer, err := r.buildWheres(r.conditions.Where)
	if err != nil {
		return "", nil, err
	}

	builder = builder.Where(sqlizer)
	builder = builder.OrderBy(r.conditions.OrderBy...)

	if r.conditions.Limit != nil {
		builder = builder.Limit(*r.conditions.Limit)
	}

	return builder.ToSql()
}

func (r *Query) buildUpdate(data map[string]any) (sql string, args []any, err error) {
	if r.err != nil {
		return "", nil, r.err
	}

	if r.conditions.Table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Update(r.conditions.Table)
	if placeholderFormat := r.placeholderFormat(); placeholderFormat != nil {
		builder = builder.PlaceholderFormat(placeholderFormat)
	}

	sqlizer, err := r.buildWheres(r.conditions.Where)
	if err != nil {
		return "", nil, err
	}

	return builder.Where(sqlizer).SetMap(data).ToSql()
}

func (r *Query) buildWhere(where Where) (any, []any, error) {
	switch query := where.query.(type) {
	case string:
		if !str.Of(query).Trim().Contains(" ", "?") {
			if len(where.args) > 1 {
				return sq.Eq{query: where.args}, nil, nil
			} else if len(where.args) == 1 {
				return sq.Eq{query: where.args[0]}, nil, nil
			}
		}
		return query, where.args, nil
	case func(db.Query) db.Query:
		// Handle nested conditions by creating a new query and applying the callback
		nestedQuery := NewQuery(r.ctx, r.driver, r.builder, r.logger, r.conditions.Table, r.txLogs)
		nestedQuery = query(nestedQuery).(*Query)

		// Build the nested conditions
		sqlizer, err := r.buildWheres(nestedQuery.conditions.Where)
		if err != nil {
			return nil, nil, err
		}

		return sqlizer, nil, nil
	default:
		return where.query, where.args, nil
	}
}

func (r *Query) buildWheres(wheres []Where) (sq.Sqlizer, error) {
	if len(wheres) == 0 {
		return nil, nil
	}

	var sqlizers []sq.Sqlizer
	for _, where := range wheres {
		query, args, err := r.buildWhere(where)
		if err != nil {
			return nil, err
		}

		sqlizer, err := r.toSqlizer(query, args)
		if err != nil {
			return nil, err
		}

		if where.or && len(sqlizers) > 0 {
			// If it's an OR condition and we have previous conditions,
			// wrap the previous conditions in an AND and create an OR condition
			if len(sqlizers) == 1 {
				sqlizers = []sq.Sqlizer{
					sq.Or{
						sqlizers[0],
						sqlizer,
					},
				}
			} else {
				sqlizers = []sq.Sqlizer{
					sq.Or{
						sq.And(sqlizers),
						sqlizer,
					},
				}
			}
		} else {
			// For regular WHERE conditions or the first condition
			sqlizers = append(sqlizers, sqlizer)
		}
	}

	if len(sqlizers) == 1 {
		return sqlizers[0], nil
	}

	return sq.And(sqlizers), nil
}

func (r *Query) clone() *Query {
	query := NewQuery(r.ctx, r.driver, r.builder, r.logger, r.conditions.Table, r.txLogs)
	query.conditions = r.conditions
	query.err = r.err

	return query
}

func (r *Query) placeholderFormat() database.PlaceholderFormat {
	if r.driver.Config().PlaceholderFormat != nil {
		return r.driver.Config().PlaceholderFormat
	}

	return nil
}

func (r *Query) toSqlizer(query any, args []any) (sq.Sqlizer, error) {
	switch q := query.(type) {
	case map[string]any:
		return sq.Eq(q), nil
	case string:
		return sq.Expr(q, args...), nil
	case sq.Sqlizer:
		return q, nil
	default:
		return nil, errors.DatabaseUnsupportedType.Args(fmt.Sprintf("%T", query), "string-keyed map or string or squirrel.Sqlizer")
	}
}

func (r *Query) trace(sql string, args []any, rowsAffected int64, err error) {
	if r.txLogs != nil {
		*r.txLogs = append(*r.txLogs, TxLog{
			ctx:          r.ctx,
			begin:        carbon.Now(),
			sql:          r.driver.Explain(sql, args...),
			rowsAffected: rowsAffected,
			err:          err,
		})
	} else {
		r.logger.Trace(r.ctx, carbon.Now(), r.driver.Explain(sql, args...), rowsAffected, err)
	}
}
