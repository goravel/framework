package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
)

// WithTestSuite covers the With / Without / WithOnly methods, which run Goravel's own eager
// loader (not GORM Preload). Each kind of relationship is exercised against every available
// driver.
type WithTestSuite struct {
	suite.Suite
	queries map[string]*TestQuery
}

func TestWithSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &WithTestSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *WithTestSuite) SetupSuite() {
	s.queries = NewTestQueryBuilder().All("", false)
}

func (s *WithTestSuite) SetupTest() {
	for _, query := range s.queries {
		query.CreateTable()
	}
}

func (s *WithTestSuite) sqlite() *TestQuery {
	if q, ok := s.queries["SQLite"]; ok {
		return q
	}
	s.T().Skip("requires sqlite driver")
	return nil
}

// ---------------------------------------------------------------------------
// Per-kind integration tests
// ---------------------------------------------------------------------------

func (s *WithTestSuite) TestWith_HasMany() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			alice := &User{Name: "wr_hm_alice", Books: []*Book{{Name: "wr_hm_a1"}, {Name: "wr_hm_a2"}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&alice))
			bob := &User{Name: "wr_hm_bob", Books: []*Book{{Name: "wr_hm_b1"}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&bob))

			var users []User
			s.Nil(query.Query().Where("name like ?", "wr_hm_%").OrderBy("name").
				With("Books").Get(&users))
			s.Len(users, 2)
			s.Equal("wr_hm_alice", users[0].Name)
			s.Len(users[0].Books, 2)
			s.Equal("wr_hm_bob", users[1].Name)
			s.Len(users[1].Books, 1)
		})
	}
}

func (s *WithTestSuite) TestWith_HasOne() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_ho_user", Address: &Address{Name: "wr_ho_addr", Province: "X"}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_ho_user").
				With("Address").First(&loaded))
			s.NotNil(loaded.Address)
			s.Equal("wr_ho_addr", loaded.Address.Name)
		})
	}
}

func (s *WithTestSuite) TestWith_BelongsTo() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_bt_user", Address: &Address{Name: "wr_bt_addr"}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var addrs []Address
			s.Nil(query.Query().Where("name = ?", "wr_bt_addr").
				With("User").Get(&addrs))
			s.Len(addrs, 1)
			s.NotNil(addrs[0].User)
			s.Equal("wr_bt_user", addrs[0].User.Name)
		})
	}
}

func (s *WithTestSuite) TestWith_Many2Many() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			role := &Role{Name: "wr_m2m_role"}
			s.Nil(query.Query().Create(&role))
			u := &User{Name: "wr_m2m_user", Roles: []*Role{role}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_m2m_user").
				With("Roles").First(&loaded))
			s.Len(loaded.Roles, 1)
			s.Equal("wr_m2m_role", loaded.Roles[0].Name)
		})
	}
}

func (s *WithTestSuite) TestWith_MorphOne() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_mo_user", House: &House{Name: "wr_mo_house"}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_mo_user").
				With("House").First(&loaded))
			s.NotNil(loaded.House)
			s.Equal("wr_mo_house", loaded.House.Name)
		})
	}
}

func (s *WithTestSuite) TestWith_MorphMany() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_mm_user", Phones: []*Phone{{Name: "wr_mm_p1"}, {Name: "wr_mm_p2"}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_mm_user").
				With("Phones").First(&loaded))
			s.Len(loaded.Phones, 2)
		})
	}
}

func (s *WithTestSuite) TestWith_HasManyThrough() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// User -> Books -> Authors (declared via userWithThrough.ThroughRelations()).
			u1 := &User{Name: "wr_th_u1", Books: []*Book{{Name: "wr_th_b1", Author: &Author{Name: "wr_th_a1"}}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u1))
			u2 := &User{Name: "wr_th_u2", Books: []*Book{{Name: "wr_th_b2", Author: &Author{Name: "wr_th_a2"}}, {Name: "wr_th_b3", Author: &Author{Name: "wr_th_a3"}}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u2))

			// Reuse the userAuthorsThrough model (defined below) which carries the through
			// declaration but reads from the same users table.
			var users []userAuthorsThrough
			s.Nil(query.Query().Model(&userAuthorsThrough{}).
				Where("name like ?", "wr_th_%").OrderBy("name").
				With("Authors").Get(&users))
			s.Len(users, 2)
			s.Len(users[0].Authors, 1)
			s.Equal("wr_th_a1", users[0].Authors[0].Name)
			s.Len(users[1].Authors, 2)
		})
	}
}

func (s *WithTestSuite) TestWith_Nested() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_n_user", Books: []*Book{{Name: "wr_n_b1", Author: &Author{Name: "wr_n_a1"}}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_n_user").
				With("Books.Author").First(&loaded))
			s.Len(loaded.Books, 1)
			s.NotNil(loaded.Books[0].Author)
			s.Equal("wr_n_a1", loaded.Books[0].Author.Name)
		})
	}
}

