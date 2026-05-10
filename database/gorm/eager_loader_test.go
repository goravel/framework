package gorm

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	gormschema "gorm.io/gorm/schema"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/errors"
)

// --- Pure helpers ----------------------------------------------------------

func TestDictKey(t *testing.T) {
	cases := []struct {
		name  string
		input any
		want  string
	}{
		{"nil", nil, ""},
		{"string", "abc", "abc"},
		{"bytes", []byte("abc"), "abc"},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"uint", uint(7), "7"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, dictKey(tc.input))
		})
	}
}

func TestChunkSize(t *testing.T) {
	q := newRelQuery(t)
	// nil config -> default
	assert.Equal(t, defaultEagerLoadChunkSize, q.chunkSize())
}

// --- Reflect helpers ------------------------------------------------------

func TestCollectEagerParentsStructPtr(t *testing.T) {
	u := &relUser{ID: 1}
	out, err := collectEagerParents(u)
	assert.NoError(t, err)
	assert.Len(t, out, 1)
}

func TestCollectEagerParentsSlice(t *testing.T) {
	users := []relUser{{ID: 1}, {ID: 2}}
	out, err := collectEagerParents(&users)
	assert.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestCollectEagerParentsSliceOfPtr(t *testing.T) {
	users := []*relUser{{ID: 1}, nil, {ID: 2}}
	out, err := collectEagerParents(&users)
	assert.NoError(t, err)
	assert.Len(t, out, 2)
}

func TestCollectEagerParentsNil(t *testing.T) {
	out, err := collectEagerParents(nil)
	assert.NoError(t, err)
	assert.Nil(t, out)

	var p *relUser
	out, err = collectEagerParents(p)
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestCollectEagerParentsNotPointer(t *testing.T) {
	out, err := collectEagerParents(relUser{})
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestCollectEagerParentsUnsupportedKind(t *testing.T) {
	v := 7
	out, err := collectEagerParents(&v)
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestNewSampleModel(t *testing.T) {
	u := relUser{ID: 1}
	rv := reflect.ValueOf(u)
	got := newSampleModel(rv)
	rt := reflect.TypeOf(got)
	assert.Equal(t, reflect.Pointer, rt.Kind())
	assert.Equal(t, "relUser", rt.Elem().Name())
	// Should be a fresh zero instance (not the original).
	assert.Equal(t, uint(0), got.(*relUser).ID)
}

func TestParseGormSchema(t *testing.T) {
	db := newStubGormDB(t)
	s, err := parseGormSchema(db, &relUser{})
	assert.NoError(t, err)
	assert.Equal(t, "rel_users", s.Table)

	_, err = parseGormSchema(db, "bad-model")
	assert.Error(t, err)
}

func TestExtractKeysDeduplicatesAndSkipsZero(t *testing.T) {
	db := newStubGormDB(t)
	s, err := parseGormSchema(db, &relUser{})
	assert.NoError(t, err)
	idField := s.FieldsByDBName["id"]
	assert.NotNil(t, idField)

	q := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, &Conditions{})

	parents := []reflect.Value{
		reflect.ValueOf(relUser{ID: 1}),
		reflect.ValueOf(relUser{ID: 2}),
		reflect.ValueOf(relUser{ID: 1}), // dup
		reflect.ValueOf(relUser{ID: 0}), // zero - skipped
	}
	keys := extractKeys(q, parents, idField)
	assert.Len(t, keys, 2)
}

// --- setRelationField ------------------------------------------------------

type withPtrRel struct {
	ID      uint
	Profile *relProfile
}

type withSlicePtrRel struct {
	ID    uint
	Books []*relBook
}

type withSliceStructRel struct {
	ID    uint
	Books []relBook
}

func TestSetRelationField_PtrAssignment(t *testing.T) {
	parent := withPtrRel{}
	rv := reflect.ValueOf(&parent).Elem()
	row := reflect.ValueOf(&relProfile{Bio: "x"})
	err := setRelationField(rv, "Profile", []reflect.Value{row})
	assert.NoError(t, err)
	assert.Equal(t, "x", parent.Profile.Bio)
}

func TestSetRelationField_PtrEmptyClearsField(t *testing.T) {
	parent := withPtrRel{Profile: &relProfile{Bio: "stale"}}
	rv := reflect.ValueOf(&parent).Elem()
	err := setRelationField(rv, "Profile", nil)
	assert.NoError(t, err)
	assert.Nil(t, parent.Profile)
}

func TestSetRelationField_SliceOfPtrs(t *testing.T) {
	parent := withSlicePtrRel{}
	rv := reflect.ValueOf(&parent).Elem()
	rows := []reflect.Value{
		reflect.ValueOf(&relBook{Title: "a"}),
		reflect.ValueOf(&relBook{Title: "b"}),
	}
	err := setRelationField(rv, "Books", rows)
	assert.NoError(t, err)
	assert.Len(t, parent.Books, 2)
}

func TestSetRelationField_SliceOfStructs(t *testing.T) {
	parent := withSliceStructRel{}
	rv := reflect.ValueOf(&parent).Elem()
	rows := []reflect.Value{
		reflect.ValueOf(&relBook{Title: "a"}),
		reflect.ValueOf(&relBook{Title: "b"}),
	}
	err := setRelationField(rv, "Books", rows)
	assert.NoError(t, err)
	assert.Len(t, parent.Books, 2)
	assert.Equal(t, "a", parent.Books[0].Title)
}

func TestSetRelationField_UnknownField(t *testing.T) {
	parent := withPtrRel{}
	rv := reflect.ValueOf(&parent).Elem()
	err := setRelationField(rv, "Missing", nil)
	assert.True(t, errors.Is(err, errors.OrmEagerLoadCannotAssign))
}

// withInterfaceRel exercises the MorphTo field shape: an `any` field that the loader fills with
// a *RelatedModel value chosen at runtime via the morph map.
type withInterfaceRel struct {
	ID        uint
	Imageable any
}

func TestSetRelationField_InterfaceAssignment(t *testing.T) {
	parent := withInterfaceRel{}
	rv := reflect.ValueOf(&parent).Elem()
	row := reflect.ValueOf(&relBook{Title: "x"})
	err := setRelationField(rv, "Imageable", []reflect.Value{row})
	assert.NoError(t, err)
	got, ok := parent.Imageable.(*relBook)
	assert.True(t, ok)
	assert.Equal(t, "x", got.Title)
}

func TestSetRelationField_InterfaceEmptyClearsField(t *testing.T) {
	parent := withInterfaceRel{Imageable: &relBook{Title: "stale"}}
	rv := reflect.ValueOf(&parent).Elem()
	err := setRelationField(rv, "Imageable", nil)
	assert.NoError(t, err)
	assert.Nil(t, parent.Imageable)
}

// --- runEagerLoads no-op paths --------------------------------------------

func TestRunEagerLoadsNoParents(t *testing.T) {
	q := newRelQuery(t)
	err := q.runEagerLoads(nil, []eagerLoadEntry{{relation: "Books"}})
	assert.NoError(t, err)
}

func TestRunEagerLoadsNoEntries(t *testing.T) {
	q := newRelQuery(t)
	parents := []reflect.Value{reflect.ValueOf(relUser{ID: 1})}
	err := q.runEagerLoads(parents, nil)
	assert.NoError(t, err)
}

func TestApplyEagerLoadsNothingQueued(t *testing.T) {
	q := newRelQuery(t)
	users := &[]relUser{}
	err := q.applyEagerLoads(users)
	assert.NoError(t, err)
}

func TestApplyEagerLoadsEmptyDest(t *testing.T) {
	q := newRelQuery(t)
	q.conditions.eagerLoad = []eagerLoadEntry{{relation: "Books"}}
	users := &[]relUser{}
	err := q.applyEagerLoads(users)
	assert.NoError(t, err)
}

func TestRecurseNestedNoop(t *testing.T) {
	q := newRelQuery(t)
	err := q.recurseNested(nil, []eagerLoadEntry{{relation: "X"}})
	assert.NoError(t, err)
	err = q.recurseNested([]reflect.Value{reflect.ValueOf(relUser{})}, nil)
	assert.NoError(t, err)
}

func TestMaybeRecurseEmpty_NotMany(t *testing.T) {
	q := newRelQuery(t)
	err := q.maybeRecurseEmpty(nil, "X", false, nil)
	assert.NoError(t, err)
}

func TestMaybeRecurseEmpty_ManyAssignsEmptySlices(t *testing.T) {
	q := newRelQuery(t)
	u1 := &withSlicePtrRel{ID: 1, Books: []*relBook{{Title: "a"}}}
	u2 := &withSlicePtrRel{ID: 2}
	parents := []reflect.Value{reflect.ValueOf(u1).Elem(), reflect.ValueOf(u2).Elem()}
	err := q.maybeRecurseEmpty(parents, "Books", true, nil)
	assert.NoError(t, err)
	assert.Empty(t, u1.Books)
	assert.Empty(t, u2.Books)
}

// --- Phase D: Pivot column hydration tests --------------------------------

// roleUserPivot is a sample custom Pivot model used by the struct-only Pivot tests below. The
// gorm tags lock column names so the test data (keyed by db column) maps deterministically into
// struct fields.
type roleUserPivot struct {
	UserID   uint   `gorm:"column:user_id"`
	RoleID   uint   `gorm:"column:role_id"`
	Priority string `gorm:"column:priority"`
	Notes    string `gorm:"column:notes"`
}

type roleWithPivot struct {
	ID    uint
	Name  string
	Pivot roleUserPivot `gorm:"-"`
}

type roleWithoutPivot struct {
	ID   uint
	Name string
}

// roleWithCustomPivotField has the pivot data on a non-default field name — exercises
// PivotField configuration.
type roleWithCustomPivotField struct {
	ID        uint
	Name      string
	UserPivot roleUserPivot `gorm:"-"`
}

// roleWithBadPivot has a Pivot field that isn't a struct — exercises the field-not-struct guard.
type roleWithBadPivot struct {
	ID    uint
	Name  string
	Pivot string `gorm:"-"`
}

func TestWritePivotField_HydratesStructField(t *testing.T) {
	role := &roleWithPivot{ID: 1, Name: "admin"}
	plan := mustPivotPlan(t, "Pivot", &roleUserPivot{})
	err := writePivotField(t.Context(), reflect.ValueOf(role), map[string]any{
		"priority": "high",
		"notes":    "test",
	}, plan)
	assert.NoError(t, err)
	assert.Equal(t, "high", role.Pivot.Priority)
	assert.Equal(t, "test", role.Pivot.Notes)
}

func TestWritePivotField_TypedColumns(t *testing.T) {
	role := &roleWithPivot{ID: 1}
	plan := mustPivotPlan(t, "Pivot", &roleUserPivot{})
	err := writePivotField(t.Context(), reflect.ValueOf(role), map[string]any{
		"user_id": uint(7),
		"role_id": uint(99),
	}, plan)
	assert.NoError(t, err)
	assert.Equal(t, uint(7), role.Pivot.UserID)
	assert.Equal(t, uint(99), role.Pivot.RoleID)
}

func TestWritePivotField_CustomFieldName(t *testing.T) {
	role := &roleWithCustomPivotField{ID: 1}
	plan := mustPivotPlan(t, "UserPivot", &roleUserPivot{})
	err := writePivotField(t.Context(), reflect.ValueOf(role), map[string]any{
		"user_id":  uint(7),
		"priority": "high",
	}, plan)
	assert.NoError(t, err)
	assert.Equal(t, uint(7), role.UserPivot.UserID)
	assert.Equal(t, "high", role.UserPivot.Priority)
}

func TestWritePivotField_NoPivotField_ReturnsNil(t *testing.T) {
	role := &roleWithoutPivot{ID: 1, Name: "admin"}
	plan := mustPivotPlan(t, "Pivot", &roleUserPivot{})
	err := writePivotField(t.Context(), reflect.ValueOf(role), map[string]any{"priority": "high"}, plan)
	assert.NoError(t, err, "writePivotField must silently skip when configured field is absent")
}

func TestWritePivotField_UnknownColumn_Skipped(t *testing.T) {
	role := &roleWithPivot{ID: 1}
	plan := mustPivotPlan(t, "Pivot", &roleUserPivot{})
	err := writePivotField(t.Context(), reflect.ValueOf(role), map[string]any{
		"priority":   "high",
		"unknown_xy": "ignored",
	}, plan)
	assert.NoError(t, err)
	assert.Equal(t, "high", role.Pivot.Priority)
}

func TestPreparePivotHydration_NoFieldOnRelated_ReturnsNil(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		relatedModel: &roleWithoutPivot{},
		pivotField:   "Pivot",
	}
	plan, err := preparePivotHydration(q, desc)
	assert.NoError(t, err)
	assert.Nil(t, plan, "no field by configured name means nothing to hydrate")
}

func TestPreparePivotHydration_FieldNotStruct_Errors(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		relatedModel: &roleWithBadPivot{},
		pivotField:   "Pivot",
	}
	_, err := preparePivotHydration(q, desc)
	assert.True(t, errors.Is(err, errors.OrmRelationPivotFieldNotStruct))
}

func TestPreparePivotHydration_DefaultPivot_FromFieldType(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		relatedModel: &roleWithPivot{},
		pivotField:   "Pivot",
	}
	plan, err := preparePivotHydration(q, desc)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "Pivot", plan.fieldName)
	// Selected columns include every db-tagged field on roleUserPivot.
	assert.ElementsMatch(t, []string{"user_id", "role_id", "priority", "notes"}, plan.extraColumns)
}

func TestPreparePivotHydration_CustomFieldName_FromFieldType(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	desc := &relationDescriptor{
		relatedModel: &roleWithCustomPivotField{},
		pivotField:   "UserPivot",
	}
	plan, err := preparePivotHydration(q, desc)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "UserPivot", plan.fieldName)
}

// mustPivotPlan builds a pivotHydrationPlan for use in writePivotField tests, bypassing
// preparePivotHydration so we can exercise writePivotField independently.
func mustPivotPlan(t *testing.T, fieldName string, pivotProto any) *pivotHydrationPlan {
	t.Helper()
	q := newRelQueryWith(t, &relUser{})
	usingSchema, err := parseGormSchema(q.instance, pivotProto)
	if err != nil {
		t.Fatalf("parse pivot schema: %v", err)
	}
	cols := make([]string, 0, len(usingSchema.Fields))
	byCol := make(map[string]*gormschema.Field, len(usingSchema.Fields))
	for _, f := range usingSchema.Fields {
		if f.DBName == "" {
			continue
		}
		cols = append(cols, f.DBName)
		byCol[f.DBName] = f
	}
	return &pivotHydrationPlan{
		fieldName:     fieldName,
		extraColumns:  cols,
		fieldByColumn: byCol,
	}
}
