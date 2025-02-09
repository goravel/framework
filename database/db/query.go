package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
)

type Query struct {
	conditions Conditions
	instance   *sqlx.DB
}

func NewQuery(instance *sqlx.DB, table string) *Query {
	return &Query{
		conditions: Conditions{
			table: table,
		},
		instance: instance,
	}
}

func (r *Query) Where(query any, args ...any) db.Query {
	r.conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
	})

	return r
}

func (r *Query) Get(dest any) error {
	sql, args, err := r.buildSelect()
	fmt.Println(sql, args, err)
	if err != nil {
		return err
	}

	return r.instance.Select(dest, sql, args...)
}

func (r *Query) buildSelect() (sql string, args []any, err error) {
	if r.conditions.table == "" {
		return "", nil, errors.DatabaseTableIsRequired
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	builder := psql.Select("*").From(r.conditions.table)

	for _, where := range r.conditions.where {
		builder = builder.Where(where.query, where.args...)
	}

	return builder.ToSql()
}
