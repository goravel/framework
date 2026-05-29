package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// QueriesRelationshipsTestSuite covers the QueryWithRelations contract: Has, WhereHas, DoesntHave,
// WithCount, HasMorph and the through-relation variants. It runs against every available driver
// using the same harness as QueryTestSuite.
type QueriesRelationshipsTestSuite struct {
	suite.Suite
	queries map[string]*TestQuery
}

func TestQueriesRelationshipsTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &QueriesRelationshipsTestSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *QueriesRelationshipsTestSuite) SetupSuite() {
	// Run against every available driver. Tests that require docker (mysql / postgres / sqlserver)
	// will spin up containers; sqlite is in-process.
	s.queries = NewTestQueryBuilder().All("", false)
}

func (s *QueriesRelationshipsTestSuite) SetupTest() {
	for _, query := range s.queries {
		query.CreateTable()
	}
}

// rq is a tiny ergonomic wrapper that returns the (already relation-capable) Query for chaining.
// It exists so test bodies remain stable if Query stops embedding QueriesRelationships in the
// future; today it is a passthrough.
func rq(q contractsorm.Query) contractsorm.Query { return q }

func (s *QueriesRelationshipsTestSuite) TestHas_Existence() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			alice := &User{Name: "rel_has_alice"}
			s.Nil(query.Query().Create(alice))
			s.Nil(query.Query().Relation(alice, "Books").SaveMany([]*Book{{Name: "ab1"}, {Name: "ab2"}}))
			bob := &User{Name: "rel_has_bob"}
			s.Nil(query.Query().Create(bob))
			carol := &User{Name: "rel_has_carol"}
			s.Nil(query.Query().Create(carol))
			s.Nil(query.Query().Relation(carol, "Books").Save(&Book{Name: "cb1"}))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.Has("Books").
				Where("name like ?", "rel_has_%").Get(&users))
			names := namesOf(users)
			s.ElementsMatch([]string{"rel_has_alice", "rel_has_carol"}, names)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestHas_CountComparison() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			alice := &User{Name: "rel_hasc_alice"}
			s.Nil(query.Query().Create(alice))
			s.Nil(query.Query().Relation(alice, "Books").SaveMany([]*Book{{Name: "h1"}, {Name: "h2"}, {Name: "h3"}}))
			bob := &User{Name: "rel_hasc_bob"}
			s.Nil(query.Query().Create(bob))
			s.Nil(query.Query().Relation(bob, "Books").Save(&Book{Name: "h4"}))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.Has("Books", ">=", 2).
				Where("name like ?", "rel_hasc_%").Get(&users))
			s.Len(users, 1)
			s.Equal("rel_hasc_alice", users[0].Name)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestDoesntHave() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			withBooks := &User{Name: "rel_dh_with"}
			s.Nil(query.Query().Create(withBooks))
			s.Nil(query.Query().Relation(withBooks, "Books").Save(&Book{Name: "dhb"}))
			without := &User{Name: "rel_dh_without"}
			s.Nil(query.Query().Create(without))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.DoesntHave("Books").
				Where("name like ?", "rel_dh_%").Get(&users))
			s.Len(users, 1)
			s.Equal("rel_dh_without", users[0].Name)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestWhereHas_Callback() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			match := &User{Name: "rel_wh_match"}
			s.Nil(query.Query().Create(match))
			s.Nil(query.Query().Relation(match, "Books").Save(&Book{Name: "wh_target"}))
			other := &User{Name: "rel_wh_other"}
			s.Nil(query.Query().Create(other))
			s.Nil(query.Query().Relation(other, "Books").Save(&Book{Name: "wh_other"}))

			rq := rq(query.Query())
			cb := func(q contractsorm.Query) contractsorm.Query {
				return q.Where("name = ?", "wh_target")
			}
			var users []User
			s.Nil(rq.WhereHas("Books", cb).
				Where("name like ?", "rel_wh_%").Get(&users))
			s.Len(users, 1)
			s.Equal("rel_wh_match", users[0].Name)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestHas_BelongsTo() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{Name: "rel_bt_user"}
			s.Nil(query.Query().Create(user))
			addr := &Address{Name: "rel_bt_address"}
			s.Nil(query.Query().Create(addr))
			s.Nil(query.Query().Relation(addr, "User").Associate(user))

			rq := rq(query.Query())
			var addresses []Address
			s.Nil(rq.Has("User").
				Where("name = ?", "rel_bt_address").Get(&addresses))
			s.Len(addresses, 1)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestHas_ManyToMany() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			role := &Role{Name: "rel_mtm_role"}
			s.Nil(query.Query().Create(role))
			withRole := &User{Name: "rel_mtm_with"}
			s.Nil(query.Query().Create(withRole))
			s.Nil(query.Query().Relation(withRole, "Roles").Save(role))
			noRole := &User{Name: "rel_mtm_no"}
			s.Nil(query.Query().Create(noRole))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.Has("Roles").
				Where("name like ?", "rel_mtm_%").Get(&users))
			names := namesOf(users)
			s.Contains(names, "rel_mtm_with")
			s.NotContains(names, "rel_mtm_no")
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestHasMorph() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			withHouse := &User{Name: "rel_hm_with"}
			s.Nil(query.Query().Create(withHouse))
			s.Nil(query.Query().Relation(withHouse, "House").Save(&House{Name: "rel_hm_house"}))
			noHouse := &User{Name: "rel_hm_no"}
			s.Nil(query.Query().Create(noHouse))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.HasMorph("House", []any{&User{}}).
				Where("name like ?", "rel_hm_%").Get(&users))
			s.Len(users, 1)
			s.Equal("rel_hm_with", users[0].Name)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestNestedHas() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// User -> Books -> Author. Only carol has a book with an Author.
			carol := &User{Name: "rel_nested_carol"}
			s.Nil(query.Query().Create(carol))
			carolBook := &Book{Name: "carol_book"}
			s.Nil(query.Query().Relation(carol, "Books").Save(carolBook))
			s.Nil(query.Query().Relation(carolBook, "Author").Save(&Author{Name: "carol_author"}))

			dan := &User{Name: "rel_nested_dan"}
			s.Nil(query.Query().Create(dan))
			s.Nil(query.Query().Relation(dan, "Books").Save(&Book{Name: "dan_book"}))

			rq := rq(query.Query())
			var users []User
			s.Nil(rq.Has("Books.Author").
				Where("name like ?", "rel_nested_%").Get(&users))
			names := namesOf(users)
			s.Contains(names, "rel_nested_carol")
			s.NotContains(names, "rel_nested_dan")
		})
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func namesOf(users []User) []string {
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, u.Name)
	}
	return out
}

