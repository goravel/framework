package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
)

type Query struct {
	conditions Conditions
	config     database.Config
	instance   *sqlx.DB
}

func NewQuery(config database.Config, instance *sqlx.DB, table string) *Query {
	return &Query{
		conditions: Conditions{
			table: table,
		},
		config:   config,
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
	// TODO: use logger instead of println
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

	builder := sq.Select("*")
	if r.config.PlaceholderFormat != nil {
		builder = builder.PlaceholderFormat(r.config.PlaceholderFormat)
	}

	builder = builder.From(r.conditions.table)

	for _, where := range r.conditions.where {
		builder = builder.Where(where.query, where.args...)
	}

	return builder.ToSql()
}
