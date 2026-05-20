package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"
)

// This file tests the exact SQL emitted by Goravel's relation system for every relation kind.
// SQL strings are captured from the stub dialector (see relation_sql_capture_test.go) and pinned
// here so any change to the query builder is flagged loudly.

// ---------------------------------------------------------------------------
// Related() SQL — one query method per relation kind
// ---------------------------------------------------------------------------

func TestRelated_HasOne_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 7}, "Profile")
	sql := newRelationSQL(t, rel, &relProfile{})
	assert.Equal(t, `SELECT * FROM "rel_profiles" WHERE "user_id" = ?`, sql)
}

func TestRelated_HasMany_SQL_Exact(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 7}, "Books")
	sql := newRelationSQL(t, rel, &[]relBook{})
	assert.Equal(t, `SELECT * FROM "rel_books" WHERE "user_id" = ?`, sql)
}

func TestRelated_BelongsTo_SQL_Exact(t *testing.T) {
	q := newRelQueryWith(t, &relBook{})
	rel := q.Related(&relBook{AuthorID: 5}, "Author")
	sql := newRelationSQL(t, rel, &relUser{})
	assert.Equal(t, `SELECT * FROM "rel_users" WHERE "id" = ?`, sql)
}

func TestRelated_Many2Many_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 7}, "Roles")
	sql := newRelationSQL(t, rel, &[]relRole{})
	assert.Equal(t, `SELECT "rel_roles"."id","rel_roles"."name" FROM "rel_roles" INNER JOIN rel_user_roles ON rel_user_roles.rel_role_id = rel_roles.id WHERE rel_user_roles.rel_user_id = ?`, sql)
}

func TestRelated_MorphOne_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 9}, "Logo")
	sql := newRelationSQL(t, rel, &relLogo{})
	assert.Equal(t, `SELECT * FROM "rel_logos" WHERE "logoable_id" = ? AND "logoable_type" = ?`, sql)
}

func TestRelated_MorphMany_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relUser{})
	rel := q.Related(&relUser{ID: 9}, "Houses")
	sql := newRelationSQL(t, rel, &[]relHouse{})
	assert.Equal(t, `SELECT * FROM "rel_houses" WHERE "houseable_id" = ? AND "houseable_type" = ?`, sql)
}

func TestRelated_MorphToMany_SQL(t *testing.T) {
	q := newRelQueryWith(t, &morphPost{})
	rel := q.Related(&morphPost{ID: 3}, "Tags")
	sql := newRelationSQL(t, rel, &[]morphTag{})
	assert.Equal(t, `SELECT "morph_tags"."id","morph_tags"."name" FROM "morph_tags" INNER JOIN taggables ON taggables.morph_tag_id = morph_tags.id WHERE taggables.taggable_id = ? AND taggables.taggable_type = ?`, sql)
}

func TestRelated_MorphedByMany_SQL(t *testing.T) {
	q := newRelQueryWith(t, &morphTag{})
	rel := q.Related(&morphTag{ID: 1}, "Posts")
	sql := newRelationSQL(t, rel, &[]morphPost{})
	assert.Equal(t, `SELECT "morph_posts"."id","morph_posts"."title" FROM "morph_posts" INNER JOIN taggables ON taggables.morph_post_id = morph_posts.id WHERE taggables.taggable_id = ? AND taggables.taggable_type = ?`, sql)
}

func TestRelated_HasManyThrough_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	rel := q.Related(&relCountry{ID: 1}, "Posts")
	sql := newRelationSQL(t, rel, &[]relPost{})
	assert.Equal(t, `SELECT "rel_posts"."id","rel_posts"."title","rel_posts"."user_id" FROM "rel_posts" INNER JOIN rel_users ON rel_posts.rel_user_id = rel_users.id WHERE rel_users.rel_country_id = ?`, sql)
}

func TestRelated_HasOneThrough_SQL(t *testing.T) {
	q := newRelQueryWith(t, &relCountry{})
	rel := q.Related(&relCountry{ID: 1}, "FirstPost")
	sql := newRelationSQL(t, rel, &relPost{})
	assert.Equal(t, `SELECT "rel_posts"."id","rel_posts"."title","rel_posts"."user_id" FROM "rel_posts" INNER JOIN rel_users ON rel_posts.rel_user_id = rel_users.id WHERE rel_users.rel_country_id = ?`, sql)
}

// ---------------------------------------------------------------------------
// compileExistenceSubquery SQL — used by Has / WhereHas / DoesntHave
// ---------------------------------------------------------------------------

func runExistenceSQL(t *testing.T, model any, relation string, dest any) string {
	t.Helper()
	q := newRelQueryWith(t, model)
	desc, err := resolveRelation(q.instance, model, relation)
	assert.NoError(t, err)
	inner := q.compileExistenceSubquery(desc, nil)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(dest)
	return stmt.Statement.SQL.String()
}