// ---------------------------------------------------------------------------
// HasManyThrough integration: User -> Authors through Books.
// Ported from /libs/fedaco/test/relations/database-relation-has-many-through-integration.spec.ts
// ---------------------------------------------------------------------------

// userWithThrough re-uses the users table but declares a HasManyThrough relation via the unified
// Relations() entry point.
type userWithThrough struct{}

func (userWithThrough) TableName() string { return "users" }

func (userWithThrough) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Authors": contractsorm.HasManyThrough{
			Related:        &Author{},
			Through:        &Book{},
			FirstKey:       "user_id", // FK on Book pointing at User
			SecondKey:      "book_id", // FK on Author pointing at Book
			LocalKey:       "id",
			SecondLocalKey: "id",
		},
	}
}

func (s *QueriesRelationshipsTestSuite) TestHasManyThrough_WhereHas() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// Seed: u1 has a book with author "match"; u2 has a book without an author.
			u1 := &User{Name: "rel_th_u1"}
			s.Nil(query.Query().Create(u1))
			b1 := &Book{Name: "th_book1"}
			s.Nil(query.Query().Relation(u1, "Books").Save(b1))
			s.Nil(query.Query().Relation(b1, "Author").Save(&Author{Name: "th_author_match"}))

			u2 := &User{Name: "rel_th_u2"}
			s.Nil(query.Query().Create(u2))
			s.Nil(query.Query().Relation(u2, "Books").Save(&Book{Name: "th_book2"}))

			rq := rq(query.Query().Model(&userWithThrough{}))
			cb := func(q contractsorm.Query) contractsorm.Query {
				return q.Where("authors.name = ?", "th_author_match")
			}
			var users []User
			s.Nil(rq.WhereHas("Authors", cb).
				Where("name like ?", "rel_th_%").Get(&users))
			s.Len(users, 1)
			s.Equal("rel_th_u1", users[0].Name)
		})
	}
}

// ---------------------------------------------------------------------------
// SQL-shape assertions ported from libs/fedaco/test/fedaco-builder-relation.spec.ts.
//
// These tests use sqlite + ToRawSql() to compile a deterministic, fully bound SQL string for
// each query, then compare it byte-for-byte. Identifier quoting (backticks) matches sqlite's
// preferred style; the `users.deleted_at IS NULL` tail is added by Goravel's soft-delete scope.
// ---------------------------------------------------------------------------

