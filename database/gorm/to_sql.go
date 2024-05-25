package gorm

import (
	"gorm.io/gorm"
)

type ToSql struct {
	query *QueryImpl
	raw   bool
}

func NewToSql(query *QueryImpl, raw bool) *ToSql {
	return &ToSql{
		query: query,
		raw:   raw,
	}
}

func (r *ToSql) Count() string {
	query := r.query.buildConditions()
	var count int64

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Count(&count))
}

func (r *ToSql) Create(value any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Create(value))
}

func (r *ToSql) Delete(value any, conds ...any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Delete(value, conds...))
}

func (r *ToSql) Find(dest any, conds ...any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Find(dest, conds...))
}

func (r *ToSql) First(dest any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).First(dest))
}

func (r *ToSql) Get(dest any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Find(dest))
}

func (r *ToSql) Pluck(column string, dest any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Pluck(column, dest))
}

func (r *ToSql) Save(value any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Save(value))
}

func (r *ToSql) Sum(column string, dest any) string {
	query := r.query.buildConditions()

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Select("SUM(" + column + ")").Find(dest))
}

func (r *ToSql) Update(column any, value ...any) string {
	query := r.query.buildConditions()
	if _, ok := column.(string); !ok && len(value) > 0 {
		return ""
	}

	if c, ok := column.(string); ok && len(value) > 0 {
		query.instance.Statement.Dest = map[string]any{c: value[0]}
	}
	if len(value) == 0 {
		query.instance.Statement.Dest = column
	}

	return r.sql(query.instance.Session(&gorm.Session{DryRun: true}).Updates(query.instance.Statement.Dest))
}

func (r *ToSql) sql(db *gorm.DB) string {
	sql := db.Statement.SQL.String()
	if !r.raw {
		return sql
	}

	return r.query.instance.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Raw(sql, db.Statement.Vars...)
	})
}
