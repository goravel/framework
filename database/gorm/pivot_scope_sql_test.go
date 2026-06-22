package gorm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// This file pins the exact SQL produced when OnPivotQuery is wired up, across the three read
// paths where it must take effect:
//
//  1. Query.Related(...) for Many2Many / MorphToMany  -> new_relation.go:newMany2ManyQuery
//  2. Eager-load Many2Many                            -> eager_loader.go:loadMany2Many
//  3. Eager-load MorphToMany / MorphedByMany          -> eager_loader.go:loadMorphToMany
//
// (1) is tested by inspecting the SQL Related() emits.
// (2) and (3) are tested by replicating the exact *gormio.DB chain those code paths build, then
//     applying applyOnPivotQuery — i.e. testing the wiring in isolation, since chunkedFindMaps
//     would otherwise need a real backing DB to execute against.

// --- Fixtures --------------------------------------------------------------

// scopeUser declares a Many2Many to scopeRole with an OnPivotQuery that filters on active = 1.
type scopeUser struct {
	ID    uint
	Name  string
	Roles []*scopeRole `gorm:"-"`
}

func (scopeUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Roles": contractsorm.Many2Many{
			Related: &scopeRole{},
			Table:   "scope_user_roles",
			OnPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
				return q.Where("active", 1)
			},
		},
	}
}

type scopeRole struct {
	ID   uint
	Name string
}

// scopeMorphPost declares a MorphToMany to scopeMorphTag with an OnPivotQuery filtering on
// deleted_at IS NULL.
type scopeMorphPost struct {
	ID    uint
	Title string
	Tags  []*scopeMorphTag `gorm:"-"`
}

func (scopeMorphPost) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Tags": contractsorm.MorphToMany{
			Related: &scopeMorphTag{},
			Name:    "taggable",
			OnPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
				return q.WhereNull("deleted_at")
			},
		},
	}
}

type scopeMorphTag struct {
	ID   uint
	Name string
}

// --- (1) Related() SQL -----------------------------------------------------

func TestRelated_Many2Many_WithOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &scopeUser{})
	rel := q.Related(&scopeUser{ID: 7}, "Roles")
	sql := newRelationSQL(t, rel, &[]scopeRole{})
	assert.Equal(t,
		`SELECT "scope_roles"."id","scope_roles"."name" FROM "scope_roles" INNER JOIN scope_user_roles ON scope_user_roles.scope_role_id = scope_roles.id WHERE scope_user_roles.scope_user_id = ? AND scope_user_roles.active = ?`,
		sql,
	)
}

func TestRelated_MorphToMany_WithOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &scopeMorphPost{})
	rel := q.Related(&scopeMorphPost{ID: 3}, "Tags")
	sql := newRelationSQL(t, rel, &[]scopeMorphTag{})
	assert.Equal(t,
		`SELECT "scope_morph_tags"."id","scope_morph_tags"."name" FROM "scope_morph_tags" INNER JOIN taggables ON taggables.scope_morph_tag_id = scope_morph_tags.id WHERE taggables.taggable_id = ? AND taggables.taggable_type = ? AND taggables.deleted_at IS NULL`,
		sql,
	)
}

// Sanity: when OnPivotQuery is nil, no extra WHERE is appended.
func TestRelated_Many2Many_WithoutOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 7}, "Roles")
	sql := newRelationSQL(t, rel, &[]relRole{})
	assert.Equal(t,
		`SELECT "rel_roles"."id","rel_roles"."name" FROM "rel_roles" INNER JOIN rel_user_roles ON rel_user_roles.rel_role_id = rel_roles.id WHERE rel_user_roles.rel_user_id = ?`,
		sql,
	)
}

// --- (2) Eager-load Many2Many pivot SELECT --------------------------------

