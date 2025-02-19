package db

import (
	"fmt"
	"sort"

	sq "github.com/Masterminds/squirrel"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
)

type Query struct {
	builder    db.Builder
	conditions Conditions
	config     database.Config
}

func NewQuery(config database.Config, builder db.Builder, table string) *Query {
	return &Query{
		builder: builder,
		conditions: Conditions{
			table: table,
		},
		config: config,
	}
}

func (r *Query) First(dest any) error {
	sql, args, err := r.buildSelect()
	// TODO: use logger instead of println
	fmt.Println(sql, args, err)
	if err != nil {
		return err
	}

	return r.builder.Get(dest, sql, args...)
}

func (r *Query) Get(dest any) error {
	sql, args, err := r.buildSelect()
	// TODO: use logger instead of println
	fmt.Println(sql, args, err)
	if err != nil {
		return err
	}

	return r.builder.Select(dest, sql, args...)
}

func (r *Query) Insert(data any) (*db.Result, error) {
	mapData, err := convertToSliceMap(data)
	if err != nil {
		return nil, err
	}

	sql, args, err := r.buildInsert(mapData)
	if err != nil {
		return nil, err
	}
	// TODO: use logger instead of println
	fmt.Println(sql, args, err)
	result, err := r.builder.Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &db.Result{
		RowsAffected: rowsAffected,
	}, nil
}

func (r *Query) Where(query any, args ...any) db.Query {
	q := NewQuery(r.config, r.builder, r.conditions.table)
	q.conditions = r.conditions
	q.conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
	})

	return q
}

func (r *Query) buildInsert(data []map[string]any) (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	builder := sq.Insert(r.conditions.table)
	if r.config.PlaceholderFormat != nil {
		builder = builder.PlaceholderFormat(r.config.PlaceholderFormat)
	}

	first := data[0]
	builder = builder.SetMap(first)

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
	if r.config.PlaceholderFormat != nil {
		builder = builder.PlaceholderFormat(r.config.PlaceholderFormat)
	}

	builder = builder.From(r.conditions.table)

	for _, where := range r.conditions.where {
		query, ok := where.query.(string)
		if ok {
			if !str.Of(query).Trim().Contains(" ", "?") {
				if len(where.args) > 1 {
					builder = builder.Where(sq.Eq{query: where.args})
				} else if len(where.args) == 1 {
					builder = builder.Where(sq.Eq{query: where.args[0]})
				}
				continue
			}
		}

		builder = builder.Where(where.query, where.args...)
	}

	return builder.ToSql()
}
