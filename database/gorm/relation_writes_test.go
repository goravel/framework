package gorm

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	dbcontract "github.com/goravel/framework/contracts/database/db"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// setRelationFKOnChild is the FK-and-morph-type writer that SaveRelation calls before
// persistence. We test it directly because the stub dialector can't run the INSERT step.

func TestSetRelationFKOnChild_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)

	parent := &relUser{ID: 7}
	child := &relBook{Title: "x"}
	err = q.setRelationFKOnChild(parent, child, desc)
	assert.NoError(t, err)
	assert.Equal(t, uint(7), child.UserID)
}

func TestSetRelationFKOnChild_MorphMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Houses")
	assert.NoError(t, err)

	parent := &relUser{ID: 9}
	child := &relHouse{Address: "x"}
	err = q.setRelationFKOnChild(parent, child, desc)
	assert.NoError(t, err)
	assert.Equal(t, uint(9), child.HouseableID)
	assert.Equal(t, "rel_users", child.HouseableType)
}

func TestSetRelationFKOnChild_MorphOne(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Logo")
	assert.NoError(t, err)

	parent := &relUser{ID: 11}
	child := &relLogo{URL: "x"}
	err = q.setRelationFKOnChild(parent, child, desc)
	assert.NoError(t, err)
	assert.Equal(t, uint(11), child.LogoableID)
	assert.Equal(t, "rel_users", child.LogoableType)
}

// SaveRelation guard / dispatch tests. These don't reach the INSERT step (they error / return
// early before persistence), so they're safe to run against the stub dialector.

func TestSaveRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveRelation(relUser{ID: 1}, "Books", &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestSaveRelation_NilChild(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveRelation(&relUser{ID: 1}, "Books", nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestSaveRelation_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	err := q.SaveRelation(&relBook{}, "Author", &relUser{})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSaveRelation_UnsupportedKind_HasManyThrough(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	err := q.SaveRelation(&relCountry{}, "Posts", &relPost{})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSaveRelation_RelationNotFound(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveRelation(&relUser{}, "DoesNotExist", &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationNotFound))
}

func TestSaveManyRelation_NonSlice(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveManyRelation(&relUser{ID: 1}, "Books", "not a slice")
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

// Sanity: *Query satisfies the helper signatures the Orm wrapper relies on, and the
// contractsorm.Query interface is still satisfied (the new methods don't break that contract).
var _ interface {
	SaveRelation(parent any, relation string, child any) error
	SaveManyRelation(parent any, relation string, children any) error
	AssociateRelation(parent any, relation string, owner any) error
	DissociateRelation(parent any, relation string) error
} = (*Query)(nil)
var _ contractsorm.Query = (*Query)(nil)

// --- Sync / Toggle / UpdateExistingPivot ----------------------------------

func TestSyncRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.SyncRelation(relUser{}, "Roles", []any{1, 2})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestSyncRelation_UnsupportedKind(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.SyncRelation(&relUser{ID: 1}, "Books", []any{1})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSyncWithoutDetachingRelation_UnsupportedKind(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.SyncWithoutDetachingRelation(&relUser{ID: 1}, "Books", []any{1})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestToggleRelation_UnsupportedKind(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.ToggleRelation(&relUser{ID: 1}, "Books", []any{1})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestUpdateExistingPivotRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.UpdateExistingPivotRelation(relUser{}, "Roles", 1, map[string]any{"x": 1})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestUpdateExistingPivotRelation_UnsupportedKind(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.UpdateExistingPivotRelation(&relUser{ID: 1}, "Books", 1, map[string]any{"x": 1})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestUpdateExistingPivotRelation_EmptyAttrs_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rows, err := q.UpdateExistingPivotRelation(&relUser{ID: 1}, "Roles", 1, map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), rows)
}

func TestBasePivotRow_Many2Many(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	row := q.basePivotRow(desc, uint(7), uint(99), nil)
	assert.Equal(t, uint(7), row[desc.pivotParentRef.foreignColumn])
	assert.Equal(t, uint(99), row[desc.pivotRelatedRef.foreignColumn])
	_, hasMorphType := row[desc.morphTypeColumn]
	assert.False(t, hasMorphType, "pure Many2Many must not include morph_type")
}

func TestBasePivotRow_MorphToMany_IncludesType(t *testing.T) {
	q := newRelQueryWith(t, &morphPost{})
	desc, err := resolveRelation(q.instance, &morphPost{}, "Tags")
	assert.NoError(t, err)

	row := q.basePivotRow(desc, uint(3), uint(11), nil)
	assert.Equal(t, uint(3), row["taggable_id"])
	assert.Equal(t, "morph_posts", row["taggable_type"]) // table-name fallback
	assert.Equal(t, uint(11), row[desc.pivotRelatedRef.foreignColumn])
}

func TestBasePivotRow_AttrsOverlay(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	row := q.basePivotRow(desc, uint(7), uint(99), map[string]any{
		"priority": "high",
		"notes":    "x",
	})
	assert.Equal(t, "high", row["priority"])
	assert.Equal(t, "x", row["notes"])
	assert.Equal(t, uint(7), row[desc.pivotParentRef.foreignColumn])
}

func TestAttachRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AttachRelation(relUser{}, "Roles", []any{1})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestAttachRelation_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AttachRelation(&relUser{ID: 1}, "Books", []any{1})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestAttachRelation_EmptyIDs_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AttachRelation(&relUser{ID: 1}, "Roles", nil)
	assert.NoError(t, err)
}

func TestAttachWithPivotRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AttachWithPivotRelation(relUser{}, "Roles", map[any]map[string]any{1: {"priority": "high"}})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestAttachWithPivotRelation_EmptyMap_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AttachWithPivotRelation(&relUser{ID: 1}, "Roles", map[any]map[string]any{})
	assert.NoError(t, err)
}

func TestDetachRelation_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.DetachRelation(&relUser{ID: 1}, "Books", nil)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestDetachRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.DetachRelation(relUser{}, "Roles", nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestMutateAssociate_BelongsTo_SetsFK(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	desc, err := resolveRelation(q.instance, &relBook{}, "Author")
	assert.NoError(t, err)

	parent := &relBook{Title: "x", AuthorID: 0}
	owner := &relUser{ID: 42}
	err = q.mutateAssociate(parent, owner, desc, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(42), parent.AuthorID)
}

func TestMutateDissociate_BelongsTo_ClearsFK(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	desc, err := resolveRelation(q.instance, &relBook{}, "Author")
	assert.NoError(t, err)

	parent := &relBook{Title: "x", AuthorID: 99}
	err = q.mutateDissociate(parent, desc, false)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), parent.AuthorID)
}

func TestAssociateRelation_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	err := q.AssociateRelation(relBook{}, "Author", &relUser{ID: 1})
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestAssociateRelation_NilOwner(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	err := q.AssociateRelation(&relBook{}, "Author", nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestAssociateRelation_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.AssociateRelation(&relUser{}, "Books", &relBook{})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestDissociateRelation_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.DissociateRelation(&relUser{}, "Books")
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

// MorphTo Associate / Dissociate exercise the morph_type column path. Uses the morphImage
// fixture from relation_test.go.

func TestMutateAssociate_MorphTo_SetsFKAndType(t *testing.T) {
	q := newRelQueryWith(t, &morphImage{})
	desc, err := resolveRelation(q.instance, &morphImage{}, "Imageable")
	assert.NoError(t, err)

	parent := &morphImage{}
	owner := &relUser{ID: 3, Name: "n"}
	err = q.mutateAssociate(parent, owner, desc, true)
	assert.NoError(t, err)
	assert.Equal(t, uint(3), parent.ImageableID)
	assert.Equal(t, "rel_users", parent.ImageableType) // table-name fallback when not registered
}

func TestMutateDissociate_MorphTo_ClearsFKAndType(t *testing.T) {
	q := newRelQueryWith(t, &morphImage{})
	desc, err := resolveRelation(q.instance, &morphImage{}, "Imageable")
	assert.NoError(t, err)

	parent := &morphImage{ImageableID: 5, ImageableType: "post"}
	err = q.mutateDissociate(parent, desc, true)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), parent.ImageableID)
	assert.Equal(t, "", parent.ImageableType)
}

// Phase A/B tests: HasOneOrMany and BelongsToMany convenience methods

func TestCreateRelation_UnsupportedKind_HasManyThrough(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	err := q.CreateRelation(&relCountry{ID: 1}, "Posts", &relPost{})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestCreateRelation_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	err := q.CreateRelation(&relBook{ID: 1}, "Author", &relUser{})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestFindOrNewRelation_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	var dest relUser
	err := q.FindOrNewRelation(&relBook{ID: 1}, "Author", uint(5), &dest)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestFirstOrNewRelation_UnsupportedKind_Through(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	var dest relPost
	err := q.FirstOrNewRelation(&relCountry{ID: 1}, "Posts", map[string]any{"title": "x"}, nil, &dest)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestFirstOrCreateRelation_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	var dest relUser
	err := q.FirstOrCreateRelation(&relBook{ID: 1}, "Author", map[string]any{"name": "x"}, nil, &dest)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestUpdateOrCreateRelation_UnsupportedKind_Through(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	var dest relPost
	err := q.UpdateOrCreateRelation(&relCountry{ID: 1}, "Posts", map[string]any{"title": "x"}, map[string]any{"content": "y"}, &dest)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

// Phase C tests: PivotTimestamps

func TestBasePivotRow_Timestamps_IncludesCreatedUpdatedAt(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	// Enable timestamps by setting the resolved column names directly.
	desc.pivotCreatedAtColumn = "created_at"
	desc.pivotUpdatedAtColumn = "updated_at"

	row := q.basePivotRow(desc, uint(7), uint(99), nil)
	assert.Equal(t, uint(7), row[desc.pivotParentRef.foreignColumn])
	assert.Equal(t, uint(99), row[desc.pivotRelatedRef.foreignColumn])

	_, hasCreatedAt := row["created_at"]
	assert.True(t, hasCreatedAt, "pivot row must include created_at when desc.pivotCreatedAtColumn is set")

	_, hasUpdatedAt := row["updated_at"]
	assert.True(t, hasUpdatedAt, "pivot row must include updated_at when desc.pivotUpdatedAtColumn is set")
}

func TestBasePivotRow_Timestamps_AttrsCanOverride(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	desc.pivotCreatedAtColumn = "created_at"
	desc.pivotUpdatedAtColumn = "updated_at"

	customTime := "2024-01-01 00:00:00"
	row := q.basePivotRow(desc, uint(7), uint(99), map[string]any{
		"created_at": customTime,
	})

	assert.Equal(t, customTime, row["created_at"], "caller-supplied attrs must override timestamp")
	_, hasUpdatedAt := row["updated_at"]
	assert.True(t, hasUpdatedAt, "updated_at should still be set")
}

func TestBasePivotRow_NoTimestamps_OmitsBoth(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	// Empty resolved columns mean "don't auto-stamp".
	row := q.basePivotRow(desc, uint(7), uint(99), nil)
	_, hasCreatedAt := row["created_at"]
	_, hasUpdatedAt := row["updated_at"]
	assert.False(t, hasCreatedAt)
	assert.False(t, hasUpdatedAt)
}

func TestBasePivotRow_OnlyUpdatedAt_OmitsCreatedAt(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)

	// Pivot struct may declare only UpdatedAt (e.g. ledger-style writes); only stamp updated_at.
	desc.pivotUpdatedAtColumn = "updated_at"

	row := q.basePivotRow(desc, uint(7), uint(99), nil)
	_, hasCreatedAt := row["created_at"]
	_, hasUpdatedAt := row["updated_at"]
	assert.False(t, hasCreatedAt)
	assert.True(t, hasUpdatedAt)
}

// Phase G tests: SyncWithPivot / SyncWithPivotValues / ToggleWithPivot

func TestSyncRelationWithPivot_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.SyncRelationWithPivot(&relUser{ID: 1}, "Books", map[any]map[string]any{uint(1): {"priority": "high"}})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSyncRelationWithPivotValues_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	_, err := q.SyncRelationWithPivotValues(&relBook{ID: 1}, "Author", []any{uint(1)}, map[string]any{"priority": "high"})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSyncWithoutDetachingRelationWithPivot_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.SyncWithoutDetachingRelationWithPivot(&relUser{ID: 1}, "Books", map[any]map[string]any{uint(1): {"priority": "high"}})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestToggleRelationWithPivot_UnsupportedKind_HasMany(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	_, err := q.ToggleRelationWithPivot(&relUser{ID: 1}, "Books", map[any]map[string]any{uint(1): {"priority": "high"}})
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

// Phase H tests: castKey — normalises SyncResult ids back to the related model's PK type.

func TestCastKey(t *testing.T) {
	uintT := reflect.TypeFor[uint]()
	intT := reflect.TypeFor[int]()
	int64T := reflect.TypeFor[int64]()
	stringT := reflect.TypeFor[string]()
	float64T := reflect.TypeFor[float64]()

	cases := []struct {
		name string
		in   any
		t    reflect.Type
		want any
	}{
		{"nil value passthrough", nil, uintT, nil},
		{"nil type passthrough", uint(7), nil, uint(7)},
		{"same-type passthrough", uint(7), uintT, uint(7)},
		{"int -> uint", int(7), uintT, uint(7)},
		{"int64 -> uint (gorm scan typical)", int64(42), uintT, uint(42)},
		{"uint -> int", uint(7), intT, int(7)},
		{"uint -> int64", uint(7), int64T, int64(7)},
		{"float -> int", float64(3), intT, int(3)},
		{"string numeric -> uint", "42", uintT, uint(42)},
		{"string numeric -> int (negative)", "-7", intT, int(-7)},
		{"string numeric -> float", "3.14", float64T, float64(3.14)},
		{"int -> string (decimal, not Unicode)", int(65), stringT, "65"},
		{"uint -> string", uint(99), stringT, "99"},
		{"float -> string", float64(3.14), stringT, "3.14"},
		{"[]byte -> string", []byte("abc"), stringT, "abc"},
		{"non-numeric string -> uint passthrough", "abc", uintT, "abc"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := castKey(tt.in, tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCastKeys(t *testing.T) {
	uintT := reflect.TypeFor[uint]()

	t.Run("nil slice passthrough", func(t *testing.T) {
		assert.Nil(t, castKeys(nil, uintT))
	})
	t.Run("empty slice", func(t *testing.T) {
		assert.Equal(t, []any{}, castKeys([]any{}, uintT))
	})
	t.Run("mixed input types -> uniform uint", func(t *testing.T) {
		got := castKeys([]any{int(1), int64(2), "3", uint(4)}, uintT)
		assert.Equal(t, []any{uint(1), uint(2), uint(3), uint(4)}, got)
	})
	t.Run("nil keyType leaves values untouched", func(t *testing.T) {
		got := castKeys([]any{int(1), "2"}, nil)
		assert.Equal(t, []any{int(1), "2"}, got)
	})
}

// relatedKeyType wiring: descriptor must carry the related model's PK type so castKey can use it.
func TestDescriptor_RelatedKeyType_Many2Many(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Roles")
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeFor[uint](), desc.relatedKeyType)
}

func TestDescriptor_RelatedKeyType_MorphToMany(t *testing.T) {
	q := newRelQueryWith(t, &morphPost{})
	desc, err := resolveRelation(q.instance, &morphPost{}, "Tags")
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeFor[uint](), desc.relatedKeyType)
}

// Pivot timestamp resolution tests — exercise the priority order between Pivot struct
// autoCreateTime/autoUpdateTime tags, CreatedAt/UpdatedAt convention, and the relation-level
// PivotTimestamps fallback.

type tsTaggedPivot struct {
	UserID  uint      `gorm:"column:user_id"`
	RoleID  uint      `gorm:"column:role_id"`
	Stamped time.Time `gorm:"autoCreateTime"`
	Edited  time.Time `gorm:"autoUpdateTime"`
}

type tsTaggedRole struct {
	ID    uint
	Name  string
	Pivot tsTaggedPivot `gorm:"-"`
}

type tsConventionPivot struct {
	UserID    uint `gorm:"column:user_id"`
	RoleID    uint `gorm:"column:role_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type tsConventionRole struct {
	ID    uint
	Name  string
	Pivot tsConventionPivot `gorm:"-"`
}

type tsCreatedOnlyPivot struct {
	UserID    uint `gorm:"column:user_id"`
	RoleID    uint `gorm:"column:role_id"`
	CreatedAt time.Time
}

type tsCreatedOnlyRole struct {
	ID    uint
	Pivot tsCreatedOnlyPivot `gorm:"-"`
}

type tsCustomColumnPivot struct {
	UserID  uint      `gorm:"column:user_id"`
	RoleID  uint      `gorm:"column:role_id"`
	Stamped time.Time `gorm:"autoCreateTime;column:made_on"`
	Edited  time.Time `gorm:"autoUpdateTime;column:edited_at"`
}

type tsCustomColumnRole struct {
	ID    uint
	Pivot tsCustomColumnPivot `gorm:"-"`
}

func TestResolvePivotTimestamps_AutoCreateTimeTag(t *testing.T) {
	db := newStubGormDB(t)
	created, updated, err := resolvePivotTimestamps(db, &tsTaggedRole{}, "Pivot", false)
	assert.NoError(t, err)
	assert.Equal(t, "stamped", created)
	assert.Equal(t, "edited", updated)
}

func TestResolvePivotTimestamps_Convention(t *testing.T) {
	db := newStubGormDB(t)
	created, updated, err := resolvePivotTimestamps(db, &tsConventionRole{}, "Pivot", false)
	assert.NoError(t, err)
	assert.Equal(t, "created_at", created)
	assert.Equal(t, "updated_at", updated)
}

func TestResolvePivotTimestamps_OnlyCreatedAt(t *testing.T) {
	db := newStubGormDB(t)
	created, updated, err := resolvePivotTimestamps(db, &tsCreatedOnlyRole{}, "Pivot", false)
	assert.NoError(t, err)
	assert.Equal(t, "created_at", created)
	assert.Equal(t, "", updated, "no UpdatedAt field means don't auto-stamp on update")
}

func TestResolvePivotTimestamps_CustomColumnTag(t *testing.T) {
	db := newStubGormDB(t)
	created, updated, err := resolvePivotTimestamps(db, &tsCustomColumnRole{}, "Pivot", false)
	assert.NoError(t, err)
	assert.Equal(t, "made_on", created)
	assert.Equal(t, "edited_at", updated)
}

func TestResolvePivotTimestamps_NoStruct_FallbackEnabled(t *testing.T) {
	db := newStubGormDB(t)
	// roleWithoutPivot has no Pivot field — falls through to relation-level PivotTimestamps.
	created, updated, err := resolvePivotTimestamps(db, &roleWithoutPivot{}, "Pivot", true)
	assert.NoError(t, err)
	assert.Equal(t, "created_at", created)
	assert.Equal(t, "updated_at", updated)
}

func TestResolvePivotTimestamps_NoStruct_FallbackDisabled(t *testing.T) {
	db := newStubGormDB(t)
	created, updated, err := resolvePivotTimestamps(db, &roleWithoutPivot{}, "Pivot", false)
	assert.NoError(t, err)
	assert.Equal(t, "", created)
	assert.Equal(t, "", updated)
}

func TestResolvePivotTimestamps_StructHasOneCol_FallbackFillsOther(t *testing.T) {
	db := newStubGormDB(t)
	// Struct provides only CreatedAt; PivotTimestamps: true fills updated_at default.
	created, updated, err := resolvePivotTimestamps(db, &tsCreatedOnlyRole{}, "Pivot", true)
	assert.NoError(t, err)
	assert.Equal(t, "created_at", created, "struct-provided column wins")
	assert.Equal(t, "updated_at", updated, "fallback fills the column the struct didn't provide")
}

// syncResultChanged is a pure function used by syncCore/syncCoreWithPivot to decide whether to
// call touchIfTouching. Test all branches.

func TestSyncResultChanged_AllEmpty(t *testing.T) {
	out := &dbcontract.SyncResult{}
	assert.False(t, syncResultChanged(out))
}

func TestSyncResultChanged_HasAttached(t *testing.T) {
	out := &dbcontract.SyncResult{Attached: []any{1}}
	assert.True(t, syncResultChanged(out))
}

func TestSyncResultChanged_HasDetached(t *testing.T) {
	out := &dbcontract.SyncResult{Detached: []any{2}}
	assert.True(t, syncResultChanged(out))
}

func TestSyncResultChanged_HasUpdated(t *testing.T) {
	out := &dbcontract.SyncResult{Updated: []any{3}}
	assert.True(t, syncResultChanged(out))
}

func TestSyncResultChanged_AllPopulated(t *testing.T) {
	out := &dbcontract.SyncResult{Attached: []any{1}, Detached: []any{2}, Updated: []any{3}}
	assert.True(t, syncResultChanged(out))
}

// applyAttrMap overlays an attrs map onto a target struct via GORM's schema. Tests cover the
// happy path, the early-return for empty attrs, and the silent skip for unknown columns.

func TestApplyAttrMap_EmptyAttrs_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	dest := &relUser{ID: 5}
	err := q.applyAttrMap(dest, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint(5), dest.ID, "dest unchanged")
}

func TestApplyAttrMap_EmptyMap_NoOp(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	dest := &relUser{ID: 5}
	err := q.applyAttrMap(dest, map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, uint(5), dest.ID, "dest unchanged")
}

func TestApplyAttrMap_UnknownColumn_Skipped(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	dest := &relUser{ID: 5}
	// "no_such_column" doesn't map to any field — applyAttrMap silently skips.
	err := q.applyAttrMap(dest, map[string]any{"no_such_column": "x"})
	assert.NoError(t, err)
	assert.Equal(t, uint(5), dest.ID, "dest unchanged")
}

// CreateRelation additional coverage for Many2Many path

// FindOrNewRelation additional coverage

// FirstOrCreateRelation additional coverage for Many2Many path

func TestKindName_AllKinds(t *testing.T) {
	assert.Equal(t, "hasOne", kindName(relKindHasOne))
	assert.Equal(t, "hasMany", kindName(relKindHasMany))
	assert.Equal(t, "belongsTo", kindName(relKindBelongsTo))
	assert.Equal(t, "many2Many", kindName(relKindMany2Many))
	assert.Equal(t, "morphOne", kindName(relKindMorphOne))
	assert.Equal(t, "morphMany", kindName(relKindMorphMany))
	assert.Equal(t, "morphTo", kindName(relKindMorphTo))
	assert.Equal(t, "morphToMany", kindName(relKindMorphToMany))
	assert.Equal(t, "hasOneThrough", kindName(relKindHasOneThrough))
	assert.Equal(t, "hasManyThrough", kindName(relKindHasManyThrough))
	assert.Equal(t, "kind=999", kindName(999))
}

// SaveRelationWithPivot coverage

func TestSaveRelationWithPivot_UnsupportedKind_BelongsTo(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	err := q.SaveRelationWithPivot(&relBook{ID: 1}, "Author", &relUser{}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSaveRelationWithPivot_NotPointerParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveRelationWithPivot(relUser{}, "Roles", &relRole{}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

func TestSaveRelationWithPivot_NilChild(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveRelationWithPivot(&relUser{ID: 1}, "Roles", nil, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationParentNotPointer))
}

// SaveManyRelationWithPivot coverage

func TestSaveManyRelationWithPivot_NonSlice(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.SaveManyRelationWithPivot(&relUser{ID: 1}, "Roles", "not a slice", nil)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

func TestSaveManyRelationWithPivot_InvalidElement(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	// Slice of non-struct/non-pointer elements
	err := q.SaveManyRelationWithPivot(&relUser{ID: 1}, "Roles", []int{1, 2, 3}, nil)
	assert.True(t, errors.Is(err, errors.OrmRelationKindNotSupported))
}

// CreateManyRelation coverage

func TestCreateManyRelation_NonSlice(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	err := q.CreateManyRelation(&relUser{ID: 1}, "Books", "not a slice")
	assert.True(t, errors.Is(err, errors.OrmRelationUnsupported))
}

// Additional error path coverage