// pivotEagerSQL replicates the *gormio.DB chain that loadMany2Many constructs for the pivot
// SELECT (eager_loader.go:435), then applies applyOnPivotQuery just like the production code
// does. Returns the DryRun SQL.
func pivotEagerSQL(t *testing.T, db *gormio.DB, desc *relationDescriptor, parentKeys []any) string {
	t.Helper()
	pivotParentCol := desc.pivotParentRef.foreignColumn
	pivotRelatedCol := desc.pivotRelatedRef.foreignColumn
	q := db.Session(&gormio.Session{NewDB: true}).
		Table(desc.pivotTable).
		Select(quoteIdent(pivotParentCol), quoteIdent(pivotRelatedCol)).
		Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(pivotParentCol)), parentKeys)
	q = applyOnPivotQuery(q, desc)
	stmt := q.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	return stmt.Statement.SQL.String()
}

func TestEagerLoad_Many2Many_PivotSELECT_WithOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &scopeUser{})
	desc, err := resolveRelation(q.instance, &scopeUser{}, "Roles")
	assert.NoError(t, err)

	sql := pivotEagerSQL(t, q.instance, desc, []any{uint(1), uint(2)})
	// The pivot SELECT must include the IN-clause on parent FK AND the OnPivotQuery scope.
	assert.Equal(t,
		`SELECT scope_user_id,scope_role_id FROM "scope_user_roles" WHERE scope_user_roles.scope_user_id IN (?,?) AND scope_user_roles.active = ?`,
		sql,
	)
}

// Sanity: without OnPivotQuery, the pivot SELECT has only the parent IN-clause.
func TestEagerLoad_Many2Many_PivotSELECT_WithoutOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	sql := pivotEagerSQL(t, q.instance, desc, []any{uint(1), uint(2)})
	assert.Equal(t,
		`SELECT rel_user_id,rel_role_id FROM "rel_user_roles" WHERE rel_user_roles.rel_user_id IN (?,?)`,
		sql,
	)
}

// --- (3) Eager-load MorphToMany pivot SELECT ------------------------------

// pivotEagerMorphSQL is the MorphToMany variant: same as pivotEagerSQL plus the morph-type
// WHERE that loadMorphToMany adds (eager_loader.go:571).
func pivotEagerMorphSQL(t *testing.T, db *gormio.DB, desc *relationDescriptor, parentKeys []any) string {
	t.Helper()
	pivotParentCol := desc.pivotParentRef.foreignColumn
	pivotRelatedCol := desc.pivotRelatedRef.foreignColumn
	q := db.Session(&gormio.Session{NewDB: true}).
		Table(desc.pivotTable).
		Select(quoteIdent(pivotParentCol), quoteIdent(pivotRelatedCol)).
		Where(fmt.Sprintf("%s.%s IN ?", quoteIdent(desc.pivotTable), quoteIdent(pivotParentCol)), parentKeys).
		Where(fmt.Sprintf("%s.%s = ?", quoteIdent(desc.pivotTable), quoteIdent(desc.morphTypeColumn)), desc.morphValue)
	q = applyOnPivotQuery(q, desc)
	stmt := q.Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	return stmt.Statement.SQL.String()
}

func TestEagerLoad_MorphToMany_PivotSELECT_WithOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &scopeMorphPost{})
	desc, err := resolveRelation(q.instance, &scopeMorphPost{}, "Tags")
	assert.NoError(t, err)

	sql := pivotEagerMorphSQL(t, q.instance, desc, []any{uint(1), uint(2)})
	// Order: parent IN clause, morph-type clause, then OnPivotQuery scope.
	assert.Equal(t,
		`SELECT taggable_id,scope_morph_tag_id FROM "taggables" WHERE taggables.taggable_id IN (?,?) AND taggables.taggable_type = ? AND taggables.deleted_at IS NULL`,
		sql,
	)
}

// Sanity: without OnPivotQuery, the morph pivot SELECT has only parent IN + morph-type WHERE.
func TestEagerLoad_MorphToMany_PivotSELECT_WithoutOnPivotQuery_SQL(t *testing.T) {
	q := newRelQueryWith(t, &morphPost{})
	desc, err := resolveRelation(q.instance, &morphPost{}, "Tags")
	assert.NoError(t, err)

	sql := pivotEagerMorphSQL(t, q.instance, desc, []any{uint(1), uint(2)})
	assert.Equal(t,
		`SELECT taggable_id,morph_tag_id FROM "taggables" WHERE taggables.taggable_id IN (?,?) AND taggables.taggable_type = ?`,
		sql,
	)
}
