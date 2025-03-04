package db

import (
	"github.com/goravel/framework/errors"
)

type ToSql struct {
	query *Query
	raw   bool
}

func NewToSql(query *Query, raw bool) *ToSql {
	return &ToSql{query: query, raw: raw}
}

func (r *ToSql) Count() string {
	r.query.conditions.Selects = []string{"COUNT(*)"}

	return r.generate(r.query.buildSelect())
}

func (r *ToSql) Delete() string {
	return r.generate(r.query.buildDelete())
}

func (r *ToSql) First() string {
	return r.generate(r.query.buildSelect())
}

func (r *ToSql) Get() string {
	return r.generate(r.query.buildSelect())
}

func (r *ToSql) Insert(data any) string {
	mapData, err := convertToSliceMap(data)
	if err != nil {
		return r.generate("", nil, err)
	}
	if len(mapData) == 0 {
		return r.generate("", nil, errors.DatabaseDataIsEmpty)
	}

	return r.generate(r.query.buildInsert(mapData))
}

func (r *ToSql) Pluck(column string, dest any) string {
	r.query.conditions.Selects = []string{column}

	return r.generate(r.query.buildSelect())
}

func (r *ToSql) Update(column any, value ...any) string {
	columnStr, ok := column.(string)
	if ok {
		if len(value) != 1 {
			return r.generate("", nil, errors.DatabaseInvalidArgumentNumber.Args(len(value), "1"))
		}

		return r.Update(map[string]any{columnStr: value[0]})
	}

	mapData, err := convertToMap(column)
	if err != nil {
		return r.generate("", nil, err)
	}

	return r.generate(r.query.buildUpdate(mapData))
}

func (r *ToSql) generate(sql string, args []any, err error) string {
	if err != nil {
		r.query.logger.Errorf(r.query.ctx, errors.DatabaseFailedToGetSql.Args(err).Error())

		return ""
	}

	if r.raw {
		return r.query.driver.Explain(sql, args...)
	}

	return sql
}