func TestCompileExistenceSubquery_HasOne_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relUser{}, "Profile", &relProfile{})
	assert.Equal(t, `SELECT 1 FROM "rel_profiles" WHERE rel_profiles.user_id = rel_users.id`, sql)
}

func TestCompileExistenceSubquery_HasMany_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relUser{}, "Books", &[]relBook{})
	assert.Equal(t, `SELECT 1 FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileExistenceSubquery_BelongsTo_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relBook{}, "Author", &[]relUser{})
	assert.Equal(t, `SELECT 1 FROM "rel_users" WHERE rel_users.id = rel_books.author_id`, sql)
}

func TestCompileExistenceSubquery_Many2Many_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relUser{}, "Roles", &[]relRole{})
	assert.Equal(t, `SELECT 1 FROM "rel_roles" INNER JOIN rel_user_roles ON rel_user_roles.rel_role_id = rel_roles.id WHERE rel_user_roles.rel_user_id = rel_users.id`, sql)
}

func TestCompileExistenceSubquery_MorphOne_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relUser{}, "Logo", &relLogo{})
	assert.Equal(t, `SELECT 1 FROM "rel_logos" WHERE rel_logos.logoable_id = rel_users.id AND rel_logos.logoable_type = ?`, sql)
}

func TestCompileExistenceSubquery_MorphMany_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relUser{}, "Houses", &[]relHouse{})
	assert.Equal(t, `SELECT 1 FROM "rel_houses" WHERE rel_houses.houseable_id = rel_users.id AND rel_houses.houseable_type = ?`, sql)
}

func TestCompileExistenceSubquery_MorphToMany_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &morphPost{}, "Tags", &[]morphTag{})
	assert.Equal(t, `SELECT 1 FROM "morph_tags" INNER JOIN taggables ON taggables.morph_tag_id = morph_tags.id WHERE taggables.taggable_id = morph_posts.id AND taggables.taggable_type = ?`, sql)
}

func TestCompileExistenceSubquery_MorphedByMany_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &morphTag{}, "Posts", &[]morphPost{})
	assert.Equal(t, `SELECT 1 FROM "morph_posts" INNER JOIN taggables ON taggables.morph_post_id = morph_posts.id WHERE taggables.taggable_id = morph_tags.id AND taggables.taggable_type = ?`, sql)
}

func TestCompileExistenceSubquery_HasManyThrough_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relCountry{}, "Posts", &[]relPost{})
	assert.Equal(t, `SELECT 1 FROM "rel_posts" INNER JOIN rel_users ON rel_posts.rel_user_id = rel_users.id WHERE rel_users.rel_country_id = rel_countries.id`, sql)
}

func TestCompileExistenceSubquery_HasOneThrough_SQL(t *testing.T) {
	sql := runExistenceSQL(t, &relCountry{}, "FirstPost", &relPost{})
	assert.Equal(t, `SELECT 1 FROM "rel_posts" INNER JOIN rel_users ON rel_posts.rel_user_id = rel_users.id WHERE rel_users.rel_country_id = rel_countries.id`, sql)
}

// ---------------------------------------------------------------------------
// compileAggregateSubquery SQL — used by WithCount / WithMax / WithMin / WithSum / WithAvg / WithExists
// ---------------------------------------------------------------------------

func runAggregateSQL(t *testing.T, sub selectSub) string {
	t.Helper()
	q := newRelQueryWith(t, &relUser{})
	desc, err := resolveRelation(q.instance, &relUser{}, "Books")
	assert.NoError(t, err)
	inner := q.compileAggregateSubquery(desc, sub)
	stmt := inner.Session(&gormio.Session{DryRun: true}).Find(&relBook{})
	return stmt.Statement.SQL.String()
}

func TestCompileAggregateSubquery_Count_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "*", function: "count"})
	assert.Equal(t, `SELECT COUNT(*) FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileAggregateSubquery_Sum_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "id", function: "sum"})
	assert.Equal(t, `SELECT SUM(rel_books.id) FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileAggregateSubquery_Max_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "id", function: "max"})
	assert.Equal(t, `SELECT MAX(rel_books.id) FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileAggregateSubquery_Min_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "id", function: "min"})
	assert.Equal(t, `SELECT MIN(rel_books.id) FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileAggregateSubquery_Avg_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "id", function: "avg"})
	assert.Equal(t, `SELECT AVG(rel_books.id) FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}

func TestCompileAggregateSubquery_Exists_SQL(t *testing.T) {
	sql := runAggregateSQL(t, selectSub{relation: "Books", column: "*", function: "exists"})
	assert.Equal(t, `SELECT 1 FROM "rel_books" WHERE rel_books.user_id = rel_users.id`, sql)
}
