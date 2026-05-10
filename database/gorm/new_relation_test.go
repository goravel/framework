package gorm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm/morphmap"
	"github.com/goravel/framework/errors"
)

// dryRunFind wraps the inner *gormio.DB pulled out of a Goravel Query, runs Find in DryRun mode,
// and returns the resulting SQL string. Used to verify the WHERE / JOIN shape produced by
// Related per kind.
func newRelationSQL(t *testing.T, q contractsorm.Query, dest any) string {
	t.Helper()
	gq, ok := q.(*Query)
	if !ok {
		t.Fatalf("expected *gorm.Query, got %T", q)
	}
	stmt := gq.buildConditions().instance.Session(&gormio.Session{DryRun: true}).Find(dest)
	return stmt.Statement.SQL.String()
}

func TestRelated_HasMany_Where(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 7}, "Books")
	sql := newRelationSQL(t, rel, &[]relBook{})
	assert.Contains(t, sql, "rel_books")
	assert.Contains(t, sql, "user_id")
}

func TestRelated_BelongsTo_Where(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	rel := q.Related(&relBook{AuthorID: 5}, "Author")
	sql := newRelationSQL(t, rel, &[]relUser{})
	assert.Contains(t, sql, "rel_users")
	// BelongsTo: WHERE related.<pk> = parent.<fk>
	assert.Contains(t, strings.ToLower(sql), "id")
}

func TestRelated_MorphMany_AddsTypeFilter(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 9}, "Houses")
	sql := newRelationSQL(t, rel, &[]relHouse{})
	assert.Contains(t, sql, "houseable_id")
	assert.Contains(t, sql, "houseable_type")
}

func TestRelated_MorphOne_AddsTypeFilter(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 9}, "Logo")
	sql := newRelationSQL(t, rel, &relLogo{})
	assert.Contains(t, sql, "logoable_id")
	assert.Contains(t, sql, "logoable_type")
}

func TestRelated_HasManyThrough(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	rel := q.Related(&relCountry{ID: 1}, "Posts")
	sql := newRelationSQL(t, rel, &[]relPost{})
	assert.Contains(t, sql, "rel_posts")
	assert.Contains(t, sql, "INNER JOIN")
	assert.Contains(t, sql, "rel_users")
}

func TestRelated_NotPointer(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(relUser{ID: 1}, "Books") // value, not pointer
	gq, ok := rel.(*Query)
	assert.True(t, ok)
	assert.True(t, errors.Is(gq.instance.Error, errors.OrmRelationParentNotPointer))
}

func TestRelated_NilParent(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(nil, "Books")
	gq, ok := rel.(*Query)
	assert.True(t, ok)
	assert.True(t, errors.Is(gq.instance.Error, errors.OrmRelationParentNotPointer))
}

func TestRelated_RelationNotFound(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 1}, "DoesNotExist")
	gq, ok := rel.(*Query)
	assert.True(t, ok)
	assert.True(t, errors.Is(gq.instance.Error, errors.OrmRelationNotFound))
}

// --- MorphTo --------------------------------------------------------------

// morphParentLikePost is a sample model registered in the morph map for MorphTo tests.
type morphParentLikePost struct {
	ID    uint
	Title string
}

func (morphParentLikePost) TableName() string { return "morph_parent_like_posts" }

func TestRelated_MorphTo_ResolvedViaMorphMap(t *testing.T) {
	morphmap.Reset()
	defer morphmap.Reset()
	morphmap.Register(map[string]any{"post": &morphParentLikePost{}})

	q := newRelQueryWith(t, &morphImage{})
	rel := q.Related(&morphImage{ID: 1, ImageableID: 42, ImageableType: "post"}, "Imageable")
	sql := newRelationSQL(t, rel, &morphParentLikePost{})
	assert.Contains(t, sql, "morph_parent_like_posts")
	assert.Contains(t, sql, "id")
}

func TestRelated_MorphTo_UnregisteredType(t *testing.T) {
	morphmap.Reset()
	defer morphmap.Reset()

	q := newRelQueryWith(t, &morphImage{})
	rel := q.Related(&morphImage{ID: 1, ImageableID: 42, ImageableType: "unknown"}, "Imageable")
	gq, ok := rel.(*Query)
	assert.True(t, ok)
	assert.True(t, errors.Is(gq.instance.Error, errors.OrmMorphTypeUnknown))
}

func TestRelated_MorphTo_EmptyType_YieldsZeroRowQuery(t *testing.T) {
	morphmap.Reset()
	defer morphmap.Reset()

	q := newRelQueryWith(t, &morphImage{})
	rel := q.Related(&morphImage{ID: 1, ImageableID: 0, ImageableType: ""}, "Imageable")
	gq, ok := rel.(*Query)
	assert.True(t, ok)
	// Apply conditions and run Find against an arbitrary table to verify the WHERE renders.
	stmt := gq.buildConditions().instance.Table("any_table").Session(&gormio.Session{DryRun: true}).Find(&[]map[string]any{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "1 = 0")
}

// --- MorphToMany ---------------------------------------------------------

func TestRelated_MorphToMany_AddsPivotTypeFilter(t *testing.T) {
	q := newRelQueryWith(t, &morphPost{})
	rel := q.Related(&morphPost{ID: 3}, "Tags")
	sql := newRelationSQL(t, rel, &[]morphTag{})
	assert.Contains(t, sql, "INNER JOIN")
	assert.Contains(t, sql, "taggables")
	assert.Contains(t, sql, "taggable_type")
}

// --- OnQuery hook ----------------------------------------------------------

// scopedUser declares a HasMany whose OnQuery scope filters out unpublished books. Every code
// path that builds a query for this relation (Related, eager load, existence) must apply
// the scope.
type scopedUser struct {
	ID    uint
	Books []*scopedBook `gorm:"-"`
}

func (scopedUser) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Books": contractsorm.HasMany{
			Related:    &scopedBook{},
			ForeignKey: "user_id",
			OnQuery: func(q contractsorm.Query) contractsorm.Query {
				return q.Where("published", true)
			},
		},
	}
}

type scopedBook struct {
	ID        uint
	UserID    uint
	Title     string
	Published bool
}

func TestRelated_OnQuery_AppliedToReturnedQuery(t *testing.T) {
	q := newRelQueryWith(t, &scopedUser{})
	rel := q.Related(&scopedUser{ID: 5}, "Books")
	sql := newRelationSQL(t, rel, &[]scopedBook{})
	// The generated WHERE must include both the FK constraint and the OnQuery's published=true.
	assert.Contains(t, sql, "user_id")
	assert.Contains(t, sql, "published")
}

func TestCompileExistenceSubquery_OnQuery_AppliedInExistenceCheck(t *testing.T) {
	q := newRelQueryWith(t, &scopedUser{})
	desc, err := resolveRelation(q.instance, &scopedUser{}, "Books")
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&[]scopedBook{})
	sql := stmt.Statement.SQL.String()
	assert.Contains(t, sql, "published")
}
