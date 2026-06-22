package gorm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	gormschema "gorm.io/gorm/schema"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

// stubDialector is a no-op dialector that lets us spin up a *gormio.DB without an actual
// connection. It registers the standard callbacks so DryRun-mode SQL can still be built.
type stubDialector struct{}

func (stubDialector) Name() string { return "stub" }
func (stubDialector) Initialize(db *gormio.DB) error {
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})
	return nil
}
func (stubDialector) Migrator(db *gormio.DB) gormio.Migrator             { return nil }
func (stubDialector) DataTypeOf(*gormschema.Field) string                { return "TEXT" }
func (stubDialector) DefaultValueOf(*gormschema.Field) clause.Expression { return clause.Expr{} }
func (stubDialector) BindVarTo(writer clause.Writer, _ *gormio.Statement, _ any) {
	_ = writer.WriteByte('?')
}
func (stubDialector) QuoteTo(writer clause.Writer, str string) {
	_, _ = writer.WriteString(`"` + str + `"`)
}
func (stubDialector) Explain(sql string, _ ...any) string { return sql }

func newStubGormDB(t *testing.T) *gormio.DB {
	t.Helper()
	db, err := gormio.Open(stubDialector{}, &gormio.Config{})
	if err != nil {
		t.Fatalf("open stub gorm: %v", err)
	}
	return db
}

// --- Test fixtures ---------------------------------------------------------

type relUser struct {
	ID      uint
	Name    string
	Books   []*relBook  `gorm:"-"`
	Profile *relProfile `gorm:"-"`
	Roles   []*relRole  `gorm:"-"`
	Houses  []*relHouse `gorm:"-"`
	Logo    *relLogo    `gorm:"-"`
}

func (relUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Books":   contractsorm.HasMany{Related: &relBook{}, ForeignKey: "user_id"},
		"Profile": contractsorm.HasOne{Related: &relProfile{}, ForeignKey: "user_id"},
		"Roles":   contractsorm.Many2Many{Related: &relRole{}, Table: "rel_user_roles"},
		"Houses":  contractsorm.MorphMany{Related: &relHouse{}, Name: "houseable"},
		"Logo":    contractsorm.MorphOne{Related: &relLogo{}, Name: "logoable"},
	}
}

type relBook struct {
	ID       uint
	Title    string
	UserID   uint
	AuthorID uint
	Author   *relUser `gorm:"-"`
}

func (relBook) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Author": contractsorm.BelongsTo{Related: &relUser{}, ForeignKey: "author_id"},
	}
}

type relProfile struct {
	ID     uint
	Bio    string
	UserID uint
}

type relRole struct {
	ID   uint
	Name string
}

type relHouse struct {
	ID            uint
	Address       string
	HouseableID   uint
	HouseableType string
}

type relLogo struct {
	ID           uint
	URL          string
	LogoableID   uint
	LogoableType string
}

// relCountry / relPost via relUser participate in a HasManyThrough setup.
type relCountry struct {
	ID   uint
	Name string
}

func (relCountry) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Posts": contractsorm.HasManyThrough{
			Related: &relPost{},
			Through: &relUser{},
		},
		"FirstPost": contractsorm.HasOneThrough{
			Related: &relPost{},
			Through: &relUser{},
		},
		"NoRelated": contractsorm.HasManyThrough{},
		"BadKind":   unknownRelation{},
	}
}

// unknownRelation satisfies contractsorm.Relation but isn't one of the known per-kind structs.
// Used to exercise the resolver's default branch (OrmMorphRelationKindUnknown) — a defensive
// path that triggers if someone hand-rolls a Relation impl outside the standard set.
type unknownRelation struct{}

func (unknownRelation) Kind() contractsorm.RelationKind { return "weird" }

type relPost struct {
	ID     uint
	Title  string
	UserID uint
}

// --- Pure helpers ----------------------------------------------------------

// --- Schema-dependent helpers ---------------------------------------------

func TestTableNameFor(t *testing.T) {
	db := newStubGormDB(t)
	name, err := tableNameFor(db, &relUser{})
	assert.NoError(t, err)
	assert.Equal(t, "rel_users", name)

	// Invalid (non-struct) model surfaces parse error.
	_, err = tableNameFor(db, "not-a-model")
	assert.Error(t, err)
}

// --- resolveRelation across all kinds -------------------------------------

func TestResolveRelation_Empty(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &relUser{}, "")
	assert.True(t, errors.Is(err, errors.OrmQueryEmptyRelation))
}

