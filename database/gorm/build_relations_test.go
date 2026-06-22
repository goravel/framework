package gorm

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// newRelQueryWith returns a Query whose conditions.model is preset, ready for build/compile tests.
func newRelQueryWith(t *testing.T, model any) *Query {
	t.Helper()
	db := newStubGormDB(t)
	conditions := Conditions{model: model}
	return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, &conditions)
}

func dryRunSQL(t *testing.T, db *gormio.DB) string {
	t.Helper()
	stmt := db.Session(&gormio.Session{DryRun: true}).Find(&relUser{})
	return stmt.Statement.SQL.String()
}

// --- buildRelations -------------------------------------------------------

func TestBuildRelations_NoOps(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	got := q.buildRelations(q.instance)
	assert.Same(t, q.instance, got)
}

func TestBuildRelations_NoParent(t *testing.T) {
	q := newRelQuery(t)
	q.conditions.relations = []relationExistence{{relation: "Books", operator: ">=", count: 1, conjunction: "and"}}
	got := q.buildRelations(q.instance)
	assert.True(t, errors.Is(got.Error, errors.OrmQueryEmptyRelation))
}

func TestBuildRelations_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{relation: "Books", operator: ">=", count: 1, conjunction: "and"}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "EXISTS")
	assert.Contains(t, sql, "rel_books")
}

func TestBuildRelations_DoesntHaveBuildsNotExists(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{relation: "Books", operator: "<", count: 1, conjunction: "and"}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "NOT EXISTS")
}

func TestBuildRelations_CountComparison(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{relation: "Books", operator: ">", count: 3, conjunction: "and"}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	// when not the EXISTS-eligible shape, build a (?) > ? clause instead
	assert.Contains(t, sql, "COUNT(*)")
}

func TestBuildRelations_ResolveError(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{relation: "Missing", operator: ">=", count: 1, conjunction: "and"}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	assert.True(t, errors.Is(out.Error, errors.OrmRelationNotFound))
}

func TestBuildRelations_OrConjunction(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{
		{relation: "Books", operator: ">=", count: 1, conjunction: "and"},
		{relation: "Roles", operator: ">=", count: 1, conjunction: "or"},
	}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "OR")
}

// --- compileExistenceSubquery (covers HasOne/HasMany/BelongsTo/M2M/Morph/Through SQL) ---

func TestCompileExistenceSubquery_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
	sql := stmt.Statement.SQL.String()
	assert.True(t, strings.Contains(sql, "rel_books"))
}

func TestCompileExistenceSubquery_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	desc, err := resolveRelation(q.instance, &relBook{}, "Author")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	assert.NotNil(t, inner)
}

func TestCompileExistenceSubquery_Many2Many(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relRole{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "rel_user_roles")
}

func TestCompileExistenceSubquery_Morph(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Houses")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	assert.NotNil(t, inner)
}

func TestCompileExistenceSubquery_Through(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	desc, err := resolveRelation(q.instance, &relCountry{}, "Posts")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relPost{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "rel_users") // through table
}

func TestCompileExistenceSubquery_WithCallback(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	cb := contractsorm.RelationCallback(func(qq contractsorm.Query) contractsorm.Query {
		return qq.Where("title = ?", "x")
	})
	inner := q.compileExistenceSubquery(desc, cb)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "title")
}

func TestCompileExistenceSubquery_NestedRelation(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books.Author")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	assert.NotNil(t, inner)
}

// --- compileAggregateSubquery ---------------------------------------------

func TestCompileAggregateSubquery_Count(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	sub := selectSub{relation: "Books", column: "*", function: "count"}
	inner := q.compileAggregateSubquery(desc, sub)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "COUNT(*)")
}

func TestCompileAggregateSubquery_Sum(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	sub := selectSub{relation: "Books", column: "id", function: "sum"}
	inner := q.compileAggregateSubquery(desc, sub)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "SUM(")
}

func TestCompileAggregateSubquery_Exists(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	sub := selectSub{relation: "Books", column: "*", function: "exists"}
	inner := q.compileAggregateSubquery(desc, sub)
	assert.NotNil(t, inner)
}

// --- buildSelectSubAggregates ---------------------------------------------

func TestBuildSelectSubAggregates_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	got := q.buildSelectSubAggregates(q.instance)
	assert.Same(t, q.instance, got)
}

func TestBuildSelectSubAggregates_NoParent(t *testing.T) {
	q := newRelQuery(t)
	q.conditions.selectSubs = []selectSub{{relation: "Books", column: "*", function: "count", alias: "books_count"}}
	got := q.buildSelectSubAggregates(q.instance)
	assert.True(t, errors.Is(got.Error, errors.OrmQueryEmptyRelation))
}

func TestBuildSelectSubAggregates_Count(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.selectSubs = []selectSub{{relation: "Books", column: "*", function: "count", alias: "books_count"}}
	out := q.buildSelectSubAggregates(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "books_count")
}

func TestBuildSelectSubAggregates_Exists(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.selectSubs = []selectSub{{relation: "Books", column: "*", function: "exists", alias: "has_books"}}
	out := q.buildSelectSubAggregates(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "CASE WHEN EXISTS")
	assert.Contains(t, sql, "has_books")
}

func TestBuildSelectSubAggregates_BadRelationRecordsError(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.selectSubs = []selectSub{{relation: "Missing", column: "*", function: "count"}}
	out := q.buildSelectSubAggregates(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	assert.Error(t, out.Error)
}

// --- applyMorphExistence --------------------------------------------------

func TestApplyMorphExistence_BuildsSQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{
		relation:    "Houses",
		operator:    ">=",
		count:       1,
		conjunction: "and",
		morphTypes:  []any{&relUser{}},
	}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "houseable_type")
}

func TestApplyMorphExistence_NonMorphRelationRecordsError(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{
		relation:    "Books", // hasMany, not morph
		operator:    ">=",
		count:       1,
		conjunction: "and",
		morphTypes:  []any{&relUser{}},
	}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	assert.Error(t, out.Error)
}

func TestApplyMorphExistence_OrConjunction(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{
		relation:    "Houses",
		operator:    ">=",
		count:       1,
		conjunction: "or",
		morphTypes:  []any{&relUser{}},
	}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	assert.NotNil(t, out)
}

func TestApplyMorphExistence_DoesntHave(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{
		relation:    "Houses",
		operator:    "<",
		count:       1,
		conjunction: "and",
		morphTypes:  []any{&relUser{}},
	}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "NOT EXISTS")
}

func TestApplyMorphExistence_CountComparison(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	q.conditions.relations = []relationExistence{{
		relation:    "Houses",
		operator:    ">",
		count:       3,
		conjunction: "and",
		morphTypes:  []any{&relUser{}},
	}}
	out := q.buildRelations(q.instance.Session(&gormio.Session{}).Model(&relUser{}))
	sql := dryRunSQL(t, out)
	assert.Contains(t, sql, "COUNT(*)")
}