func sqliteOnly(s *QueriesRelationshipsTestSuite) *TestQuery {
	if q, ok := s.queries[sqliteDriverName()]; ok {
		return q
	}
	s.T().Skip("requires sqlite driver")
	return nil
}

func (s *QueriesRelationshipsTestSuite) TestSQL_Has() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.Has("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id) AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_DoesntHave() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.DoesntHave("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE NOT EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id) AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_HasCount() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.Has("Books", ">=", 3).ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE (SELECT COUNT(*) FROM `books` WHERE books.user_id = users.id) >= 3 AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_NestedHas() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.Has("Books.Author").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id AND EXISTS (SELECT 1 FROM `authors` WHERE authors.book_id = books.id)) AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_WhereHasWithCallback() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	cb := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name = ?", "wh_target")
	}
	var users []User
	sql := rq.WhereHas("Books", cb).ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id AND name = 'wh_target') AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_OrHas() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.Where("name = ?", "x").OrHas("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE (name = 'x' OR EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id)) AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_OrDoesntHave() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.Where("name = ?", "x").OrDoesntHave("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE (name = 'x' OR NOT EXISTS (SELECT 1 FROM `books` WHERE books.user_id = users.id)) AND `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_WithCount() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.WithCount("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT users.*, (SELECT COUNT(*) FROM `books` WHERE books.user_id = users.id) AS books_count FROM `users` WHERE `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_WithCountAndCallback() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	cb := func(q contractsorm.Query) contractsorm.Query {
		return q.Where("name = ?", "active")
	}
	var users []User
	sql := rq.WithCount(contractsorm.RelationCount{Name: "Books", Callback: cb}).ToRawSql().Get(&users)
	s.Equal(
		"SELECT users.*, (SELECT COUNT(*) FROM `books` WHERE books.user_id = users.id AND name = 'active') AS books_count FROM `users` WHERE `users`.`deleted_at` IS NULL",
		sql,
	)
}

func (s *QueriesRelationshipsTestSuite) TestSQL_WithCountWithAlias() {
	q := sqliteOnly(s)
	if q == nil {
		return
	}
	rq := rq(q.Query().Model(&User{}))
	var users []User
	sql := rq.WithCount(contractsorm.RelationCount{Name: "Books", Alias: "book_total"}).ToRawSql().Get(&users)
	s.Equal(
		"SELECT users.*, (SELECT COUNT(*) FROM `books` WHERE books.user_id = users.id) AS book_total FROM `users` WHERE `users`.`deleted_at` IS NULL",
		sql,
	)
}

func sqliteDriverName() string {
	return "SQLite"
}

// ---------------------------------------------------------------------------
// Aggregate retrieval: WithCount / WithMax / WithMin / WithSum / WithAvg / WithExists
//
// WithAggregate emits a sub-select column with a deterministic alias (see aggregateAlias in
// database/gorm/queries_relationships.go). To read the value back, declare a struct field tagged
// with that alias as its `gorm:"column:..."`. We use a DTO struct here rather than amending User,
// to keep the demonstration self-contained and avoid affecting other suites.
// ---------------------------------------------------------------------------

// userAggregates is a DTO that maps to the same `users` table but exposes the aggregate alias
// columns. It is populated via Model(&User{}).Get(&[]userAggregates{}) — Model controls the FROM
// table and relation resolution; the DTO controls how rows are scanned.
type userAggregates struct {
	ID               uint     `gorm:"column:id"`
	Name             string   `gorm:"column:name"`
	BooksCount       int64    `gorm:"column:books_count"`
	BooksAuthorCount int64    `gorm:"column:books_author_count"`
	PopularBooks     int64    `gorm:"column:popular_books"`
	BooksMaxID       *int64   `gorm:"column:books_max_id"`
	BooksMinID       *int64   `gorm:"column:books_min_id"`
	BooksSumID       *int64   `gorm:"column:books_sum_id"`
	BooksAvgID       *float64 `gorm:"column:books_avg_id"`
	BooksExists      bool     `gorm:"column:books_exists"`
}