func TestResolveRelation_NotFound(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &relUser{}, "Missing")
	assert.True(t, errors.Is(err, errors.OrmRelationNotFound))
}

func TestResolveRelation_HasMany(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Books")
	assert.NoError(t, err)
	assert.Equal(t, relKindHasMany, desc.kind)
	assert.Equal(t, "rel_users", desc.parentTable)
	assert.Equal(t, "rel_books", desc.relatedTable)
	assert.NotEmpty(t, desc.references)
}

func TestResolveRelation_HasOne(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Profile")
	assert.NoError(t, err)
	assert.Equal(t, relKindHasOne, desc.kind)
	assert.Equal(t, "rel_profiles", desc.relatedTable)
}

func TestResolveRelation_BelongsTo(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relBook{}, "Author")
	assert.NoError(t, err)
	assert.Equal(t, relKindBelongsTo, desc.kind)
	assert.Equal(t, "rel_users", desc.relatedTable)
}

func TestResolveRelation_Many2Many(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Roles")
	assert.NoError(t, err)
	assert.Equal(t, relKindMany2Many, desc.kind)
	assert.Equal(t, "rel_user_roles", desc.pivotTable)
	assert.Equal(t, "rel_users", desc.pivotParentRef.primaryTable)
	assert.Equal(t, "rel_roles", desc.pivotRelatedRef.primaryTable)
}

func TestResolveRelation_MorphMany(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Houses")
	assert.NoError(t, err)
	assert.Equal(t, relKindMorphMany, desc.kind)
	assert.Equal(t, "houseable_type", desc.morphTypeColumn)
	assert.Equal(t, "houseable_id", desc.morphIDColumn)
	assert.NotEmpty(t, desc.references)
}

func TestResolveRelation_MorphOne(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Logo")
	assert.NoError(t, err)
	assert.Equal(t, relKindMorphOne, desc.kind)
	assert.Equal(t, "logoable_type", desc.morphTypeColumn)
}

func TestResolveRelation_Nested(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Books.Author")
	assert.NoError(t, err)
	assert.Equal(t, "Books", desc.name)
	assert.NotNil(t, desc.nested)
	assert.Equal(t, "Author", desc.nested.name)
	assert.Equal(t, relKindBelongsTo, desc.nested.kind)
}

func TestResolveRelation_HasManyThrough(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relCountry{}, "Posts")
	assert.NoError(t, err)
	assert.Equal(t, relKindHasManyThrough, desc.kind)
	assert.Equal(t, "rel_posts", desc.relatedTable)
	assert.Equal(t, "rel_users", desc.throughTable)
	// Through default keys come from naming conventions:
	//   firstKey  = singular(parentTable) + "_id"
	//   secondKey = singular(throughTable) + "_id"
	//   localKey / secondLocalKey default to "id".
	assert.Equal(t, "rel_country_id", desc.firstKey)
	assert.Equal(t, "rel_user_id", desc.secondKey)
	assert.Equal(t, "id", desc.localKey)
	assert.Equal(t, "id", desc.secondLocalKey)
}

func TestResolveRelation_HasOneThrough(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relCountry{}, "FirstPost")
	assert.NoError(t, err)
	assert.Equal(t, relKindHasOneThrough, desc.kind)
}

func TestResolveRelation_ThroughNotConfigured(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &relCountry{}, "NoRelated")
	assert.True(t, errors.Is(err, errors.OrmRelationThroughNotConfigured))
}

func TestResolveRelation_ThroughBadKind(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &relCountry{}, "BadKind")
	assert.True(t, errors.Is(err, errors.OrmMorphRelationKindUnknown))
}

func TestResolveRelation_ThroughNotImplemented(t *testing.T) {
	db := newStubGormDB(t)
	// relUser does NOT implement ModelWithThroughRelations.
	_, err := resolveRelation(db, &relUser{}, "Anything")
	assert.True(t, errors.Is(err, errors.OrmRelationNotFound))
}

// Sanity: relatedModel is a fresh pointer to the related struct type.
func TestResolveRelation_RelatedModelType(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &relUser{}, "Books")
	assert.NoError(t, err)
	rt := reflect.TypeOf(desc.relatedModel)
	assert.Equal(t, reflect.Pointer, rt.Kind())
	assert.Equal(t, "relBook", rt.Elem().Name())
}

// --- Morph relation fixtures ---