func (s *WithTestSuite) TestWith_Callback() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_cb_user", Books: []*Book{{Name: "wr_cb_keep"}, {Name: "wr_cb_drop"}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			cb := func(q contractsorm.Query) contractsorm.Query {
				return q.Where("name = ?", "wr_cb_keep")
			}
			s.Nil(query.Query().Where("name = ?", "wr_cb_user").
				With("Books", cb).First(&loaded))
			s.Len(loaded.Books, 1)
			s.Equal("wr_cb_keep", loaded.Books[0].Name)
		})
	}
}

func (s *WithTestSuite) TestWith_Map() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			role := &Role{Name: "wr_map_role"}
			s.Nil(query.Query().Create(&role))
			u := &User{Name: "wr_map_user", Books: []*Book{{Name: "wr_map_book"}}, Roles: []*Role{role}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_map_user").
				With(map[string]contractsorm.RelationCallback{
					"Books": nil,
					"Roles": nil,
				}).First(&loaded))
			s.Len(loaded.Books, 1)
			s.Len(loaded.Roles, 1)
		})
	}
}

func (s *WithTestSuite) TestWith_Columns() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wr_col_user", Books: []*Book{{Name: "wr_col_b1"}}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wr_col_user").
				With("Books:id,name").First(&loaded))
			s.Len(loaded.Books, 1)
			s.Equal("wr_col_b1", loaded.Books[0].Name)
			// The loader auto-adds the FK column (user_id) so dictionary grouping by FK works,
			// even when the user prunes it from the column list. Other unselected columns stay
			// at zero value — verified here with CreatedAt which we did not request.
			s.Nil(loaded.Books[0].CreatedAt)
		})
	}
}

func (s *WithTestSuite) TestWithout() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wrr_user", Books: []*Book{{Name: "wrr_b"}}, Address: &Address{Name: "wrr_a"}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wrr_user").
				With("Books", "Address").
				Without("Books").
				First(&loaded))
			s.Len(loaded.Books, 0, "Books should not be loaded after Without")
			s.NotNil(loaded.Address)
		})
	}
}

func (s *WithTestSuite) TestWithOnly() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			u := &User{Name: "wro_user", Books: []*Book{{Name: "wro_b"}}, Address: &Address{Name: "wro_a"}}
			s.Nil(query.Query().Select(orm.Associations).Create(&u))

			var loaded User
			s.Nil(query.Query().Where("name = ?", "wro_user").
				With("Books", "Address").
				WithOnly("Books").
				First(&loaded))
			s.Len(loaded.Books, 1)
			s.Nil(loaded.Address, "Address should not be loaded after WithOnly")
		})
	}
}

// TestWith_ChunkedIN verifies that the loader splits IN clauses into batches when the
// parent count exceeds the chunk size, working around hard limits like Oracle 1000 / SQLite 999.
// We use sqlite (which has the strictest default of 999) and seed > 999 parents to confirm.
func (s *WithTestSuite) TestWith_ChunkedIN() {
	q := s.sqlite()
	if q == nil {
		return
	}
	const total = 1100 // > SQLite's default SQLITE_MAX_VARIABLE_NUMBER of 999

	for i := 0; i < total; i++ {
		u := &User{Name: fmt.Sprintf("wr_chunk_%04d", i), Books: []*Book{{Name: fmt.Sprintf("wr_chunk_b_%04d", i)}}}
		s.Nil(q.Query().Select(orm.Associations).Create(&u))
	}

	var users []User
	s.Nil(q.Query().Where("name like ?", "wr_chunk_%").
		With("Books").Get(&users))
	s.Len(users, total, "all parents should be returned")

	loaded := 0
	for _, u := range users {
		loaded += len(u.Books)
	}
	s.Equal(total, loaded, "every parent should have its book loaded across chunked IN queries")
}

// ---------------------------------------------------------------------------
// SQL-shape assertions (sqlite + ToRawSql)
// ---------------------------------------------------------------------------

func (s *WithTestSuite) TestSQL_With_HasMany_DoesNotJoinPreload() {
	q := s.sqlite()
	if q == nil {
		return
	}
	// With defers loading until after the main query runs, so the *outer* SQL must look
	// identical to a plain Get — no joins, no GORM preload markers.
	var users []User
	sql := q.Query().Model(&User{}).With("Books").ToRawSql().Get(&users)
	s.Equal(
		"SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL",
		sql,
	)
}

// ---------------------------------------------------------------------------
// Test-only model: same shape as User but declares HasManyThrough Authors via Books.
// ---------------------------------------------------------------------------

type userAuthorsThrough struct {
	Model
	SoftDeletes
	Name    string
	Authors []*Author `gorm:"-"`
}

func (userAuthorsThrough) TableName() string { return "users" }

func (userAuthorsThrough) Relations() map[string]contractsorm.Relation {
	return map[string]contractsorm.Relation{
		"Authors": contractsorm.HasManyThrough{
			Related:        &Author{},
			Through:        &Book{},
			FirstKey:       "user_id",
			SecondKey:      "book_id",
			LocalKey:       "id",
			SecondLocalKey: "id",
		},
	}
}