func (s *QueriesRelationshipsTestSuite) TestWithCount_Retrieve() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// u1: 2 books, u2: 0 books, u3: 1 book with an author.
			u1 := &User{Name: "agg_count_u1"}
			s.Nil(query.Query().Create(u1))
			s.Nil(query.Query().Relation(u1, "Books").SaveMany([]*Book{{Name: "ab1"}, {Name: "ab2"}}))
			u2 := &User{Name: "agg_count_u2"}
			s.Nil(query.Query().Create(u2))
			u3 := &User{Name: "agg_count_u3"}
			s.Nil(query.Query().Create(u3))
			b3 := &Book{Name: "ab3"}
			s.Nil(query.Query().Relation(u3, "Books").Save(b3))
			s.Nil(query.Query().Relation(b3, "Author").Save(&Author{Name: "Author1"}))

			var rows []userAggregates
			s.Nil(query.Query().Model(&User{}).Where("name like ?", "agg_count_%").OrderBy("name").
				WithCount("Books").
				WithCount("Books.Author").
				Get(&rows))

			s.Len(rows, 3)
			s.Equal(int64(2), rows[0].BooksCount, "u1 has 2 books")
			s.Equal(int64(0), rows[1].BooksCount, "u2 has 0 books")
			s.Equal(int64(1), rows[2].BooksCount, "u3 has 1 book")
			s.Equal(int64(0), rows[0].BooksAuthorCount, "u1's books have no authors")
			s.Equal(int64(1), rows[2].BooksAuthorCount, "u3's book has an author (nested count)")
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestWithCount_CustomAliasAndCallback() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// Three books — only two start with "pop_". Custom alias + callback narrows the count.
			u := &User{Name: "agg_alias_u"}
			s.Nil(query.Query().Create(u))
			s.Nil(query.Query().Relation(u, "Books").SaveMany([]*Book{{Name: "pop_x"}, {Name: "pop_y"}, {Name: "boring"}}))

			cb := func(q contractsorm.Query) contractsorm.Query {
				return q.Where("name like ?", "pop_%")
			}
			var rows []userAggregates
			s.Nil(query.Query().Model(&User{}).Where("name = ?", "agg_alias_u").
				WithCount(contractsorm.RelationCount{Name: "Books", Alias: "popular_books", Callback: cb}).
				Get(&rows))

			s.Len(rows, 1)
			s.Equal(int64(2), rows[0].PopularBooks, "callback filters to 2 books named pop_*")
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestWithMaxMinSumAvg_Retrieve() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// Three books with consecutive auto-increment IDs (1, 2, 3) on a fresh table.
			u := &User{Name: "agg_num_u"}
			s.Nil(query.Query().Create(u))
			s.Nil(query.Query().Relation(u, "Books").SaveMany([]*Book{{Name: "n1"}, {Name: "n2"}, {Name: "n3"}}))

			var rows []userAggregates
			s.Nil(query.Query().Model(&User{}).Where("name = ?", "agg_num_u").
				WithMax("Books", "id").
				WithMin("Books", "id").
				WithSum("Books", "id").
				WithAvg("Books", "id").
				Get(&rows))

			s.Len(rows, 1)
			s.NotNil(rows[0].BooksMaxID)
			s.NotNil(rows[0].BooksMinID)
			s.NotNil(rows[0].BooksSumID)
			s.NotNil(rows[0].BooksAvgID)
			minID := *rows[0].BooksMinID
			maxID := *rows[0].BooksMaxID
			s.Equal(int64(2), maxID-minID, "3 consecutive IDs => max-min = 2")
			s.Equal(minID+(minID+1)+(minID+2), *rows[0].BooksSumID)
			s.InDelta(float64(minID)+1.0, *rows[0].BooksAvgID, 0.001)
		})
	}
}

func (s *QueriesRelationshipsTestSuite) TestWithExists_Retrieve() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			withBooks := &User{Name: "agg_exist_yes"}
			s.Nil(query.Query().Create(withBooks))
			s.Nil(query.Query().Relation(withBooks, "Books").Save(&Book{Name: "ex1"}))
			withoutBooks := &User{Name: "agg_exist_no"}
			s.Nil(query.Query().Create(withoutBooks))

			var rows []userAggregates
			s.Nil(query.Query().Model(&User{}).Where("name like ?", "agg_exist_%").OrderBy("name").
				WithExists("Books").
				Get(&rows))

			s.Len(rows, 2)
			s.False(rows[0].BooksExists, "agg_exist_no has no books")
			s.True(rows[1].BooksExists, "agg_exist_yes has books")
		})
	}
}
