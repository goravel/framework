package gorm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"
)

func TestApplyOneOfManyJoin_HasOne(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		kind:         relKindHasOne,
		relatedTable: "books",
		references: []referenceKey{{
			primaryTable:  "users",
			primaryColumn: "id",
			foreignTable:  "books",
			foreignColumn: "user_id",
		}},
	}
	cfg := &oneOfManyConfig{column: "published_at", aggregate: "MAX"}
	inner := q.freshSession().Table("books")

	got := q.applyOneOfManyJoin(inner, desc, cfg)
	stmt := got.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "INNER JOIN (")
	assert.Contains(t, sql, "MAX(")
	assert.Contains(t, strings.ToLower(sql), `published_at`)
	assert.Contains(t, sql, `user_id`)
}

func TestApplyOneOfManyJoin_MorphOne_IncludesTypeColumn(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		kind:            relKindMorphOne,
		relatedTable:    "logos",
		morphTypeColumn: "logoable_type",
		morphIDColumn:   "logoable_id",
		references: []referenceKey{{
			primaryTable:  "users",
			primaryColumn: "id",
			foreignTable:  "logos",
			foreignColumn: "logoable_id",
		}},
	}
	cfg := &oneOfManyConfig{column: "id", aggregate: "MIN"}
	inner := q.freshSession().Table("logos")

	got := q.applyOneOfManyJoin(inner, desc, cfg)
	stmt := got.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "MIN(")
	assert.Contains(t, sql, "logoable_id")
	assert.Contains(t, sql, "logoable_type") // morph type included in GROUP BY / JOIN
}

func TestApplyOneOfManyJoin_UnsupportedKind_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		kind:         relKindHasMany,
		relatedTable: "books",
		references: []referenceKey{{
			primaryTable:  "users",
			primaryColumn: "id",
			foreignTable:  "books",
			foreignColumn: "user_id",
		}},
	}
	cfg := &oneOfManyConfig{column: "id", aggregate: "MAX"}
	inner := q.freshSession().Table("books")

	got := q.applyOneOfManyJoin(inner, desc, cfg)
	stmt := got.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	sql := stmt.Statement.SQL.String()
	assert.NotContains(t, strings.ToUpper(sql), "INNER JOIN (")
}

func TestApplyOneOfManyJoin_NilCfg_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		kind:         relKindHasOne,
		relatedTable: "books",
		references: []referenceKey{{
			primaryTable:  "users",
			primaryColumn: "id",
			foreignTable:  "books",
			foreignColumn: "user_id",
		}},
	}
	inner := q.freshSession().Table("books")

	got := q.applyOneOfManyJoin(inner, desc, nil)
	stmt := got.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	sql := stmt.Statement.SQL.String()
	assert.NotContains(t, strings.ToUpper(sql), "INNER JOIN (")
}

func TestQuery_OfManyShortcuts(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})

	latest := q.LatestOfMany().(*Query)
	assert.NotNil(t, latest.conditions.oneOfMany)
	assert.Equal(t, "MAX", latest.conditions.oneOfMany.aggregate)
	assert.Equal(t, "id", latest.conditions.oneOfMany.column)

	oldest := q.OldestOfMany("created_at").(*Query)
	assert.NotNil(t, oldest.conditions.oneOfMany)
	assert.Equal(t, "MIN", oldest.conditions.oneOfMany.aggregate)
	assert.Equal(t, "created_at", oldest.conditions.oneOfMany.column)

	custom := q.OfMany("score", "AVG").(*Query)
	assert.Equal(t, "AVG", custom.conditions.oneOfMany.aggregate)
	assert.Equal(t, "score", custom.conditions.oneOfMany.column)
}
