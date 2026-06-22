package gorm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// dryRunPivotSQL renders the SQL emitted by a SELECT 1 FROM <table> ... query that pq has
// already mutated, so test cases can assert on exact WHERE-clause shape.
func dryRunPivotSQL(t *testing.T, db *gormio.DB, table string) string {
	t.Helper()
	stmt := db.Session(&gormio.Session{DryRun: true}).Table(table).Find(&[]map[string]any{})
	return stmt.Statement.SQL.String()
}

func TestPivotQuery_Where_TwoArgs_DefaultsToEqual(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.Where("active", 1)
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.active = ?")
}

func TestPivotQuery_Where_ThreeArgs_UsesOperator(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.Where("priority", ">=", 5)
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.priority >= ?")
}

func TestPivotQuery_WhereIn(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.WhereIn("scope", []any{"team", "global"})
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.scope IN (?,?)")
}

func TestPivotQuery_WhereNotIn(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.WhereNotIn("scope", []any{"archived"})
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.scope NOT IN (?)")
}

func TestPivotQuery_WhereNull(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.WhereNull("deleted_at")
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.deleted_at IS NULL")
}

func TestPivotQuery_WhereNotNull(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.WhereNotNull("expires_at")
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.expires_at IS NOT NULL")
}

func TestPivotQuery_Chained_AccumulatesAllClauses(t *testing.T) {
	db := newStubGormDB(t)
	pq := newPivotQuery(db.Table("rel_user_roles"), "rel_user_roles")
	pq.Where("active", 1).
		WhereIn("scope", []any{"team", "global"}).
		WhereNull("deleted_at")
	sql := dryRunPivotSQL(t, pq.db, "rel_user_roles")
	// All three predicates should be ANDed into the final query.
	assert.Contains(t, sql, "rel_user_roles.active = ?")
	assert.Contains(t, sql, "rel_user_roles.scope IN (?,?)")
	assert.Contains(t, sql, "rel_user_roles.deleted_at IS NULL")
	// Naive count: three "AND" or three predicates joined.
	assert.Equal(t, 2, strings.Count(sql, " AND "), "three predicates should be joined by two ANDs")
}

func TestApplyOnPivotQuery_NilCallback_NoOp(t *testing.T) {
	db := newStubGormDB(t)
	desc := &relationDescriptor{pivotTable: "rel_user_roles", onPivotQuery: nil}
	q := db.Table("rel_user_roles")
	out := applyOnPivotQuery(q, desc)
	assert.Same(t, q, out, "nil callback must return the original query unchanged")
}

func TestApplyOnPivotQuery_AppliesCallback(t *testing.T) {
	db := newStubGormDB(t)
	desc := &relationDescriptor{
		pivotTable: "rel_user_roles",
		onPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
			return q.Where("active", 1)
		},
	}
	q := db.Table("rel_user_roles")
	out := applyOnPivotQuery(q, desc)
	sql := dryRunPivotSQL(t, out, "rel_user_roles")
	assert.Contains(t, sql, "rel_user_roles.active = ?")
}

// Descriptor wiring: OnPivotQuery declared on Many2Many / MorphToMany / MorphedByMany must land
// on the relationDescriptor.
type pivotWireUser struct {
	ID    uint
	Roles []*relRole `gorm:"-"`
}

func (pivotWireUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Roles": contractsorm.Many2Many{
			Related: &relRole{},
			Table:   "pivot_wire_user_roles",
			OnPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
				return q.Where("active", 1)
			},
		},
	}
}

func TestDescriptor_OnPivotQuery_Many2Many(t *testing.T) {
	q := newRelQueryWith(t, &pivotWireUser{})
	desc, err := resolveRelation(q.instance, &pivotWireUser{}, "Roles")
	assert.NoError(t, err)
	assert.NotNil(t, desc.onPivotQuery, "Many2Many.OnPivotQuery must land on descriptor")

	// Verify the callback actually runs.
	pq := newPivotQuery(newStubGormDB(t).Table("pivot_wire_user_roles"), "pivot_wire_user_roles")
	desc.onPivotQuery(pq)
	sql := dryRunPivotSQL(t, pq.db, "pivot_wire_user_roles")
	assert.Contains(t, sql, "pivot_wire_user_roles.active = ?")
}

type pivotWireMorphPost struct {
	ID   uint
	Tags []*morphTag `gorm:"-"`
}

func (pivotWireMorphPost) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Tags": contractsorm.MorphToMany{
			Related: &morphTag{},
			Name:    "taggable",
			OnPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
				return q.WhereNull("deleted_at")
			},
		},
	}
}

func TestDescriptor_OnPivotQuery_MorphToMany(t *testing.T) {
	q := newRelQueryWith(t, &pivotWireMorphPost{})
	desc, err := resolveRelation(q.instance, &pivotWireMorphPost{}, "Tags")
	assert.NoError(t, err)
	assert.NotNil(t, desc.onPivotQuery, "MorphToMany.OnPivotQuery must land on descriptor")
}

type pivotWireTag struct {
	ID    uint
	Posts []*morphPost `gorm:"-"`
}

func (pivotWireTag) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Posts": contractsorm.MorphedByMany{
			Related: &morphPost{},
			Name:    "taggable",
			OnPivotQuery: func(q contractsorm.PivotQuery) contractsorm.PivotQuery {
				return q.Where("verified", 1)
			},
		},
	}
}