type morphImage struct {
	ID            uint
	URL           string
	ImageableID   uint
	ImageableType string
	Imageable     any `gorm:"-"`
}

func (morphImage) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Imageable": contractsorm.MorphTo{Name: "imageable"},
	}
}

type morphPost struct {
	ID    uint
	Title string
	Tags  []*morphTag `gorm:"-"`
}

func (morphPost) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Tags": contractsorm.MorphToMany{Related: &morphTag{}, Name: "taggable"},
	}
}

type morphTag struct {
	ID    uint
	Name  string
	Posts []*morphPost `gorm:"-"`
}

func (morphTag) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Posts": contractsorm.MorphedByMany{Related: &morphPost{}, Name: "taggable"},
	}
}

type morphBadKind struct{}

func (morphBadKind) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"X": unknownRelation{},
	}
}

type morphMissingRelated struct{}

func (morphMissingRelated) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"X": contractsorm.MorphMany{Name: "imageable"},
	}
}

// --- Morph relation resolution tests ---

func TestResolveRelation_MorphTo(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &morphImage{}, "Imageable")
	assert.NoError(t, err)
	assert.Equal(t, relKindMorphTo, desc.kind)
	assert.Equal(t, "imageable_type", desc.morphTypeColumn)
	assert.Equal(t, "imageable_id", desc.morphIDColumn)
	assert.Equal(t, "id", desc.morphOwnerKey)
	// MorphTo has no single related model; it's resolved per-row.
	assert.Nil(t, desc.relatedModel)
}

func TestResolveRelation_MorphToMany(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &morphPost{}, "Tags")
	assert.NoError(t, err)
	assert.Equal(t, relKindMorphToMany, desc.kind)
	assert.Equal(t, "taggables", desc.pivotTable)
	assert.Equal(t, "taggable_type", desc.morphTypeColumn)
	assert.Equal(t, "taggable_id", desc.morphIDColumn)
	assert.False(t, desc.morphInverse)
	// morphValue defaults to the parent's table name when no MorphClass / morph map override.
	assert.Equal(t, "morph_posts", desc.morphValue)
}

func TestResolveRelation_MorphedByMany(t *testing.T) {
	db := newStubGormDB(t)
	desc, err := resolveRelation(db, &morphTag{}, "Posts")
	assert.NoError(t, err)
	assert.Equal(t, relKindMorphToMany, desc.kind)
	assert.True(t, desc.morphInverse)
	// For inverse, the morph value pins on the related's morph value.
	assert.Equal(t, "morph_posts", desc.morphValue)
}

func TestResolveRelation_MorphBadKind(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &morphBadKind{}, "X")
	assert.True(t, errors.Is(err, errors.OrmMorphRelationKindUnknown))
}

func TestResolveRelation_MorphMissingRelated(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &morphMissingRelated{}, "X")
	assert.True(t, errors.Is(err, errors.OrmMorphRelationMissingField))
}

// --- Forbidden GORM relation tags ---

type forbiddenPolymorphicParent struct {
	ID     uint
	Houses []*forbiddenPolymorphicChild `gorm:"polymorphic:Houseable"`
}

type forbiddenPolymorphicChild struct {
	ID            uint
	HouseableID   uint
	HouseableType string
}

type forbiddenForeignKeyParent struct {
	ID    uint
	Books []*forbiddenForeignKeyChild `gorm:"foreignKey:ParentID"`
}

type forbiddenForeignKeyChild struct {
	ID       uint
	ParentID uint
}

type forbiddenMany2ManyParent struct {
	ID    uint
	Roles []*forbiddenMany2ManyChild `gorm:"many2many:parent_roles"`
}

type forbiddenMany2ManyChild struct {
	ID uint
}

func TestResolveRelation_ForbidsPolymorphicTag(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &forbiddenPolymorphicParent{}, "Houses")
	assert.True(t, errors.Is(err, errors.OrmRelationTagForbidden))
}

func TestResolveRelation_ForbidsForeignKeyTag(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &forbiddenForeignKeyParent{}, "Books")
	assert.True(t, errors.Is(err, errors.OrmRelationTagForbidden))
}

func TestResolveRelation_ForbidsMany2ManyTag(t *testing.T) {
	db := newStubGormDB(t)
	_, err := resolveRelation(db, &forbiddenMany2ManyParent{}, "Roles")
	assert.True(t, errors.Is(err, errors.OrmRelationTagForbidden))
}