func TestDescriptor_OnPivotQuery_MorphedByMany(t *testing.T) {
	q := newRelQueryWith(t, &pivotWireTag{})
	desc, err := resolveRelation(q.instance, &pivotWireTag{}, "Posts")
	assert.NoError(t, err)
	assert.NotNil(t, desc.onPivotQuery, "MorphedByMany.OnPivotQuery must land on descriptor")
}

// Touches wiring — the relation declarations carry Touches: true through to the descriptor.

type touchesUser struct {
	ID    uint
	Roles []*relRole `gorm:"-"`
}

func (touchesUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Roles": contractsorm.Many2Many{Related: &relRole{}, Table: "touches_user_roles", Touches: true},
	}
}

func TestDescriptor_Touches_Many2Many(t *testing.T) {
	q := newRelQueryWith(t, &touchesUser{})
	desc, err := resolveRelation(q.instance, &touchesUser{}, "Roles")
	assert.NoError(t, err)
	assert.True(t, desc.touches)
}

type touchesPost struct {
	ID   uint
	Tags []*morphTag `gorm:"-"`
}

func (touchesPost) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Tags": contractsorm.MorphToMany{Related: &morphTag{}, Name: "taggable", Touches: true},
	}
}

func TestDescriptor_Touches_MorphToMany(t *testing.T) {
	q := newRelQueryWith(t, &touchesPost{})
	desc, err := resolveRelation(q.instance, &touchesPost{}, "Tags")
	assert.NoError(t, err)
	assert.True(t, desc.touches)
}

type touchesTag struct {
	ID    uint
	Posts []*morphPost `gorm:"-"`
}

func (touchesTag) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Posts": contractsorm.MorphedByMany{Related: &morphPost{}, Name: "taggable", Touches: true},
	}
}

func TestDescriptor_Touches_MorphedByMany(t *testing.T) {
	q := newRelQueryWith(t, &touchesTag{})
	desc, err := resolveRelation(q.instance, &touchesTag{}, "Posts")
	assert.NoError(t, err)
	assert.True(t, desc.touches)
}

// Default: when Touches is omitted, descriptor.touches is false. Sanity guard against accidental
// always-on behavior.
func TestDescriptor_Touches_DefaultsFalse(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)
	assert.False(t, desc.touches)
}

func TestTouchIfTouching_DescTouchesFalse_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{touches: false}
	// parent ptr is nil-safe here because the function returns before dereferencing.
	err := q.touchIfTouching(desc, &relUser{ID: 7}, uint(7))
	assert.NoError(t, err)
}

func TestTouchIfTouching_NoUpdatedAtField_NoOp(t *testing.T) {
	// relRole has no UpdatedAt field — touchIfTouching must silently skip even when touches=true.
	q := newRelQueryWith(t, &relRole{})
	desc := &relationDescriptor{touches: true}
	err := q.touchIfTouching(desc, &relRole{ID: 1}, uint(1))
	assert.NoError(t, err)
}

// PivotField wiring — defaults to "Pivot" when omitted, takes the user-specified value otherwise.

type pivotFieldUser struct {
	ID    uint
	Roles []*relRole `gorm:"-"`
}

func (pivotFieldUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Roles":          contractsorm.Many2Many{Related: &relRole{}, Table: "pf_user_roles"},
		"AuditedRoles":   contractsorm.Many2Many{Related: &relRole{}, Table: "pf_audit_roles", PivotField: "AuditPivot"},
		"TaggedPosts":    contractsorm.MorphToMany{Related: &morphTag{}, Name: "taggable", PivotField: "TagPivot"},
		"InversedTagged": contractsorm.MorphedByMany{Related: &morphPost{}, Name: "taggable", PivotField: "InversePivot"},
	}
}

func TestDescriptor_PivotField_DefaultsToPivot(t *testing.T) {
	q := newRelQueryWith(t, &pivotFieldUser{})
	desc, err := resolveRelation(q.instance, &pivotFieldUser{}, "Roles")
	assert.NoError(t, err)
	assert.Equal(t, "Pivot", desc.pivotField)
}

func TestDescriptor_PivotField_CustomMany2Many(t *testing.T) {
	q := newRelQueryWith(t, &pivotFieldUser{})
	desc, err := resolveRelation(q.instance, &pivotFieldUser{}, "AuditedRoles")
	assert.NoError(t, err)
	assert.Equal(t, "AuditPivot", desc.pivotField)
}

func TestDescriptor_PivotField_CustomMorphToMany(t *testing.T) {
	q := newRelQueryWith(t, &pivotFieldUser{})
	desc, err := resolveRelation(q.instance, &pivotFieldUser{}, "TaggedPosts")
	assert.NoError(t, err)
	assert.Equal(t, "TagPivot", desc.pivotField)
}

func TestDescriptor_PivotField_CustomMorphedByMany(t *testing.T) {
	q := newRelQueryWith(t, &pivotFieldUser{})
	desc, err := resolveRelation(q.instance, &pivotFieldUser{}, "InversedTagged")
	assert.NoError(t, err)
	assert.Equal(t, "InversePivot", desc.pivotField)
}
