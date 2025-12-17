package tests

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/goravel/framework/contracts/database/orm"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/convert"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type QueryTestSuite struct {
	suite.Suite
	queries         map[string]*TestQuery
	additionalQuery *TestQuery
}

func TestQueryTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &QueryTestSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *QueryTestSuite) SetupSuite() {
	s.queries = NewTestQueryBuilder().All("", false)
	s.additionalQuery = NewTestQueryBuilder().Postgres("", false)
}

func (s *QueryTestSuite) SetupTest() {
	for _, query := range s.queries {
		query.CreateTable()
	}
	s.additionalQuery.CreateTable()
}

func (s *QueryTestSuite) TearDownSuite() {
	if s.queries[sqlite.Name] != nil {
		docker, err := s.queries[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *QueryTestSuite) TestAssociation() {
	for driver, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "Find",
				setup: func() {
					user := User{
						Name: "association_find_name",
						Address: &Address{
							Name: "association_find_address",
						},
						age: 1,
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)

					var userAddress Address
					s.Nil(query.Query().Model(&user1).Association("Address").Find(&userAddress))
					s.True(userAddress.ID > 0)
					s.Equal("association_find_address", userAddress.Name)
				},
			},
			{
				name: "hasOne Append",
				setup: func() {
					user := User{
						Name: "association_has_one_append_name",
						Address: &Address{
							Name: "association_has_one_append_address",
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID), driver)
					s.True(user1.ID > 0, driver)
					s.Nil(query.Query().Model(&user1).Association("Address").Append(&Address{Name: "association_has_one_append_address1"}), driver)

					s.Nil(query.Query().Load(&user1, "Address"), driver)
					s.True(user1.Address.ID > 0, driver)
					s.Equal("association_has_one_append_address1", user1.Address.Name, driver)
				},
			},
			{
				name: "hasMany Append",
				setup: func() {
					user := User{
						Name: "association_has_many_append_name",
						Books: []*Book{
							{Name: "association_has_many_append_address1"},
							{Name: "association_has_many_append_address2"},
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Query().Model(&user1).Association("Books").Append(&Book{Name: "association_has_many_append_address3"}))

					s.Nil(query.Query().Load(&user1, "Books"))
					s.Equal(3, len(user1.Books))
					s.Equal("association_has_many_append_address3", user1.Books[2].Name)
				},
			},
			{
				name: "hasOne Replace",
				setup: func() {
					user := User{
						Name: "association_has_one_append_name",
						Address: &Address{
							Name: "association_has_one_append_address",
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Query().Model(&user1).Association("Address").Replace(&Address{Name: "association_has_one_append_address1"}))

					s.Nil(query.Query().Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.Equal("association_has_one_append_address1", user1.Address.Name)
				},
			},
			{
				name: "hasMany Replace",
				setup: func() {
					user := User{
						Name: "association_has_many_replace_name",
						Books: []*Book{
							{Name: "association_has_many_replace_address1"},
							{Name: "association_has_many_replace_address2"},
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Query().Model(&user1).Association("Books").Replace(&Book{Name: "association_has_many_replace_address3"}))

					s.Nil(query.Query().Load(&user1, "Books"))
					s.Equal(1, len(user1.Books))
					s.Equal("association_has_many_replace_address3", user1.Books[0].Name)
				},
			},
			{
				name: "Delete",
				setup: func() {
					user := User{
						Name: "association_delete_name",
						Address: &Address{
							Name: "association_delete_address",
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					// No ID when Delete
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Query().Model(&user1).Association("Address").Delete(&Address{Name: "association_delete_address"}))

					s.Nil(query.Query().Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.Equal("association_delete_address", user1.Address.Name)

					// Has ID when Delete
					var user2 User
					s.Nil(query.Query().Find(&user2, user.ID))
					s.True(user2.ID > 0)
					var userAddress Address
					userAddress.ID = user1.Address.ID
					s.Nil(query.Query().Model(&user2).Association("Address").Delete(&userAddress))

					s.Nil(query.Query().Load(&user2, "Address"))
					s.Nil(user2.Address)
				},
			},
			{
				name: "Clear",
				setup: func() {
					user := User{
						Name: "association_clear_name",
						Address: &Address{
							Name: "association_clear_address",
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					// No ID when Delete
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Query().Model(&user1).Association("Address").Clear())

					s.Nil(query.Query().Load(&user1, "Address"))
					s.Nil(user1.Address)
				},
			},
			{
				name: "Count",
				setup: func() {
					user := User{
						Name: "association_count_name",
						Books: []*Book{
							{Name: "association_count_address1"},
							{Name: "association_count_address2"},
						},
					}

					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Equal(int64(2), query.Query().Model(&user1).Association("Books").Count())
				},
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestBelongsTo() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "belongs_to_name",
				Address: &Address{
					Name: "belongs_to_address",
				},
			}

			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var userAddress Address
			s.Nil(query.Query().With("User").Where("name = ?", "belongs_to_address").First(&userAddress))
			s.True(userAddress.ID > 0)
			s.True(userAddress.User.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestCount() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			count, err := query.Query().Model(&User{}).Where("name = ?", "count_user").Count()
			s.Nil(err)
			s.True(count > 0)

			count, err = query.Query().Table("users").Where("name = ?", "count_user").Count()
			s.Nil(err)
			s.True(count > 0)
		})
	}
}

func (s *QueryTestSuite) TestCreate() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success by struct",
				setup: func() {
					user := User{Name: "create_user"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
				},
			},
			{
				name: "batch create success by struct",
				setup: func() {
					users := []User{
						{Name: "batch_create_user_by_struct_1"},
						{Name: "batch_create_user_by_struct_2"},
					}
					s.Nil(query.Query().Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)
				},
			},
			{
				name: "success by map",
				setup: func() {
					s.Nil(query.Query().Table("users").Create(map[string]any{
						"name":       "create_by_map_name1",
						"avatar":     "create_by_map_avatar1",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}))

					var user1 User
					err := query.Query().Where("name", "create_by_map_name1").
						Where("avatar", "create_by_map_avatar1").First(&user1)
					s.NoError(err)
					s.True(user1.ID > 0)

					s.Nil(query.Query().Model(User{}).Create(map[string]any{
						"Name":      "create_by_map_name2",
						"Avatar":    "create_by_map_avatar2",
						"CreatedAt": carbon.Now(),
						"UpdatedAt": carbon.Now(),
					}))

					var user2 User
					err = query.Query().Where("name", "create_by_map_name2").
						Where("avatar", "create_by_map_avatar2").First(&user2)
					s.NoError(err)
					s.True(user2.ID > 0)
				},
			},
			{
				name: "batch create success by map",
				setup: func() {
					s.Nil(query.Query().Table("users").Create([]map[string]any{
						{
							"name":       "batch_create_by_map_name1",
							"avatar":     "batch_create_by_map_avatar1",
							"created_at": carbon.Now(),
							"updated_at": carbon.Now(),
						},
						{
							"name":       "batch_create_by_map_name2",
							"avatar":     "batch_create_by_map_avatar2",
							"created_at": carbon.Now(),
							"updated_at": carbon.Now(),
						},
					}))

					var users1 []User
					err := query.Query().Where("name", "batch_create_by_map_name1").OrWhere("name", "batch_create_by_map_name2").Find(&users1)
					s.NoError(err)
					s.Len(users1, 2)

					// The []map should be a pointer, otherwise gorm will throw an error
					s.Nil(query.Query().Model(User{}).Create(&[]map[string]any{
						{
							"Name":      "batch_create_by_map_name3",
							"Avatar":    "batch_create_by_map_avatar3",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
						{
							"Name":      "batch_create_by_map_name4",
							"Avatar":    "batch_create_by_map_avatar4",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
					}))

					var users2 []User
					err = query.Query().Where("name", "batch_create_by_map_name3").OrWhere("name", "batch_create_by_map_name4").Find(&users2)
					s.NoError(err)
					s.Len(users2, 2)
				},
			},
			{
				name: "success when refresh connection",
				setup: func() {
					config := s.additionalQuery.Driver().Pool().Writers[0]
					config.Connection = "dummy"
					mockDatabaseConfig(query.MockConfig(), config)

					people := People{Body: "create_people"}
					s.Nil(query.Query().Create(&people))
					s.True(people.ID > 0)

					count, err := query.Query().Table("peoples").Where("body", "create_people").Count()
					s.NoError(err)
					s.True(count == 0)

					s.Nil(query.Query().Model(&People{}).Create(map[string]any{
						"body":       "create_people1",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}))

					var people1 People
					s.Nil(query.Query().Where("body", "create_people1").First(&people1))
					s.True(people1.ID > 0)

					count, err = query.Query().Table("peoples").Where("body", "create_people1").Count()
					s.NoError(err)
					s.True(count == 0)
				},
			},
			{
				name: "success when create with no relationships",
				setup: func() {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "success when create with select Associations",
				setup: func() {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Query().Select(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "success when create with select fields",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Query().Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "success when create with omit fields",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Query().Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "success create with omit Associations",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Query().Omit(gorm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "error when set select and omit at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Query().Omit(gorm.Associations).Select("Name").Create(&user), errors.OrmQuerySelectAndOmitsConflict.Error())
				},
			},
			{
				name: "error when select that set fields and Associations at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Query().Select("Name", gorm.Associations).Create(&user), errors.OrmQueryAssociationsConflict.Error())
				},
			},
			{
				name: "error when omit that set fields and Associations at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Query().Omit("Name", gorm.Associations).Create(&user), errors.OrmQueryAssociationsConflict.Error())
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestCursor() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "cursor_user", Avatar: "cursor_avatar", Address: &Address{Name: "cursor_address"}, Books: []*Book{
				{Name: "cursor_book"},
			}}
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "cursor_user", Avatar: "cursor_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "cursor_user", Avatar: "cursor_avatar2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)
			res, err := query.Query().Delete(&user2)
			s.Nil(err)
			s.Equal(int64(1), res.RowsAffected)

			// success
			users := query.Query().Model(&User{}).Where("name = ?", "cursor_user").WithTrashed().With("Address").With("Books").Cursor()
			var size int
			var addressNum int
			var bookNum int
			for row := range users {
				s.Nil(row.Err())
				var tempUser User
				s.Nil(row.Scan(&tempUser))
				s.True(tempUser.ID > 0)
				s.True(len(tempUser.Name) > 0)
				s.NotEmpty(tempUser.CreatedAt.String())
				s.NotEmpty(tempUser.UpdatedAt.String())
				s.Equal(tempUser.DeletedAt.Valid, tempUser.ID == user2.ID)
				size++

				if tempUser.Address != nil {
					addressNum++
				}
				bookNum += len(tempUser.Books)
			}
			s.Equal(3, size)
			s.Equal(1, addressNum)
			s.Equal(1, bookNum)

			// error
			for row := range query.Query().Table("not_exist").Cursor() {
				err1 := row.Err()
				s.Error(err1)

				err2 := row.Scan(map[string]any{})
				s.Error(err2)

				s.Equal(err1, err2)
			}
		})
	}
}

func (s *QueryTestSuite) TestDBRaw() {
	userName := "db_raw"
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: userName}

			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)
			switch driver {
			case sqlserver.Name, mysql.Name:
				res, err := query.Query().Model(&user).Update("Name", databasedb.Raw("concat(name, ?)", driver))
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)
			default:
				res, err := query.Query().Model(&user).Update("Name", databasedb.Raw("name || ?", driver))
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)
			}

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.True(user1.ID > 0)
			s.True(user1.Name == userName+driver)
		})
	}
}

func (s *QueryTestSuite) TestDelete() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Delete(&user)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by table",
				setup: func() {
					user := User{Name: "delete_user_by_table", Avatar: "delete_avatar_by_table"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Table("users").Where("name", "delete_user_by_table").Delete()
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by model",
				setup: func() {
					user := User{Name: "delete_user_by_model", Avatar: "delete_avatar_by_model"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Model(&User{}).Where("name", "delete_user_by_model").Delete()
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success when refresh connection",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Delete(&user)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)

					// refresh connection
					config := s.additionalQuery.Driver().Pool().Writers[0]
					config.Connection = "dummy"
					mockDatabaseConfig(query.MockConfig(), config)

					people := People{Body: "delete_people"}
					s.Nil(query.Query().Create(&people))
					s.True(people.ID > 0)

					res, err = query.Query().Delete(&people)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var people1 People
					s.Nil(query.Query().Find(&people1, people.ID))
					s.Equal(uint(0), people1.ID)
				},
			},
			{
				name: "success by id",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Where("id", user.ID).Delete(&User{})
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by multiple",
				setup: func() {
					users := []User{{Name: "delete_user", Avatar: "delete_avatar"}, {Name: "delete_user1", Avatar: "delete_avatar1"}}
					s.Nil(query.Query().Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Query().WhereIn("id", []any{users[0].ID, users[1].ID}).Delete(&User{})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					count, err := query.Query().Model(&User{}).Where("name", "delete_user").OrWhere("name", "delete_user1").Count()
					s.Nil(err)
					s.True(count == 0)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestDistinct() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "distinct_user", Avatar: "distinct_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "distinct_user", Avatar: "distinct_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Distinct("name").Find(&users, []uint{user.ID, user1.ID}))
			s.Equal(1, len(users))

			var users1 []User
			s.Nil(query.Query().Distinct().Select("name").Find(&users1, []uint{user.ID, user1.ID}))
			s.Equal(1, len(users1))
		})
	}
}

func (s *QueryTestSuite) TestEvent_Creating() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when create by struct",
				setup: func() {
					user := User{Name: "event_creating_name"}
					s.Nil(query.Query().Create(&user))
					s.Equal("event_creating_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_creating_name", user1.Name)
					s.Equal("event_creating_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create by map",
				setup: func() {
					s.Nil(query.Query().Model(&User{}).Create(map[string]any{
						"name":       "event_creating_by_map_name",
						"avatar":     "event_creating_by_map_avatar",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}))

					var user User
					s.Nil(query.Query().Where("name", "event_creating_by_map_name").Find(&user))
					s.Equal("event_creating_by_map_avatar1", user.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrCreate(&user, User{Name: "event_creating_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_creating_FirstOrCreate_name", user.Name)
					s.Equal("event_creating_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_creating_FirstOrCreate_name", user1.Name)
					s.Equal("event_creating_FirstOrCreate_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create with omit",
				setup: func() {
					user := User{Name: "event_creating_omit_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_creating_omit_create_address"
					user.Books[0].Name = "event_creating_omit_create_book0"
					user.Books[1].Name = "event_creating_omit_create_book1"
					s.Nil(query.Query().Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_creating_omit_create_avatar", user.Avatar)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "trigger when create with select",
				setup: func() {
					user := User{Name: "event_creating_select_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_creating_select_create_address"
					user.Books[0].Name = "event_creating_select_create_book0"
					user.Books[1].Name = "event_creating_select_create_book1"
					s.Nil(query.Query().Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_creating_select_create_avatar", user.Avatar)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_creating_save_name"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_creating_save_avatar", user.Avatar)
				},
			},
			{
				name: "not trigger when creating by slice struct",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "creating event with slice struct")
					users := []User{{Name: "event_creating_slice_name"}, {Name: "event_creating_slice_name1"}}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(users))
				},
			},
			{
				name: "not trigger when creating by slice map",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "creating event with slice map")
					users := []map[string]any{
						{
							"Name":      "event_creating_slice_name",
							"Avatar":    "event_creating_slice_avatar",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
						{
							"Name":      "event_creating_slice_name1",
							"Avatar":    "event_creating_slice_avatar1",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
					}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(&users))
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Created() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when create by struct",
				setup: func() {
					user := User{Name: "event_created_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))
					s.Equal(fmt.Sprintf("event_created_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_created_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create by map",
				setup: func() {
					userMap := map[string]any{
						"name":       "event_created_by_map_name",
						"avatar":     "event_created_by_map_avatar",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}
					s.Nil(query.Query().Model(&User{}).Create(userMap))

					s.Equal("event_created_by_map_avatar1", userMap["avatar"])

					var user User
					s.Nil(query.Query().Where("name", "event_created_by_map_name").Find(&user))
					s.Equal("event_created_by_map_avatar", user.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrCreate(&user, User{Name: "event_created_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_created_FirstOrCreate_name", user.Name)
					s.Equal(fmt.Sprintf("event_created_FirstOrCreate_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_created_FirstOrCreate_name", user1.Name)
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when create with omit",
				setup: func() {
					user := User{Name: "event_created_omit_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_created_omit_create_address"
					user.Books[0].Name = "event_created_omit_create_book0"
					user.Books[1].Name = "event_created_omit_create_book1"
					s.Nil(query.Query().Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal(fmt.Sprintf("event_created_omit_create_avatar_%d", user.ID), user.Avatar)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_created_omit_create_name", user1.Name)
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when create with select",
				setup: func() {
					user := User{Name: "event_created_select_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_created_select_create_address"
					user.Books[0].Name = "event_created_select_create_book0"
					user.Books[1].Name = "event_created_select_create_book1"
					s.Nil(query.Query().Select("ID", "Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal(fmt.Sprintf("event_created_select_create_avatar_%d", user.ID), user.Avatar)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_created_select_create_name", user1.Name)
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_created_save_name"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal(fmt.Sprintf("event_created_save_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_created_save_name", user1.Name)
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "not trigger when creating by slice struct",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "created event with slice struct")
					users := []User{{Name: "event_created_slice_name"}, {Name: "event_created_slice_name1"}}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(users))
				},
			},
			{
				name: "not trigger when creating by slice map",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "created event with slice map")
					users := []map[string]any{
						{
							"Name":      "event_created_slice_name",
							"Avatar":    "event_created_slice_avatar",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
						{
							"Name":      "event_created_slice_name1",
							"Avatar":    "event_created_slice_avatar1",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
					}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(&users))
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Saving() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when create by struct",
				setup: func() {
					user := User{Name: "event_saving_create_name"}
					s.Nil(query.Query().Create(&user))
					s.Equal("event_saving_create_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saving_create_name", user1.Name)
					s.Equal("event_saving_create_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create by map",
				setup: func() {
					userMap := map[string]any{
						"name":       "event_saving_create_by_map_name",
						"avatar":     "event_saving_create_by_map_avatar",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}
					s.Nil(query.Query().Model(&User{}).Create(userMap))
					s.Equal("event_saving_create_by_map_avatar1", userMap["avatar"])

					var user1 User
					s.Nil(query.Query().Where("name", "event_saving_create_by_map_name").Find(&user1))
					s.Equal("event_saving_create_by_map_avatar1", user1.Avatar)
				},
			},
			{
				name: "trigger when create with omit",
				setup: func() {
					user := User{Name: "event_saving_omit_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_saving_omit_create_address"
					user.Books[0].Name = "event_saving_omit_create_book0"
					user.Books[1].Name = "event_saving_omit_create_book1"
					s.Nil(query.Query().Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_omit_create_avatar", user.Avatar)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "trigger when create with select",
				setup: func() {
					user := User{Name: "event_saving_select_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_saving_select_create_address"
					user.Books[0].Name = "event_saving_select_create_book0"
					user.Books[1].Name = "event_saving_select_create_book1"
					s.Nil(query.Query().Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_select_create_avatar", user.Avatar)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrCreate(&user, User{Name: "event_saving_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_saving_FirstOrCreate_name", user.Name)
					s.Equal("event_saving_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saving_FirstOrCreate_name", user1.Name)
					s.Equal("event_saving_FirstOrCreate_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_saving_save_name"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_save_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saving_save_name", user1.Name)
					s.Equal("event_saving_save_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by single column",
				setup: func() {
					user := User{Name: "event_saving_single_update_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))

					res, err := query.Query().Model(&user).Update("avatar", "event_saving_single_update_avatar")
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_saving_single_update_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saving_single_update_name", user1.Name)
					s.Equal("event_saving_single_update_avatar1", user1.Avatar)
				},
			},
			{
				name: "not trigger when creating by slice struct",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "saving event with slice struct")
					users := []User{{Name: "event_saving_slice_name"}, {Name: "event_saving_slice_name1"}}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(users))
				},
			},
			{
				name: "not trigger when creating by slice map",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "saving event with slice map")
					users := []map[string]any{
						{
							"Name":      "event_creating_slice_name",
							"Avatar":    "event_creating_slice_avatar",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
						{
							"Name":      "event_creating_slice_name1",
							"Avatar":    "event_creating_slice_avatar1",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
					}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(&users))
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Saved() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when create",
				setup: func() {
					user := User{Name: "event_saved_create_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))
					s.Equal("event_saved_create_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saved_create_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create by map",
				setup: func() {
					userMap := map[string]any{
						"name":       "event_saved_create_by_map_name",
						"avatar":     "event_saved_create_by_map_avatar",
						"created_at": carbon.Now(),
						"updated_at": carbon.Now(),
					}
					s.Nil(query.Query().Model(&User{}).Create(userMap))
					s.Equal("event_saved_create_by_map_avatar1", userMap["avatar"])

					var user1 User
					s.Nil(query.Query().Where("name", "event_saved_create_by_map_name").Find(&user1))
					s.Equal("event_saved_create_by_map_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when create with omit",
				setup: func() {
					user := User{Name: "event_saved_omit_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_saved_omit_create_address"
					user.Books[0].Name = "event_saved_omit_create_book0"
					user.Books[1].Name = "event_saved_omit_create_book1"
					s.Nil(query.Query().Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_saved_omit_create_avatar", user.Avatar)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when create with select",
				setup: func() {
					user := User{Name: "event_saved_select_create_name", Address: &Address{}, Books: []*Book{{}, {}}}
					user.Address.Name = "event_saved_select_create_address"
					user.Books[0].Name = "event_saved_select_create_book0"
					user.Books[1].Name = "event_saved_select_create_book1"
					s.Nil(query.Query().Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_saved_select_create_avatar", user.Avatar)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrCreate(&user, User{Name: "event_saved_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_saved_FirstOrCreate_name", user.Name)
					s.Equal("event_saved_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saved_FirstOrCreate_name", user1.Name)
					s.Empty(user1.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_saved_save_name", Avatar: "avatar"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saved_save_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saved_save_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by map",
				setup: func() {
					user := User{Name: "event_saved_map_update_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))

					res, err := query.Query().Model(&user).Update(map[string]any{
						"avatar": "event_saved_map_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_saved_map_update_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_saved_map_update_name", user1.Name)
					s.Equal("event_saved_map_update_avatar", user1.Avatar)
				},
			},
			{
				name: "not trigger when creating by slice struct",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "saved event with slice struct")
					users := []User{{Name: "event_saved_slice_name"}, {Name: "event_saved_slice_name1"}}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(users))
				},
			},
			{
				name: "not trigger when creating by slice map",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "saved event with slice map")
					users := []map[string]any{
						{
							"Name":      "event_creating_slice_name",
							"Avatar":    "event_creating_slice_avatar",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
						{
							"Name":      "event_creating_slice_name1",
							"Avatar":    "event_creating_slice_avatar1",
							"CreatedAt": carbon.Now(),
							"UpdatedAt": carbon.Now(),
						},
					}
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Create(&users))
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Updating() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "not trigger when create",
				setup: func() {
					user := User{Name: "event_updating_create_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "not trigger when create by save",
				setup: func() {
					user := User{Name: "event_updating_save_name", Avatar: "avatar"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_updating_save_name", Avatar: "avatar"}
					s.Nil(query.Query().Save(&user))

					user.Avatar = "event_updating_save_avatar"
					s.Nil(query.Query().Save(&user))
					s.Equal("event_updating_save_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updating_save_name", user1.Name)
					s.Equal("event_updating_save_avatar1", user1.Avatar)
				},
			},
			{
				name: "trigger when update by model",
				setup: func() {
					user := User{Name: "event_updating_model_update_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))

					res, err := query.Query().Model(&user).Update(User{
						Avatar: "event_updating_model_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal(fmt.Sprintf("event_updating_model_update_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updating_model_update_name", user1.Name)
					s.Equal(fmt.Sprintf("event_updating_model_update_avatar_%d", user.ID), user1.Avatar)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Updated() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "not trigger when create",
				setup: func() {
					user := User{Name: "event_updated_create_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "not trigger when create by save",
				setup: func() {
					user := User{Name: "event_updated_save_name", Avatar: "avatar"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_updated_save_name", Avatar: "avatar"}
					s.Nil(query.Query().Save(&user))

					user.Avatar = "event_updated_save_avatar"
					s.Nil(query.Query().Save(&user))
					s.Equal("event_updated_save_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updated_save_name", user1.Name)
					s.Equal("event_updated_save_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by model",
				setup: func() {
					user := User{Name: "event_updated_model_update_name", Avatar: "avatar"}
					s.Nil(query.Query().Create(&user))

					res, err := query.Query().Model(&user).Update(User{
						Avatar: "event_updated_model_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_updated_model_update_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updated_model_update_name", user1.Name)
					s.Equal("event_updated_model_update_avatar", user1.Avatar)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Deleting() {
	for _, query := range s.queries {
		s.Run("trigger", func() {
			user := User{Name: "event_deleting_name", Avatar: "event_deleting_avatar"}
			s.Nil(query.Query().Create(&user))

			res, err := query.Query().Delete(&user)
			s.EqualError(err, "deleting error")
			s.Nil(res)

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.True(user1.ID > 0)
		})

		s.Run("not trigger when deleting mass records", func() {
			ctx := context.WithValue(context.Background(), "event", "deleting event with mass records")
			res, err := query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Where("name", "event_deleting_name").Delete(&User{})
			s.NoError(err)
			s.Equal(int64(1), res.RowsAffected)
		})
	}
}

func (s *QueryTestSuite) TestEvent_Deleted() {
	for _, query := range s.queries {
		s.Run("trigger", func() {
			user := User{Name: "event_deleted_name", Avatar: "event_deleted_avatar"}
			s.Nil(query.Query().Create(&user))

			res, err := query.Query().Delete(&user)
			s.EqualError(err, "deleted error")
			s.Nil(res)

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.True(user1.ID == 0)
		})

		s.Run("not trigger when deleting mass records", func() {
			ctx := context.WithValue(context.Background(), "event", "deleted event with mass records")
			res, err := query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Where("name", "event_deleted_name").Delete(&User{})
			s.NoError(err)
			s.Equal(int64(0), res.RowsAffected)
		})
	}
}

func (s *QueryTestSuite) TestEvent_ForceDeleting() {
	for _, query := range s.queries {
		s.Run("trigger", func() {
			user := User{Name: "event_force_deleting_name", Avatar: "event_force_deleting_avatar"}
			s.Nil(query.Query().Create(&user))

			res, err := query.Query().ForceDelete(&user)
			s.EqualError(err, "force deleting error")
			s.Nil(res)

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.True(user1.ID > 0)
		})

		s.Run("not trigger when force deleting mass records", func() {
			ctx := context.WithValue(context.Background(), "event", "force deleting event with mass records")
			res, err := query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Where("name", "event_force_deleting_name").ForceDelete(&User{})
			s.NoError(err)
			s.Equal(int64(1), res.RowsAffected)
		})
	}
}

func (s *QueryTestSuite) TestEvent_ForceDeleted() {
	for _, query := range s.queries {
		s.Run("trigger", func() {
			user := User{Name: "event_force_deleted_name", Avatar: "event_force_deleted_avatar"}
			s.Nil(query.Query().Create(&user))

			res, err := query.Query().ForceDelete(&user)
			s.EqualError(err, "force deleted error")
			s.Nil(res)

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.True(user1.ID == 0)
		})

		s.Run("not trigger when force deleting mass records", func() {
			ctx := context.WithValue(context.Background(), "event", "force deleted event with mass records")

			user := User{Name: "event_force_deleted_name", Avatar: "event_force_deleted_avatar"}
			s.Nil(query.Query().Create(&user))

			res, err := query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Where("name", "event_force_deleted_name").ForceDelete(&User{})
			s.NoError(err)
			s.Equal(int64(1), res.RowsAffected)
		})
	}
}

func (s *QueryTestSuite) TestEvent_Restored() {
	for _, query := range s.queries {
		user := User{Name: "event_restored_name", Avatar: "event_restored_avatar"}
		s.Nil(query.Query().Create(&user))

		res, err := query.Query().Delete(&user)
		s.NoError(err)
		s.Equal(int64(1), res.RowsAffected)

		res, err = query.Query().WithTrashed().Restore(&user)
		s.NoError(err)
		s.Equal(int64(1), res.RowsAffected)
		s.Equal("event_restored_name1", user.Name)

		var user1 User
		s.Nil(query.Query().Find(&user1, user.ID))
		s.True(user1.ID > 0)
		s.Equal("event_restored_name", user1.Name)
		s.Equal("event_restored_avatar", user1.Avatar)
	}
}

func (s *QueryTestSuite) TestEvent_Restoring() {
	for _, query := range s.queries {
		user := User{Name: "event_restoring_name", Avatar: "event_restoring_avatar"}
		s.Nil(query.Query().Create(&user))

		res, err := query.Query().Delete(&user)
		s.NoError(err)
		s.Equal(int64(1), res.RowsAffected)

		res, err = query.Query().WithTrashed().Restore(&user)
		s.NoError(err)
		s.Equal(int64(1), res.RowsAffected)
		s.Equal("event_restoring_name1", user.Name)

		var user1 User
		s.Nil(query.Query().Find(&user1, user.ID))
		s.True(user1.ID > 0)
		s.Equal("event_restoring_name", user1.Name)
		s.Equal("event_restoring_avatar", user1.Avatar)
	}
}

func (s *QueryTestSuite) TestEvent_Retrieved() {
	for _, query := range s.queries {
		user := User{Name: "event_retrieved_name"}
		s.Nil(query.Query().Create(&user))
		s.True(user.ID > 0)

		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when Find",
				setup: func() {
					var user1 User
					s.Nil(query.Query().Where("name", "event_retrieved_name").Find(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "not trigger when Find by slice struct",
				setup: func() {
					ctx := context.WithValue(context.Background(), "event", "retrieved event with slice struct")

					var users []User
					s.Nil(query.Query().(contractsorm.QueryWithContext).WithContext(ctx).Model(&User{}).Where("name", "event_retrieved_name").Find(&users))
					s.Equal(1, len(users))
					s.True(users[0].ID > 0)
					s.Equal("event_retrieved_name", users[0].Name)
				},
			},
			{
				name: "trigger when First",
				setup: func() {
					var user1 User
					s.Nil(query.Query().Where("name", "event_retrieved_name").First(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)

					var user2 User
					s.Nil(query.Query().Where("name", "event_retrieved_name1").First(&user2))
					s.True(user2.ID == 0)
					s.Empty(user2.Name)
				},
			},
			{
				name: "trigger when FirstOr",
				setup: func() {
					var user1 User
					s.Nil(query.Query().Where("name", "event_retrieved_name").Find(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user1 User
					s.Nil(query.Query().FirstOrCreate(&user1, User{Name: "event_retrieved_name"}))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrFail",
				setup: func() {
					var user1 User
					s.Nil(query.Query().Where("name", "event_retrieved_name").FirstOrFail(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrNew",
				setup: func() {
					var user1 User
					s.Nil(query.Query().FirstOrNew(&user1, User{Name: "event_retrieved_name"}))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrFail",
				setup: func() {
					var user1 User
					s.Nil(query.Query().Where("name", "event_retrieved_name").FirstOrFail(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_IsDirty() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "create",
				setup: func() {
					user := User{Name: "event_creating_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_creating_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "save",
				setup: func() {
					user := User{Name: "event_saving_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by single column",
				setup: func() {
					user := User{Name: "event_updating_single_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Model(&user).Update("name", "event_updating_single_update_IsDirty_name1")
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("event_updating_single_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_single_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updating_single_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_single_update_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by map",
				setup: func() {
					user := User{Name: "event_updating_map_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Model(&user).Update(map[string]any{
						"name": "event_updating_map_update_IsDirty_name1",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_updating_map_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_map_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updating_map_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_map_update_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by model",
				setup: func() {
					user := User{Name: "event_updating_model_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					res, err := query.Query().Model(&user).Update(User{
						Name: "event_updating_model_update_IsDirty_name1",
					})
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("event_updating_model_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_model_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_updating_model_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_model_update_IsDirty_avatar", user.Avatar)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestEvent_Context() {
	for _, query := range s.queries {
		user := User{Name: "event_context"}
		s.Nil(query.Query().Create(&user))
		s.Equal("goravel", user.Avatar)
	}
}

func (s *QueryTestSuite) TestEvent_Query() {
	for _, query := range s.queries {
		user := User{Name: "event_query"}
		s.Nil(query.Query().Create(&user))
		s.True(user.ID > 0)
		s.Equal("event_query", user.Name)

		var user1 User
		s.Nil(query.Query().Where("name", "event_query1").Find(&user1))
		s.True(user1.ID > 0)
	}
}

func (s *QueryTestSuite) TestExec() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			res, err := query.Query().Exec("INSERT INTO users (name, avatar, created_at, updated_at) VALUES ('exec_user', 'exec_avatar', '2023-03-09 18:56:33', '2023-03-09 18:56:35');")
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user User
			err = query.Query().Where("name", "exec_user").First(&user)
			s.Nil(err)
			s.True(user.ID > 0)

			res, err = query.Query().Exec(fmt.Sprintf("UPDATE users set name = 'exec_user1' where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			res, err = query.Query().Exec(fmt.Sprintf("DELETE FROM users where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)
		})
	}
}

func (s *QueryTestSuite) TestExists() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "exists_user", Avatar: "exists_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "exists_user", Avatar: "exists_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			exists, err := query.Query().Model(&User{}).Where("name = ?", "exists_user").Exists()
			s.Nil(err)
			s.True(exists)

			exists, err = query.Query().Model(&User{}).Where("name = ?", "no_exists_user").Exists()
			s.Nil(err)
			s.False(exists)
		})
	}
}

func (s *QueryTestSuite) TestFind() {
	for _, query := range s.queries {
		user := User{Name: "find_user"}
		s.Nil(query.Query().Create(&user))
		s.True(user.ID > 0)

		var user2 User
		s.Nil(query.Query().Find(&user2, user.ID))
		s.True(user2.ID > 0)

		var user3 []User
		s.Nil(query.Query().Find(&user3, []uint{user.ID}))
		s.Equal(1, len(user3))

		var user4 []User
		s.Nil(query.Query().Where("id in ?", []uint{user.ID}).Find(&user4))
		s.Equal(1, len(user4))
	}
}

func (s *QueryTestSuite) TestFindOrFail() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "find_user"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					var user2 User
					s.Nil(query.Query().FindOrFail(&user2, user.ID))
					s.True(user2.ID > 0)
				},
			},
			{
				name: "error",
				setup: func() {
					var user User
					s.ErrorIs(query.Query().FindOrFail(&user, 10000), errors.OrmRecordNotFound)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestFirst() {
	for _, query := range s.queries {
		user := User{Name: "first_user"}
		s.Nil(query.Query().Create(&user))
		s.True(user.ID > 0)

		var user1 User
		s.Nil(query.Query().Where("name", "first_user").First(&user1))
		s.True(user1.ID > 0)

		// refresh connection
		config := s.additionalQuery.Driver().Pool().Writers[0]
		config.Connection = "dummy"
		mockDatabaseConfig(query.MockConfig(), config)

		people := People{Body: "first_people"}
		s.Nil(query.Query().Create(&people))
		s.True(people.ID > 0)

		var people1 People
		s.Nil(query.Query().Where("id in ?", []uint{people.ID}).First(&people1))
		s.True(people1.ID > 0)
	}
}

func (s *QueryTestSuite) TestFirstOr() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "not found, new one",
				setup: func() {
					var user User
					s.Nil(query.Query().Where("name", "first_or_user").FirstOr(&user, func() error {
						user.Name = "goravel"

						return nil
					}))
					s.Equal(uint(0), user.ID)
					s.Equal("goravel", user.Name)

				},
			},
			{
				name: "found",
				setup: func() {
					user := User{Name: "first_or_name"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					var user1 User
					s.Nil(query.Query().Where("name", "first_or_name").Find(&user1))
					s.True(user1.ID > 0)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestFirstOrCreate() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "error when empty conditions",
				setup: func() {
					var user User
					s.EqualError(query.Query().FirstOrCreate(&user), errors.OrmQueryConditionRequired.Error())
					s.True(user.ID == 0)
				},
			},
			{
				name: "success",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrCreate(&user, User{Name: "first_or_create_user"}))
					s.True(user.ID > 0)
					s.Equal("first_or_create_user", user.Name)

					var user1 User
					s.Nil(query.Query().FirstOrCreate(&user1, User{Name: "first_or_create_user"}))
					s.Equal(user.ID, user1.ID)

					var user2 User
					s.Nil(query.Query().Where("avatar", "first_or_create_avatar").FirstOrCreate(&user2, User{Name: "user"}, User{Avatar: "first_or_create_avatar2"}))
					s.True(user2.ID > 0)
					s.True(user2.Avatar == "first_or_create_avatar2")
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestFirstOrFail() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "fail",
				setup: func() {
					var user User
					s.ErrorIs(query.Query().Where("name", "first_or_fail_user").FirstOrFail(&user), errors.OrmRecordNotFound)
					s.Equal(uint(0), user.ID)
				},
			},
			{
				name: "success",
				setup: func() {
					user := User{Name: "first_or_fail_name"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("first_or_fail_name", user.Name)

					var user1 User
					s.Nil(query.Query().Where("name", "first_or_fail_name").FirstOrFail(&user1))
					s.True(user1.ID > 0)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestFirstOrNew() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "not found, new one",
				setup: func() {
					var user User
					s.Nil(query.Query().FirstOrNew(&user, User{Name: "first_or_new_name"}))
					s.Equal(uint(0), user.ID)
					s.Equal("first_or_new_name", user.Name)
					s.Empty(user.Avatar)

					var user1 User
					s.Nil(query.Query().FirstOrNew(&user1, User{Name: "first_or_new_name"}, User{Avatar: "first_or_new_avatar"}))
					s.Equal(uint(0), user1.ID)
					s.Equal("first_or_new_name", user1.Name)
					s.Equal("first_or_new_avatar", user1.Avatar)
				},
			},
			{
				name: "found",
				setup: func() {
					user := User{Name: "first_or_new_name"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("first_or_new_name", user.Name)

					var user1 User
					s.Nil(query.Query().FirstOrNew(&user1, User{Name: "first_or_new_name"}))
					s.True(user1.ID > 0)
					s.Equal("first_or_new_name", user1.Name)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestForceDelete() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "force_delete_name"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("force_delete_name", user.Name)

					res, err := query.Query().Where("name", "force_delete_name").ForceDelete(&User{})
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("force_delete_name", user.Name)

					var user1 User
					s.Nil(query.Query().WithTrashed().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by table",
				setup: func() {
					user := User{Name: "force_delete_name_by_table"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("force_delete_name_by_table", user.Name)

					res, err := query.Query().Table("users").Where("name", "force_delete_name_by_table").ForceDelete()
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("force_delete_name_by_table", user.Name)

					var user1 User
					s.Nil(query.Query().WithTrashed().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by model",
				setup: func() {
					user := User{Name: "force_delete_name_by_model"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)
					s.Equal("force_delete_name_by_model", user.Name)

					res, err := query.Query().Model(&User{}).Where("name", "force_delete_name_by_model").ForceDelete()
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("force_delete_name_by_model", user.Name)

					var user1 User
					s.Nil(query.Query().WithTrashed().Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestGet() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "get_user"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			var user1 []User
			s.Nil(query.Query().Where("id in ?", []uint{user.ID}).Get(&user1))
			s.Equal(1, len(user1))

			// refresh connection
			config := s.additionalQuery.Driver().Pool().Writers[0]
			config.Connection = "dummy"
			mockDatabaseConfig(query.MockConfig(), config)

			people := People{Body: "get_people"}
			s.Nil(query.Query().Create(&people))
			s.True(people.ID > 0)

			var people1 []People
			s.Nil(query.Query().Where("id in ?", []uint{people.ID}).Get(&people1))
			s.Equal(1, len(people1))

			var user2 []User
			s.Nil(query.Query().Where("id in ?", []uint{user.ID}).Get(&user2))
			s.Equal(1, len(user2))
		})
	}
}

func (s *QueryTestSuite) TestGlobalScopes() {
	prepareData := func(query orm.Query) {
		globalScope1 := GlobalScope{Name: "global_scope_1"}
		s.Nil(query.Create(&globalScope1))
		s.True(globalScope1.ID > 0)

		globalScope := GlobalScope{Name: "global_scope"}
		s.Nil(query.Create(&globalScope))
		s.True(globalScope.ID > 0)
	}

	for driver, query := range s.queries {
		s.Run(driver, func() {
			s.Run("Count", func() {
				s.SetupTest()
				prepareData(query.Query())

				count, err := query.Query().Model(&GlobalScope{}).Count()
				s.Nil(err)
				s.Equal(int64(1), count)
			})

			s.Run("Cursor", func() {
				s.SetupTest()
				prepareData(query.Query())

				count := 0
				for cursor := range query.Query().Model(&GlobalScope{}).Cursor() {
					count++
					var globalScope GlobalScope
					s.Nil(cursor.Scan(&globalScope))
					s.True(globalScope.ID > 0)
					s.Equal("global_scope", globalScope.Name)
				}
				s.Equal(1, count)
			})

			s.Run("Delete", func() {
				s.SetupTest()
				prepareData(query.Query())

				res, err := query.Query().Model(&GlobalScope{}).Delete()
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))
			})

			s.Run("Exec", func() {
				s.SetupTest()
				prepareData(query.Query())

				res, err := query.Query().Exec("delete from global_scopes")
				s.Nil(err)
				s.Equal(int64(2), res.RowsAffected)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))
			})

			s.Run("Exists", func() {
				s.SetupTest()
				prepareData(query.Query())

				exists, err := query.Query().Model(&GlobalScope{}).Exists()
				s.Nil(err)
				s.True(exists)

				res, err := query.Query().Model(&GlobalScope{}).Delete()
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				exists, err = query.Query().Model(&GlobalScope{}).Exists()
				s.Nil(err)
				s.False(exists)
			})

			s.Run("Find", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().Find(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("FindOrFail", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().FindOrFail(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)

				var globalScope1 GlobalScope
				s.EqualError(query.Query().Where("name", "global_scope_1").FindOrFail(&globalScope1), errors.OrmRecordNotFound.Error())
				s.Equal(uint(0), globalScope1.ID)
			})

			s.Run("First", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().First(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("FirstOr", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().FirstOr(&globalScope, func() error {
					return errors.OrmRecordNotFound
				}))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("FirstOrCreate", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().FirstOrCreate(&globalScope, User{Name: "global_scope"}))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("FirstOrFail", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().FirstOrFail(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("FirstOrNew", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().FirstOrNew(&globalScope, User{Name: "global_scope"}))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)
			})

			s.Run("ForceDelete", func() {
				s.SetupTest()
				prepareData(query.Query())

				res, err := query.Query().Model(&GlobalScope{}).ForceDelete()
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))
			})

			s.Run("Get", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(1, len(globalScopes))
				s.True(globalScopes[0].ID > 0)
				s.Equal("global_scope", globalScopes[0].Name)
			})

			s.Run("Paginate", func() {
				s.SetupTest()
				prepareData(query.Query())

				var (
					globalScopes []GlobalScope
					total        int64
				)
				s.Nil(query.Query().Paginate(1, 2, &globalScopes, &total))
				s.Equal(1, len(globalScopes))
				s.True(globalScopes[0].ID > 0)
				s.Equal("global_scope", globalScopes[0].Name)
				s.Equal(int64(1), total)
			})

			s.Run("Pluck", func() {
				s.SetupTest()
				prepareData(query.Query())

				var names []string
				s.Nil(query.Query().Model(&GlobalScope{}).Pluck("name", &names))
				s.Equal(1, len(names))
				s.Equal("global_scope", names[0])
			})

			s.Run("Restore", func() {
				s.SetupTest()
				prepareData(query.Query())

				res, err := query.Query().Model(&GlobalScope{}).Delete()
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))

				res, err = query.Query().Model(&GlobalScope{}).WithTrashed().Restore()
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(1, len(globalScopes))
				s.True(globalScopes[0].ID > 0)
				s.Equal("global_scope", globalScopes[0].Name)
			})

			s.Run("Save", func() {
				s.SetupTest()
				prepareData(query.Query())

				globalScope := GlobalScope{Name: "global_scope"}
				s.Nil(query.Query().Save(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(2, len(globalScopes))
				s.True(globalScopes[0].ID > 0)
				s.Equal("global_scope", globalScopes[0].Name)
				s.True(globalScopes[1].ID > 0)
				s.Equal("global_scope", globalScopes[1].Name)

				globalScope.Name = "global_scope_1"
				s.Nil(query.Query().Save(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope_1", globalScope.Name)

				var globalScopes1 []GlobalScope
				s.Nil(query.Query().Get(&globalScopes1))
				s.Equal(1, len(globalScopes1))
				s.True(globalScopes1[0].ID > 0)
				s.Equal("global_scope", globalScopes1[0].Name)
			})

			s.Run("Scan", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScopes []GlobalScope
				s.Nil(query.Query().Raw("SELECT id, name, created_at, updated_at, deleted_at FROM global_scopes").Scan(&globalScopes))
				s.Equal(2, len(globalScopes))
				s.True(globalScopes[0].ID > 0)
				s.Equal("global_scope_1", globalScopes[0].Name)
				s.True(globalScopes[1].ID > 0)
				s.Equal("global_scope", globalScopes[1].Name)
			})

			s.Run("Sum", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().First(&globalScope))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope", globalScope.Name)

				var sum int64
				err := query.Query().Model(&GlobalScope{}).Sum("id", &sum)
				s.Nil(err)
				s.Equal(globalScope.ID, uint(sum))
			})

			s.Run("Update", func() {
				s.SetupTest()
				prepareData(query.Query())

				res, err := query.Query().Model(&GlobalScope{}).Update("name", "global_scope_1")
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))
			})

			s.Run("UpdateOrCreate", func() {
				s.SetupTest()
				prepareData(query.Query())

				var globalScope GlobalScope
				s.Nil(query.Query().UpdateOrCreate(&globalScope, GlobalScope{Name: "global_scope"}, GlobalScope{Name: "global_scope_1"}))
				s.True(globalScope.ID > 0)
				s.Equal("global_scope_1", globalScope.Name)

				var globalScopes []GlobalScope
				s.Nil(query.Query().Get(&globalScopes))
				s.Equal(0, len(globalScopes))
			})
		})
	}
}

func (s *QueryTestSuite) TestJoin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "join_user", Avatar: "join_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			userAddress := Address{UserID: user.ID, Name: "join_address", Province: "join_province"}
			s.Nil(query.Query().Create(&userAddress))
			s.True(userAddress.ID > 0)

			type Result struct {
				UserName        string
				UserAddressName string
			}
			var result []Result
			s.Nil(query.Query().Model(&User{}).Where("users.id = ?", user.ID).Join("left join addresses ua on users.id = ua.user_id").
				Select("users.name user_name, ua.name user_address_name").Get(&result))
			s.Equal(1, len(result))
			s.Equal("join_user", result[0].UserName)
			s.Equal("join_address", result[0].UserAddressName)
		})
	}
}

func (s *QueryTestSuite) TestLockForUpdate() {
	for driver, query := range s.queries {
		if driver == sqlite.Name {
			continue
		}

		s.Run(driver, func() {
			user := User{Name: "lock_for_update_user"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			for i := 0; i < 10; i++ {
				go func() {
					tx, err := query.Query().Begin()
					s.Nil(err)

					var user1 User
					s.Nil(tx.LockForUpdate().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					user1.Name += "1"
					s.Nil(tx.Save(&user1))

					s.Nil(tx.Commit())
				}()
			}

			time.Sleep(2 * time.Second)

			var user2 User
			s.Nil(query.Query().Find(&user2, user.ID))
			s.Equal("lock_for_update_user1111111111", user2.Name)
		})
	}
}

func (s *QueryTestSuite) TestOffset() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "offset_user", Avatar: "offset_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "offset_user", Avatar: "offset_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Query().Where("name = ?", "offset_user").Offset(1).Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestOrder() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "order_user", Avatar: "order_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_user", Avatar: "order_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Query().Where("name = ?", "order_user").OrderByRaw("id desc, name asc").Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestOrderBy() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "order_asc_user", Avatar: "order_asc_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_asc_user", Avatar: "order_asc_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users1 []User
			s.Nil(query.Query().Where("name = ?", "order_asc_user").OrderBy("id").Get(&users1))
			s.True(len(users1) == 2)
			s.True(users1[0].ID == user.ID)

			var users2 []User
			s.Nil(query.Query().Where("name = ?", "order_asc_user").OrderBy("id", "DESC").Get(&users2))
			s.True(len(users2) == 2)
			s.True(users2[0].ID == user1.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrderByDesc() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "order_desc_user", Avatar: "order_desc_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_desc_user", Avatar: "order_desc_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "order_desc_user").OrderByDesc("id").Get(&users))
			usersLength := len(users)
			s.True(usersLength == 2)
			s.True(users[usersLength-1].ID == user.ID)
		})
	}
}

func (s *QueryTestSuite) TestInRandomOrder() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			for i := 0; i < 30; i++ {
				user := User{Name: "random_order_user", Avatar: "random_order_avatar"}
				s.Nil(query.Query().Create(&user))
				s.True(user.ID > 0)
			}

			var users1 []User
			s.Nil(query.Query().Where("name = ?", "random_order_user").InRandomOrder().Find(&users1))
			s.True(len(users1) == 30)

			var users2 []User
			s.Nil(query.Query().Where("name = ?", "random_order_user").InRandomOrder().Find(&users2))
			s.True(len(users2) == 30)

			s.True(users1[0].ID != users2[0].ID || users1[14].ID != users2[14].ID || users1[29].ID != users2[29].ID)
		})
	}
}

func (s *QueryTestSuite) TestInTransaction() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			s.False(query.Query().InTransaction())

			tx, err := query.Query().Begin()
			s.NotNil(tx)
			s.NoError(err)

			s.True(tx.InTransaction())
			s.NoError(tx.Commit())
			s.False(query.Query().InTransaction())

			tx, err = query.Query().Begin()
			s.NotNil(tx)
			s.NoError(err)

			s.True(tx.InTransaction())
			s.NoError(tx.Rollback())
			s.False(query.Query().InTransaction())
		})
	}
}

func (s *QueryTestSuite) TestPaginate() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "paginate_user", Avatar: "paginate_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "paginate_user", Avatar: "paginate_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "paginate_user", Avatar: "paginate_avatar2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "paginate_user", Avatar: "paginate_avatar3"}
			s.Nil(query.Query().Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "paginate_user").Paginate(1, 3, &users, nil))
			s.Equal(3, len(users))

			var users1 []User
			var total1 int64
			s.Nil(query.Query().Where("name = ?", "paginate_user").Paginate(2, 3, &users1, &total1))
			s.Equal(1, len(users1))
			s.Equal(int64(4), total1)

			var users2 []User
			var total2 int64
			s.Nil(query.Query().Model(User{}).Where("name = ?", "paginate_user").Paginate(1, 3, &users2, &total2))
			s.Equal(3, len(users2))
			s.Equal(int64(4), total2)

			var users3 []User
			var total3 int64
			s.Nil(query.Query().Table("users").Where("name = ?", "paginate_user").Paginate(1, 3, &users3, &total3))
			s.Equal(3, len(users3))
			s.Equal(int64(4), total3)
		})
	}
}

func (s *QueryTestSuite) TestPluck() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "pluck_user", Avatar: "pluck_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "pluck_user", Avatar: "pluck_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var avatars []string
			s.Nil(query.Query().Model(&User{}).Where("name = ?", "pluck_user").Pluck("avatar", &avatars))
			s.Equal(2, len(avatars))
			s.Equal("pluck_avatar", avatars[0])
			s.Equal("pluck_avatar1", avatars[1])
		})
	}
}

func (s *QueryTestSuite) TestHasOne() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "has_one_name",
				Address: &Address{
					Name: "has_one_address",
				},
			}

			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var user1 User
			s.Nil(query.Query().With("Address").Where("name = ?", "has_one_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestHasOneMorph() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "has_one_morph_name",
				House: &House{
					Name: "has_one_morph_house",
				},
			}
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.House.ID > 0)

			var user1 User
			s.Nil(query.Query().With("House").Where("name = ?", "has_one_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_one_morph_name")
			s.True(user.House.ID > 0)
			s.True(user.House.Name == "has_one_morph_house")

			var house House
			s.Nil(query.Query().Where("name = ?", "has_one_morph_house").Where("houseable_type = ?", "users").Where("houseable_id = ?", user.ID).First(&house))
			s.True(house.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestHasMany() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "has_many_name",
				Books: []*Book{
					{Name: "has_many_book1"},
					{Name: "has_many_book2"},
				},
			}

			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			var user1 User
			s.Nil(query.Query().With("Books").Where("name = ?", "has_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Books) == 2)
		})
	}
}

func (s *QueryTestSuite) TestHasManyMorph() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "has_many_morph_name",
				Phones: []*Phone{
					{Name: "has_many_morph_phone1"},
					{Name: "has_many_morph_phone2"},
				},
			}
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Phones[0].ID > 0)
			s.True(user.Phones[1].ID > 0)

			var user1 User
			s.Nil(query.Query().With("Phones").Where("name = ?", "has_many_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_many_morph_name")
			s.True(len(user.Phones) == 2)
			s.True(user.Phones[0].Name == "has_many_morph_phone1")
			s.True(user.Phones[1].Name == "has_many_morph_phone2")

			var phones []Phone
			s.Nil(query.Query().Where("name like ?", "has_many_morph_phone%").Where("phoneable_type = ?", "users").Where("phoneable_id = ?", user.ID).Find(&phones))
			s.True(len(phones) == 2)
		})
	}
}

func (s *QueryTestSuite) TestManyToMany() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := &User{
				Name: "many_to_many_name",
				Roles: []*Role{
					{Name: "many_to_many_role1"},
					{Name: "many_to_many_role2"},
				},
			}

			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Roles[0].ID > 0)
			s.True(user.Roles[1].ID > 0)

			var user1 User
			s.Nil(query.Query().With("Roles").Where("name = ?", "many_to_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Roles) == 2)

			var role Role
			s.Nil(query.Query().With("Users").Where("name = ?", "many_to_many_role1").First(&role))
			s.True(role.ID > 0)
			s.True(len(role.Users) == 1)
			s.Equal("many_to_many_name", role.Users[0].Name)
		})
	}
}

func (s *QueryTestSuite) TestLimit() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "limit_user", Avatar: "limit_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "limit_user", Avatar: "limit_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Query().Where("name = ?", "limit_user").Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestLoad() {
	for _, query := range s.queries {
		user := User{Name: "load_user", Address: &Address{}, Books: []*Book{{}, {}}, Roles: []*Role{{}, {}}}
		user.Address.Name = "load_address"
		user.Books[0].Name = "load_book0"
		user.Books[0].Author = &Author{Name: "load_book0_author"}
		user.Books[1].Name = "load_book1"
		user.Roles[0].Name = "load_role0"
		user.Roles[1].Name = "load_role1"

		s.Nil(query.Query().Select(gorm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Address.ID > 0)
		s.True(user.Books[0].ID > 0)
		s.True(user.Books[1].ID > 0)
		s.True(user.Books[0].Author.ID > 0)
		s.True(user.Roles[0].ID > 0)
		s.True(user.Roles[1].ID > 0)

		tests := []struct {
			description string
			setup       func(description string)
		}{
			{
				description: "simple load relationship",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.True(len(user1.Books) == 0)
					s.Nil(query.Query().Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 0)
					s.Nil(query.Query().Load(&user1, "Books"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 2)
				},
			},
			{
				description: "load relationship with simple condition",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(query.Query().Load(&user1, "Books", "name = ?", "load_book0"))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(1, len(user1.Books))
					s.Equal("load_book0", user.Books[0].Name)
				},
			},
			{
				description: "load relationship with func condition",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(query.Query().Load(&user1, "Books", func(query contractsorm.Query) contractsorm.Query {
						return query.Where("name = ?", "load_book0")
					}))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(1, len(user1.Books))
					s.Equal("load_book0", user.Books[0].Name)
				},
			},
			{
				description: "load nested relationship",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Equal("load_user", user1.Name)
					s.Nil(query.Query().Load(&user1, "Books.Author"))
					s.True(user1.Books[0].ID > 0)
					s.Equal("load_book0", user1.Books[0].Name)
					s.True(user1.Books[0].Author.ID > 0)
					s.Equal("load_book0_author", user1.Books[0].Author.Name)
					s.True(user1.Books[1].ID > 0)
					s.Equal("load_book1", user1.Books[1].Name)
					s.Nil(user1.Books[1].Author)
					s.Nil(query.Query().Load(&user1, "Roles.Users"))
					s.Equal("load_role0", user1.Roles[0].Name)
					s.Equal("load_user", user1.Roles[0].Users[0].Name)
					s.Equal("load_role1", user1.Roles[1].Name)
				},
			},
			{
				description: "error when relation is empty",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.EqualError(query.Query().Load(&user1, ""), errors.OrmQueryEmptyRelation.Error())
				},
			},
			{
				description: "error when id is nil",
				setup: func(description string) {
					type UserNoID struct {
						Name   string
						Avatar string
					}
					var userNoID UserNoID
					s.EqualError(query.Query().Load(&userNoID, "Book"), errors.OrmQueryEmptyId.Error())
				},
			},
		}
		for _, test := range tests {
			s.Run(test.description, func() {
				test.setup(test.description)
			})
		}
	}
}

func (s *QueryTestSuite) TestLoadMissing() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "load_missing_user", Address: &Address{}, Books: []*Book{{}, {}}}
			user.Address.Name = "load_missing_address"
			user.Books[0].Name = "load_missing_book0"
			user.Books[1].Name = "load_missing_book1"
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "load when missing",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Query().Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 0)
						s.Nil(query.Query().LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 0)
						s.Nil(query.Query().LoadMissing(&user1, "Books"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 2)
					},
				},
				{
					description: "don't load when not missing",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Query().With("Books", "name = ?", "load_missing_book0").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 1)
						s.Nil(query.Query().LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Nil(query.Query().LoadMissing(&user1, "Books"))
						s.True(len(user1.Books) == 1)
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *QueryTestSuite) TestModel() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			// model is valid
			user := User{Name: "model_user"}
			s.Nil(query.Query().Model(&User{}).Create(&user))
			s.True(user.ID > 0)

			// model is invalid
			user1 := User{Name: "model_user"}
			s.EqualError(query.Query().Model("users").Create(&user1), "unsupported data type: users: Table not set, please set it like: db.Model(&user) or db.Table(\"users\")")
		})
	}
}

func (s *QueryTestSuite) TestRaw() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "raw_user", Avatar: "raw_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			var user1 User
			s.Nil(query.Query().Raw("SELECT id, name FROM users WHERE name = ?", "raw_user").Scan(&user1))
			s.True(user1.ID > 0)
			s.Equal("raw_user", user1.Name)
			s.Empty(user1.Avatar)
		})
	}
}

func (s *QueryTestSuite) TestReuse() {
	for _, query := range s.queries {
		users := []User{{Name: "reuse_user", Avatar: "reuse_avatar"}, {Name: "reuse_user1", Avatar: "reuse_avatar1"}}
		s.Nil(query.Query().Create(&users))
		s.True(users[0].ID > 0)
		s.True(users[1].ID > 0)

		q := query.Query().Where("name", "reuse_user")

		var users1 User
		s.Nil(q.Where("avatar", "reuse_avatar").Find(&users1))
		s.True(users1.ID > 0)

		var users2 User
		s.Nil(q.Where("avatar", "reuse_avatar1").Find(&users2))
		s.True(users2.ID == 0)

		var users3 User
		s.Nil(query.Query().Where("avatar", "reuse_avatar1").Find(&users3))
		s.True(users3.ID > 0)
	}
}

func (s *QueryTestSuite) TestSave() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success when create",
				setup: func() {
					user := User{
						Name:   "save_create_user",
						Avatar: "save_create_avatar",
						House:  &House{Name: "save_create_house"},
					}
					s.Nil(query.Query().Save(&user))
					s.True(user.ID > 0)
					s.True(user.House.ID == 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_create_user", user1.Name)
				},
			},
			{
				name: "success when create with Select",
				setup: func() {
					user := User{
						Name:   "save_create_with_select_user",
						Avatar: "save_create_with_select_avatar",
						House:  &House{Name: "save_create_with_select_house"},
					}
					s.Nil(query.Query().Select("Name", "House").Save(&user))
					s.True(user.ID > 0)
					s.True(user.House.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_create_with_select_user", user1.Name)
					s.Empty(user1.Avatar)

					var house1 House
					s.Nil(query.Query().Find(&house1, user.House.ID))
					s.Equal("save_create_with_select_house", house1.Name)
				},
			},
			{
				name: "success when create with Omit",
				setup: func() {
					user := User{
						Name:   "save_create_with_omit_user",
						Avatar: "save_create_with_omit_avatar",
						House:  &House{Name: "save_create_with_omit_house"},
					}
					s.Nil(query.Query().Omit("House").Save(&user))
					s.True(user.ID > 0)
					s.True(user.House.ID == 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_create_with_omit_user", user1.Name)
					s.Equal("save_create_with_omit_avatar", user1.Avatar)
				},
			},
			{
				name: "success when update",
				setup: func() {
					user := User{Name: "save_update_user", Avatar: "save_update_avatar"}
					s.Nil(query.Query().Create(&user))
					s.True(user.ID > 0)

					user.Name = "save_update_user1"
					s.Nil(query.Query().Save(&user))

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_update_user1", user1.Name)
				},
			},
			{
				name: "success when update with Select",
				setup: func() {
					user := User{
						Name:   "save_update_with_select_user",
						Avatar: "save_update_with_select_avatar",
						House:  &House{Name: "save_update_with_select_house"},
					}
					s.Nil(query.Query().Select("Name", "Avatar").Create(&user))
					s.True(user.ID > 0)
					s.True(user.House.ID == 0)

					user.Name = "save_update_with_select_user1"
					s.Nil(query.Query().Select("Name", "House").Save(&user))

					s.True(user.House.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_update_with_select_user1", user1.Name)

					var house1 House
					s.Nil(query.Query().Find(&house1, user.House.ID))
					s.Equal("save_update_with_select_house", house1.Name)
				},
			},
			{
				name: "success when update with Omit",
				setup: func() {
					user := User{
						Name:    "save_update_with_omit_user",
						Avatar:  "save_update_with_omit_avatar",
						Address: &Address{Name: "save_update_with_omit_address"},
						House:   &House{Name: "save_update_with_omit_house"},
					}
					s.Nil(query.Query().Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)
					s.True(user.House.ID == 0)

					user.Name = "save_update_with_omit_user1"
					user.Address.Name = "save_update_with_omit_address1"
					s.Nil(query.Query().Omit("Address").Save(&user))

					s.True(user.House.ID > 0)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("save_update_with_omit_user1", user1.Name)

					var address1 Address
					s.Nil(query.Query().Find(&address1, user.Address.ID))
					s.Equal("save_update_with_omit_address", address1.Name)

					var house1 House
					s.Nil(query.Query().Find(&house1, user.House.ID))
					s.Equal("save_update_with_omit_house", house1.Name)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestSaveQuietly() {
	for _, query := range s.queries {
		user := User{Name: "event_save_quietly_name", Avatar: "save_quietly_avatar"}
		s.Nil(query.Query().SaveQuietly(&user))
		s.True(user.ID > 0)
		s.Equal("event_save_quietly_name", user.Name)
		s.Equal("save_quietly_avatar", user.Avatar)

		var user1 User
		s.Nil(query.Query().Find(&user1, user.ID))
		s.Equal("event_save_quietly_name", user1.Name)
		s.Equal("save_quietly_avatar", user1.Avatar)
	}
}

func (s *QueryTestSuite) TestScope() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{{Name: "scope_user", Avatar: "scope_avatar"}, {Name: "scope_user1", Avatar: "scope_avatar1"}}
			s.Nil(query.Query().Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)

			var users1 []User
			s.Nil(query.Query().Scopes(paginator("1", "1")).Find(&users1))

			s.Equal(1, len(users1))
			s.True(users1[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestSelect() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "select_user", Avatar: "select_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "select_user", Avatar: "select_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "select_user1", Avatar: "select_avatar1"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			type Result struct {
				Name  string
				Count string
			}
			var result []Result
			s.Nil(query.Query().Model(&User{}).Select("name, count(avatar) as count").Where("id in ?", []uint{user.ID, user1.ID, user2.ID}).GroupBy("name").Get(&result))
			s.Equal(2, len(result))
			s.Equal("select_user", result[0].Name)
			s.Equal("2", result[0].Count)
			s.Equal("select_user1", result[1].Name)
			s.Equal("1", result[1].Count)

			var result1 []Result
			s.Nil(query.Query().Model(&User{}).Select("name, count(avatar) as count").GroupBy("name").Having("name = ?", "select_user").Get(&result1))

			s.Equal(1, len(result1))
			s.Equal("select_user", result1[0].Name)
			s.Equal("2", result1[0].Count)
		})
	}
}

func (s *QueryTestSuite) TestSelectRaw() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "select_user", Avatar: "select_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "select_user", Avatar: "select_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "select_user1", Avatar: "select_avatar1"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			type Result struct {
				Name string
				Bio  string
			}
			var result []Result
			s.Nil(query.Query().Model(&User{}).SelectRaw("name, COALESCE(bio,?) as bio", "a").Where("id in ?", []uint{user.ID, user1.ID, user2.ID}).Get(&result))
			s.Equal(3, len(result))
			s.Equal("select_user", result[0].Name)
			s.Equal("a", result[0].Bio)
			s.Equal("select_user", result[1].Name)
			s.Equal("a", result[1].Bio)
			s.Equal("select_user1", result[2].Name)
			s.Equal("a", result[2].Bio)
		})
	}
}

func (s *QueryTestSuite) TestSharedLock() {
	for driver, query := range s.queries {
		if driver == sqlite.Name {
			continue
		}

		s.Run(driver, func() {
			user := User{Name: "shared_lock_user"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			tx, err := query.Query().Begin()
			s.Nil(err)
			var user1 User
			s.Nil(tx.SharedLock().Find(&user1, user.ID))
			s.True(user1.ID > 0)

			var user2 User
			s.Nil(query.Query().SharedLock().Find(&user2, user.ID))
			s.True(user2.ID > 0)

			user1.Name += "1"
			s.Nil(tx.Save(&user1))

			s.Nil(tx.Commit())

			var user3 User
			s.Nil(query.Query().Find(&user3, user.ID))
			s.Equal("shared_lock_user1", user3.Name)
		})
	}
}

func (s *QueryTestSuite) TestSoftDelete() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "soft_delete_user", Avatar: "soft_delete_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			res, err := query.Query().Where("name = ?", "soft_delete_user").Delete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user1 User
			s.Nil(query.Query().Find(&user1, user.ID))
			s.Equal(uint(0), user1.ID)

			var user2 User
			s.Nil(query.Query().WithTrashed().Find(&user2, user.ID))
			s.True(user2.ID > 0)

			res, err = query.Query().Where("name = ?", "soft_delete_user").ForceDelete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user3 User
			s.Nil(query.Query().WithTrashed().Find(&user3, user.ID))
			s.Equal(uint(0), user3.ID)
		})
	}
}

func (s *QueryTestSuite) TestRestore() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "restore_user1", Avatar: "restore_avatar"},
				{Name: "restore_user2", Avatar: "restore_avatar"},
				{Name: "restore_user3", Avatar: "restore_avatar"},
				{Name: "restore_user4", Avatar: "restore_avatar"},
			}
			s.NoError(query.Query().Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)
			s.True(users[2].ID > 0)
			s.True(users[3].ID > 0)

			res, err := query.Query().Where("avatar", "restore_avatar").Delete(&User{})
			s.Equal(int64(4), res.RowsAffected)
			s.NoError(err)

			res, err = query.Query().Where("name", "restore_user1").Restore(&User{})
			s.Equal(int64(0), res.RowsAffected)
			s.NoError(err)

			res, err = query.Query().WithTrashed().Where("name", "restore_user1").Restore(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.NoError(err)

			res, err = query.Query().Model(&User{}).WithTrashed().Where("name", "restore_user2").Restore()
			s.Equal(int64(1), res.RowsAffected)
			s.NoError(err)

			res, err = query.Query().Model(users[2]).WithTrashed().Restore()
			s.Equal(int64(1), res.RowsAffected)
			s.NoError(err)

			res, err = query.Query().WithTrashed().Restore(&users[3])
			s.Equal(int64(1), res.RowsAffected)
			s.NoError(err)

			count, err := query.Query().Model(&User{}).Where("avatar", "restore_avatar").Count()
			s.NoError(err)
			s.Equal(int64(4), count)
		})
	}
}

func (s *QueryTestSuite) TestSum() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var sum int64
			err := query.Query().Table("users").Sum("id", &sum)
			s.Nil(err)
			s.Equal(int64(3), sum)

			err = query.Query().Table("users").Sum("id", nil)
			s.Error(err)
		})
	}
}

func (s *QueryTestSuite) TestAvg() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "avg_user", Avatar: "avg_avatar", Ratio: 10}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "avg_user", Avatar: "avg_avatar1", Ratio: 20}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var avg float64
			err := query.Query().Table("users").Avg("ratio", &avg)
			s.Nil(err)
			s.Equal(float64(15), avg)

			err = query.Query().Table("users").Where("id", user.ID).Avg("ratio", nil)
			s.Error(err)
		})
	}
}

func (s *QueryTestSuite) TestMin() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "min_user", Avatar: "min_avatar", Ratio: 11}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "min_user", Avatar: "min_avatar1", Ratio: 9}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "min_user", Avatar: "min_avatar2", Ratio: 10}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			var min int64
			err := query.Query().Table("users").Min("ratio", &min)
			s.Nil(err)
			s.Equal(int64(9), min)

			err = query.Query().Table("users").Min("ratio", nil)
			s.Error(err)
		})
	}
}

func (s *QueryTestSuite) TestMax() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "max_user", Avatar: "max_avatar", Ratio: 10}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "max_user", Avatar: "max_avatar1", Ratio: 20}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "max_user", Avatar: "max_avatar2", Ratio: 30}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			var max int64
			err := query.Query().Table("users").Max("ratio", &max)
			s.Nil(err)
			s.Equal(int64(30), max)

			err = query.Query().Table("users").Max("ratio", nil)
			s.Error(err)
		})
	}
}

func (s *QueryTestSuite) TestToSql() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			switch driver {
			case postgres.Name:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", query.Query().Where("id", 1).ToSql().Find(User{}))
			case sqlserver.Name:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = @p1 AND \"users\".\"deleted_at\" IS NULL", query.Query().Where("id", 1).ToSql().Find(User{}))
			default:
				s.Equal("SELECT * FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", query.Query().Where("id", 1).ToSql().Find(User{}))
			}
		})
	}
}

func (s *QueryTestSuite) TestToRawSql() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			switch driver {
			case postgres.Name:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", query.Query().Where("id", 1).ToRawSql().Find(User{}))
			case sqlserver.Name:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1$ AND \"users\".\"deleted_at\" IS NULL", query.Query().Where("id", 1).ToRawSql().Find(User{}))
			default:
				s.Equal("SELECT * FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", query.Query().Where("id", 1).ToRawSql().Find(User{}))
			}
		})
	}
}

func (s *QueryTestSuite) TestTransactionSuccess() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
			user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
			tx, err := query.Query().Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Commit())

			var user2, user3 User
			s.Nil(query.Query().Find(&user2, user.ID))
			s.Nil(query.Query().Find(&user3, user1.ID))
		})
	}
}

func (s *QueryTestSuite) TestTransactionError() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			tx, err := query.Query().Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Rollback())

			var users []User
			s.Nil(query.Query().Where("name = ? or name = ?", "transaction_error_user", "transaction_error_user1").Find(&users))
			s.Equal(0, len(users))
		})
	}
}

func (s *QueryTestSuite) TestUpdate() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "update single column, success",
				setup: func() {
					users := []User{{Name: "updates_single_name", Avatar: "updates_single_avatar"}, {Name: "updates_single_name", Avatar: "updates_single_avatar1"}}
					s.Nil(query.Query().Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Query().Model(&User{}).Where("name = ?", "updates_single_name").Update("avatar", "update_single_avatar2")
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Query().Find(&user2, users[0].ID))
					s.Equal("update_single_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Query().Find(&user3, users[1].ID))
					s.Equal("update_single_avatar2", user3.Avatar)
				},
			},
			{
				name: "update columns by map, success",
				setup: func() {
					users := []User{{Name: "update_map_name", Avatar: "update_map_avatar"}, {Name: "update_map_name", Avatar: "update_map_avatar1"}}
					s.Nil(query.Query().Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Query().Model(&User{}).Where("name = ?", "update_map_name").Update(map[string]any{
						"avatar": "update_map_avatar2",
					})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Query().Find(&user2, users[0].ID))
					s.Equal("update_map_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Query().Find(&user3, users[0].ID))
					s.Equal("update_map_avatar2", user3.Avatar)
				},
			},
			{
				name: "update columns by model, success",
				setup: func() {
					users := []User{{Name: "update_model_name", Avatar: "update_model_avatar"}, {Name: "update_model_name", Avatar: "update_model_avatar1"}}
					s.Nil(query.Query().Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Query().Model(&User{}).Where("name = ?", "update_model_name").Update(User{Avatar: "update_model_avatar2"})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Query().Find(&user2, users[0].ID))
					s.Equal("update_model_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Query().Find(&user3, users[0].ID))
					s.Equal("update_model_avatar2", user3.Avatar)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestUpdateOrCreate() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			var user User
			err := query.Query().UpdateOrCreate(&user, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar"})
			s.Nil(err)
			s.True(user.ID > 0)

			var user1 User
			err = query.Query().Where("name", "update_or_create_user").Find(&user1)
			s.Nil(err)
			s.True(user1.ID > 0)

			var user2 User
			err = query.Query().UpdateOrCreate(&user2, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar1"})
			s.Nil(err)
			s.True(user2.ID > 0)
			s.Equal("update_or_create_avatar1", user2.Avatar)

			var user3 User
			err = query.Query().Where("avatar", "update_or_create_avatar1").Find(&user3)
			s.Nil(err)
			s.True(user3.ID > 0)

			count, err := query.Query().Model(User{}).Where("name", "update_or_create_user").Count()
			s.Nil(err)
			s.Equal(int64(1), count)
		})
	}
}

func (s *QueryTestSuite) TestWhere() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_user", Avatar: "where_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_user1", Avatar: "where_avatar1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Query().Where("name = ?", "where_user").OrWhere("avatar = ?", "where_avatar1").Find(&user2))
			s.Equal(2, len(user2))

			var user3 User
			s.Nil(query.Query().Where("name = 'where_user'").Find(&user3))
			s.True(user3.ID > 0)

			var user4 User
			s.Nil(query.Query().Where("name", "where_user").Find(&user4))
			s.True(user4.ID > 0)

			var user5 User
			s.Nil(query.Query().Where(func(query contractsorm.Query) contractsorm.Query {
				return query.Where("name = ?", "where_user").OrWhere("name", "where_user1")
			}).Where("avatar", "where_avatar").Find(&user5))
			s.True(user5.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestWhereIn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().WhereIn("id", []any{user.ID, user1.ID}).Find(&users))
			s.True(len(users) == 2)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereIn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Where("id = ?", -1).OrWhereIn("id", []any{user.ID, user1.ID}).Find(&users))
			s.True(len(users) == 2)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotIn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_in_user_2", Avatar: "where_in_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			var user3 User
			s.Nil(query.Query().Where("id = ?", user2.ID).WhereNotIn("id", []any{user.ID, user1.ID}).First(&user3))
			s.True(user3.ID == user2.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNotIn() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_in_user_2", Avatar: "where_in_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			var users []User
			s.Nil(query.Query().Where("id = ?", -1).OrWhereNotIn("id", []any{user.ID, user1.ID}).Find(&users))
			var user2Found bool
			for _, user := range users {
				if user.ID == user2.ID {
					user2Found = true
				}
			}
			s.True(user2Found)
		})
	}
}

func (s *QueryTestSuite) TestWhereBetween() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			now := carbon.Now()

			user := User{Name: "where_between_user", Avatar: "where_between_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_between_user_1", Avatar: "where_between_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_between_user_2", Avatar: "where_between_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			var users []User
			s.Nil(query.Query().WhereBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 3)

			var users1 []User
			s.Nil(query.Query().WhereBetween("created_at", now.Copy().SubDay(), now.AddDay()).Find(&users1))
			s.True(len(users1) == 3)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotBetween() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			now := carbon.Now().AddSecond()
			time.Sleep(2 * time.Second)

			user3 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_2"}
			s.Nil(query.Query().Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "where_not_between_user").WhereNotBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 1)
			s.True(users[0].ID == user3.ID)

			var users1 []User
			s.Nil(query.Query().Where("name = ?", "where_not_between_user").WhereNotBetween("created_at", now.Copy().SubDay(), now).Find(&users1))
			s.True(len(users1) == 1)
			s.True(users1[0].ID == user3.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereBetween() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "or_where_between_user", Avatar: "or_where_between_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_between_user_1", Avatar: "or_where_between_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "or_where_between_user_2", Avatar: "or_where_between_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			now := carbon.Now().AddSecond()
			time.Sleep(2 * time.Second)

			user3 := User{Name: "or_where_between_user_3", Avatar: "or_where_between_avatar_3"}
			s.Nil(query.Query().Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "or_where_between_user_3").OrWhereBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 4)

			var users1 []User
			s.Nil(query.Query().Where("name = ?", "or_where_between_user_3").OrWhereBetween("created_at", now.Copy().SubDay(), now).Find(&users1))
			s.True(len(users1) == 4)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNotBetween() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "or_where_between_user", Avatar: "or_where_between_avatar"}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_between_user_1", Avatar: "or_where_between_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "or_where_between_user_2", Avatar: "or_where_between_avatar_2"}
			s.Nil(query.Query().Create(&user2))
			s.True(user2.ID > 0)

			now := carbon.Now().AddSecond()
			time.Sleep(2 * time.Second)

			user3 := User{Name: "or_where_between_user_3", Avatar: "or_where_between_avatar_3"}
			s.Nil(query.Query().Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "or_where_between_user_3").OrWhereNotBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 1)

			var users1 []User
			s.Nil(query.Query().Where("name = ?", "or_where_between_user_3").OrWhereNotBetween("created_at", now.Copy().SubDay(), now).Find(&users1))
			s.True(len(users1) == 1)
		})
	}
}

func (s *QueryTestSuite) TestWhereNull() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			bio := "where_null_bio"
			user := User{Name: "where_null_user", Avatar: "where_null_avatar", Bio: &bio}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_null_user", Avatar: "where_null_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "where_null_user").WhereNull("bio").Find(&users))
			s.True(len(users) == 1)
			s.True(users[0].ID == user1.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNull() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			bio := "or_where_null_bio"
			user := User{Name: "or_where_null_user", Avatar: "or_where_null_avatar", Bio: &bio}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_null_user_1", Avatar: "or_where_null_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "or_where_null_user").OrWhereNull("bio").Find(&users))
			s.True(len(users) >= 2)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotNull() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			bio := "where_not_null_bio"
			user := User{Name: "where_not_null_user", Avatar: "where_not_null_avatar", Bio: &bio}
			s.Nil(query.Query().Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_not_null_user", Avatar: "where_not_null_avatar_1"}
			s.Nil(query.Query().Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Query().Where("name = ?", "where_not_null_user").WhereNotNull("bio").Find(&users))
			s.True(len(users) == 1)
			s.True(users[0].ID == user.ID)
		})
	}
}

func (s *QueryTestSuite) TestWithoutEvents() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "event_save_without_name", Avatar: "without_events_avatar"}
					s.Nil(query.Query().WithoutEvents().Save(&user))
					s.True(user.ID > 0)
					s.Equal("without_events_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Query().Find(&user1, user.ID))
					s.Equal("event_save_without_name", user1.Name)
					s.Equal("without_events_avatar", user1.Avatar)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				test.setup()
			})
		}
	}
}

func (s *QueryTestSuite) TestWith() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "with_user", Address: &Address{
				Name: "with_address",
			}, Books: []*Book{{
				Name: "with_book0",
			}, {
				Name: "with_book1",
			}}}
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			tests := []struct {
				description string
				setup       func(description string)
			}{
				{
					description: "simple",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Query().With("Address").With("Books").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.True(user1.Address.ID > 0)
						s.True(user1.Books[0].ID > 0)
						s.True(user1.Books[1].ID > 0)
					},
				},
				{
					description: "with simple conditions",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Query().With("Books", "name = ?", "with_book0").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("with_book0", user1.Books[0].Name)
					},
				},
				{
					description: "with func conditions",
					setup: func(description string) {
						var user1 User
						s.Nil(query.Query().With("Books", func(query contractsorm.Query) contractsorm.Query {
							return query.Where("name = ?", "with_book0").Select("id", "user_id", "name")
						}).Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.Equal(1, len(user1.Books))
						s.Equal("with_book0", user1.Books[0].Name)
					},
				},
			}
			for _, test := range tests {
				test.setup(test.description)
			}
		})
	}
}

func (s *QueryTestSuite) TestWithNesting() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			user := User{Name: "with_nesting_user", Books: []*Book{{
				Name:   "with_nesting_book0",
				Author: &Author{Name: "with_nesting_author0"},
			}, {
				Name:   "with_nesting_book1",
				Author: &Author{Name: "with_nesting_author1"},
			}}}
			s.Nil(query.Query().Select(gorm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[0].Author.ID > 0)
			s.True(user.Books[1].ID > 0)
			s.True(user.Books[1].Author.ID > 0)

			var user1 User
			s.Nil(query.Query().With("Books.Author").Find(&user1, user.ID))
			s.True(user1.ID > 0)
			s.Equal("with_nesting_user", user1.Name)
			s.True(user1.Books[0].ID > 0)
			s.Equal("with_nesting_book0", user1.Books[0].Name)
			s.True(user1.Books[0].Author.ID > 0)
			s.Equal("with_nesting_author0", user1.Books[0].Author.Name)
			s.True(user1.Books[1].ID > 0)
			s.Equal("with_nesting_book1", user1.Books[1].Name)
			s.True(user1.Books[1].Author.ID > 0)
			s.Equal("with_nesting_author1", user1.Books[1].Author.Name)
		})
	}
}

func (s *QueryTestSuite) TestJsonWhereClauses() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			data := []JsonData{
				{
					Data: `{"string":"first","int":123,"float":123.456,"bool":true,"array":["abc","def","ghi"],"nested":{"string":"first","int":456},"objects":[{"level":"first","value":"abc"},{"level":"second","value":"def"}]}`,
				},
				{
					Data: `{"string":"second","int":123,"float":789.123,"bool":false,"array":["jkl","def","abc"]}`,
				},
			}
			s.Nil(query.Query().Create(&data))

			tests := []struct {
				name   string
				find   func(any, ...any) error
				assert func([]JsonData)
			}{
				{
					name: "string key",
					find: query.Query().Where("data->string", "first").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "int key",
					find: query.Query().Where("data->int", 123).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "float key(multiple values)",
					find: query.Query().WhereIn("data->float", []any{123.456, 789.123}).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "bool key(pointer)",
					find: query.Query().Where("data->bool", convert.Pointer(false)).Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[1].Data, items[0].Data)
					},
				},
				{
					name: "nested key",
					find: query.Query().Where("data->nested->int", 456).Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "nested key with array",
					find: query.Query().Where("data->objects[0]->level", "first").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "key exists",
					find: query.Query().WhereJsonContainsKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[0].Data, items[0].Data)
					},
				},
				{
					name: "key does not exist",
					find: query.Query().WhereJsonDoesntContainKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 1)
						s.JSONEq(data[1].Data, items[0].Data)
					},
				},
				{
					name: "array contains",
					find: query.Query().WhereJsonContains("data->array", "abc").Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "array does not contain",
					find: query.Query().WhereJsonDoesntContain("data->array", "abc").Find,
					assert: func(items []JsonData) {
						s.Len(items, 0)
					},
				},
				{
					name: "array contains multiple values",
					find: query.Query().WhereJsonContains("data->array", []string{"abc", "def"}).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "array length",
					find: query.Query().WhereJsonLength("data->array", 2).Find,
					assert: func(items []JsonData) {
						s.Len(items, 0)
					},
				},
				{
					name: "array length greater than",
					find: query.Query().WhereJsonLength("data->array > ?", 2).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "string or float key",
					find: query.Query().Where("data->string", "first").OrWhere("data->float", 789.123).Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
				{
					name: "contains or key does not exist",
					find: query.Query().WhereJsonContains("data->array", "ghi").OrWhereJsonDoesntContainKey("data->nested->string").Find,
					assert: func(items []JsonData) {
						s.Len(items, 2)
						s.JSONEq(data[0].Data, items[0].Data)
						s.JSONEq(data[1].Data, items[1].Data)
					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var items []JsonData
					s.NoError(tt.find(&items))
					tt.assert(items)
				})
			}
		})
	}
}

func (s *QueryTestSuite) TestJsonColumnsUpdate() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			data := []JsonData{
				{
					Data: `{"string":"first","int":123,"float":123.456,"bool":true,"array":["abc","def","ghi"],"nested":{"string":"first","int":456},"objects":[{"level":"first","value":"abc"},{"level":"second","value":"def"}]}`,
				},
			}
			s.NoError(query.Query().Create(&data))

			tests := []struct {
				name   string
				update map[string]any
				assert func(before JsonData, after JsonData)
			}{
				{
					name:   "update string",
					update: map[string]any{"data->string": "updated_first"},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "updated_first")
						s.Contains(after.Data, "updated_first")
					},
				},
				{
					name:   "update int",
					update: map[string]any{"data->int": 789},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "789")
						s.Contains(after.Data, "789")
					},
				},
				{
					name:   "update float(pointer)",
					update: map[string]any{"data->float": convert.Pointer(456.789)},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "456.789")
						s.Contains(after.Data, "456.789")
					},
				},
				{
					name:   "update array",
					update: map[string]any{"data->array": []string{"uvw", "xyz"}},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "uvw")
						s.Contains(after.Data, "uvw")

						s.NotContains(before.Data, "xyz")
						s.Contains(after.Data, "xyz")
					},
				},
				{
					name: "update multiple keys",
					update: map[string]any{
						"data->bool":              false,
						"data->objects[0]->level": "first_changed",
						"data->nested->string":    "updated_nested_string",
					},
					assert: func(before JsonData, after JsonData) {
						s.NotContains(before.Data, "false")
						s.Contains(after.Data, "false")

						s.NotContains(before.Data, "first_changed")
						s.Contains(after.Data, "first_changed")

						s.NotContains(before.Data, "updated_nested_string")
						s.Contains(after.Data, "updated_nested_string")

					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var before, after JsonData
					s.NoError(query.Query().First(&before))
					res, err := query.Query().Model(&before).Update(tt.update)
					s.NoError(err)
					s.Equal(int64(1), res.RowsAffected)
					s.NoError(query.Query().Where("id", before.ID).First(&after))
					s.NotEqual(before.Data, after.Data)
					tt.assert(before, after)
				})
			}
		})
	}
}

func TestCustomConnection(t *testing.T) {
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.CreateTable()

	sqliteTestQuery := NewTestQueryBuilder().Sqlite("", false)
	sqliteTestQuery.CreateTable()

	query := postgresTestQuery.Query()

	review := Review{Body: "create_review"}
	assert.Nil(t, query.Create(&review))
	assert.True(t, review.ID > 0)

	var review1 Review
	assert.Nil(t, query.Where("body", "create_review").First(&review1))
	assert.True(t, review1.ID > 0)

	config := sqliteTestQuery.Driver().Pool().Writers[0]
	config.Connection = "sqlite"
	mockDatabaseConfig(postgresTestQuery.MockConfig(), config)

	product := Product{Name: "create_product"}
	assert.Nil(t, query.Create(&product))
	assert.True(t, product.ID > 0)

	var product1 Product
	assert.Nil(t, query.Where("name", "create_product").First(&product1))
	assert.True(t, product1.ID > 0)

	var product2 Product
	assert.Nil(t, query.Where("name", "create_product1").First(&product2))
	assert.True(t, product2.ID == 0)

	config = postgresTestQuery.Driver().Pool().Writers[0]
	config.Connection = "dummy"
	mockDatabaseConfig(postgresTestQuery.MockConfig(), config)

	person := Person{Name: "create_person"}
	assert.NotNil(t, query.Create(&person))
	assert.True(t, person.ID == 0)

	docker, err := sqliteTestQuery.Driver().Docker()
	assert.NoError(t, err)
	assert.NoError(t, docker.Shutdown())
}

func TestOrmReadWriteSeparate(t *testing.T) {
	dbs := NewTestQueryBuilder().AllWithReadWrite()

	for drive, db := range dbs {
		t.Run(drive, func(t *testing.T) {
			db["read"].CreateTable(TestTableUsers)
			db["write"].CreateTable(TestTableUsers)

			user1 := User{Name: "user"}
			assert.Nil(t, db["mix"].Query().Create(&user1))
			assert.True(t, user1.ID > 0)

			var user2 User
			assert.Nil(t, db["mix"].Query().Find(&user2, user1.ID))
			assert.True(t, user2.ID == 0)

			var user3 User
			assert.Nil(t, db["read"].Query().Find(&user3, user1.ID))
			assert.True(t, user3.ID == 0)

			var user4 User
			assert.Nil(t, db["write"].Query().Find(&user4, user1.ID))
			assert.True(t, user4.ID > 0)
		})
	}

	docker, err := dbs[sqlite.Name]["read"].Driver().Docker()
	assert.NoError(t, err)
	assert.NoError(t, docker.Shutdown())

	docker, err = dbs[sqlite.Name]["write"].Driver().Docker()
	assert.NoError(t, err)
	assert.NoError(t, docker.Shutdown())
}

func TestTablePrefixAndSingular(t *testing.T) {
	queries := NewTestQueryBuilder().All("goravel_", true)

	for drive, query := range queries {
		t.Run(drive, func(t *testing.T) {
			query.CreateTable(TestTableUser)

			user := User{Name: "user"}
			assert.Nil(t, query.Query().Create(&user))
			assert.True(t, user.ID > 0)

			var user1 User
			assert.Nil(t, query.Query().Find(&user1, user.ID))
			assert.True(t, user1.ID > 0)
		})
	}

	if queries[sqlite.Name] != nil {
		docker, err := queries[sqlite.Name].Driver().Docker()
		assert.NoError(t, err)
		assert.NoError(t, docker.Shutdown())
	}
}

func TestPostgresWithSchema(t *testing.T) {
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.WithSchema(testSchema)
	postgresTestQuery.CreateTable(TestTableUsers)

	user := User{Name: "first_user"}
	assert.Nil(t, postgresTestQuery.Query().Create(&user))
	assert.True(t, user.ID > 0)

	var user1 User
	assert.Nil(t, postgresTestQuery.Query().Where("name", "first_user").First(&user1))
	assert.True(t, user1.ID > 0)
}

func TestSqlserverWithSchema(t *testing.T) {
	sqlserverTestQuery := NewTestQueryBuilder().Sqlserver("", false)
	sqlserverTestQuery.WithSchema(testSchema)
	sqlserverTestQuery.CreateTable(TestTableSchema)

	schema := Schema{Name: "first_schema"}
	assert.Nil(t, sqlserverTestQuery.Query().Create(&schema))
	assert.True(t, schema.ID > 0)

	var schema1 Schema
	assert.Nil(t, sqlserverTestQuery.Query().Where("name", "first_schema").First(&schema1))
	assert.True(t, schema1.ID > 0)
}

// https://github.com/goravel/goravel/issues/706
func TestTimezone(t *testing.T) {
	queries := NewTestQueryBuilder().AllWithTimezone("Asia/Shanghai")

	defer func() {
		if queries[sqlite.Name] != nil {
			docker, err := queries[sqlite.Name].Driver().Docker()
			assert.NoError(t, err)
			assert.NoError(t, docker.Shutdown())
		}
	}()

	for driver, query := range queries {
		t.Run(driver, func(t *testing.T) {
			query.CreateTable()

			user := User{Name: "count_user", Avatar: "count_avatar"}
			assert.Nil(t, query.Query().Create(&user))
			assert.True(t, user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			assert.Nil(t, query.Query().Create(&user1))
			assert.True(t, user1.ID > 0)

			count, err := query.Query().Model(&User{}).Where("name = ?", "count_user").Count()
			assert.Nil(t, err)
			assert.True(t, count > 0)

			count, err = query.Query().Table("users").Where("name = ?", "count_user").Count()
			assert.Nil(t, err)
			assert.True(t, count > 0)
		})
	}
}

func paginator(page string, limit string) func(methods contractsorm.Query) contractsorm.Query {
	return func(query contractsorm.Query) contractsorm.Query {
		page, _ := strconv.Atoi(page)
		limit, _ := strconv.Atoi(limit)
		offset := (page - 1) * limit

		return query.Offset(offset).Limit(limit)
	}
}

func (s *QueryTestSuite) TestUuidColumn() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestUuidColumn_%s", driver), func() {
			id, _ := uuid.NewV7()
			// Test UUID column creation and operations
			entity := UuidEntity{
				Uuid: id.String(),
				Name: "test_uuid_entity",
			}

			err := query.Query().Create(&entity)
			s.NoError(err)
			s.NotEmpty(entity.Uuid)

			// Test finding by UUID
			var foundEntity UuidEntity
			err = query.Query().Where("uuid", entity.Uuid).First(&foundEntity)
			s.NoError(err)
			s.Equal("test_uuid_entity", foundEntity.Name)

			// Test UUID format (basic validation)
			s.Len(entity.Uuid, 36) // Standard UUID length with hyphens
			s.Contains(entity.Uuid, "-")
		})
	}
}

func (s *QueryTestSuite) TestUlidColumn() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestUlidColumn_%s", driver), func() {
			// Test ULID column creation and operations
			entity := UlidEntity{
				ID:   "01AN4Z07BY79KA1307SR9X4MV3", // Valid ULID
				Name: "test_ulid_entity",
			}

			err := query.Query().Create(&entity)
			s.NoError(err)
			s.NotEmpty(entity.ID)

			// Test finding by ULID
			var foundEntity UlidEntity
			err = query.Query().Where("id", entity.ID).First(&foundEntity)
			s.NoError(err)
			s.Equal("test_ulid_entity", foundEntity.Name)
			s.Equal(entity.ID, foundEntity.ID)

			// Test ULID format (basic validation)
			s.Len(entity.ID, 26) // Standard ULID length
		})
	}
}

func (s *QueryTestSuite) TestMorphableRelationships() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestMorphableRelationships_%s", driver), func() {
			user := User{
				Name: "test_user",
			}

			err := query.Query().Create(&user)
			s.NoError(err)

			entity := House{
				Name:          "test_morph_house",
				HouseableID:   user.ID,
				HouseableType: "users",
			}

			err = query.Query().Create(&entity)
			s.NoError(err)

			// Test finding by morph type
			var foundEntity House
			err = query.Query().Where("houseable_type", "users").First(&foundEntity)
			s.NoError(err)
			s.Equal("test_morph_house", foundEntity.Name)
			s.Equal(uint(1), foundEntity.HouseableID)
			s.Equal("users", foundEntity.HouseableType)

			// Test finding by morph type
			var userWithHouse User
			err = query.Query().Where("id", user.ID).With("House").First(&userWithHouse)
			s.NoError(err)

			s.Equal("test_user", userWithHouse.Name)
			s.NotNil(userWithHouse.House)
			s.Equal("test_morph_house", userWithHouse.House.Name)
			s.Equal(uint(1), userWithHouse.House.HouseableID)
		})
	}
}

func (s *QueryTestSuite) TestUuidMorphableRelationships() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestUuidMorphableRelationships_%s", driver), func() {
			// Test UUID morph relationships
			entity := UuidMorphableEntity{
				Name:          "test_uuid_morph_entity",
				MorphableID:   "550e8400-e29b-41d4-a716-446655440000",
				MorphableType: "User",
			}

			err := query.Query().Create(&entity)
			s.NoError(err)
			s.True(entity.ID > 0)

			// Test finding by UUID morph
			var foundEntity UuidMorphableEntity
			err = query.Query().Where("morphable_id", "550e8400-e29b-41d4-a716-446655440000").First(&foundEntity)
			s.NoError(err)
			s.Equal("test_uuid_morph_entity", foundEntity.Name)
			s.Equal("User", foundEntity.MorphableType)

			// SQL Server stores UUIDs as binary data, so we need to handle this differently
			if driver == "SQL Server" {
				// For SQL Server, we need to verify the UUID is stored correctly
				// but the format may be different (binary vs string)
				s.NotEmpty(foundEntity.MorphableID)
				// We can't do exact string comparison for SQL Server UUID format
			} else {
				// For other databases, we can do string comparison
				s.Equal("550e8400-e29b-41d4-a716-446655440000", foundEntity.MorphableID)

				// Test UUID format validation
				s.Len(foundEntity.MorphableID, 36)
				s.Contains(foundEntity.MorphableID, "-")
			}
		})
	}
}

func (s *QueryTestSuite) TestUlidMorphableRelationships() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestUlidMorphableRelationships_%s", driver), func() {
			// Test ULID morph relationships
			entity := UlidMorphableEntity{
				Name:          "test_ulid_morph_entity",
				MorphableID:   "01AN4Z07BY79KA1307SR9X4MV3",
				MorphableType: "User",
			}

			err := query.Query().Create(&entity)
			s.NoError(err)
			s.True(entity.ID > 0)

			// Test finding by ULID morph
			var foundEntity UlidMorphableEntity
			err = query.Query().Where("morphable_id", "01AN4Z07BY79KA1307SR9X4MV3").First(&foundEntity)
			s.NoError(err)
			s.Equal("test_ulid_morph_entity", foundEntity.Name)
			s.Equal("01AN4Z07BY79KA1307SR9X4MV3", foundEntity.MorphableID)
			s.Equal("User", foundEntity.MorphableType)

			// Test ULID format validation
			s.Len(foundEntity.MorphableID, 26)
		})
	}
}

func (s *QueryTestSuite) TestMorphablePolymorphicQueries() {
	for driver, query := range s.queries {
		s.Run(fmt.Sprintf("TestMorphablePolymorphicQueries_%s", driver), func() {
			// Create morph entities for different types
			entities := []MorphableEntity{
				{Name: "user_morph_1", MorphableID: 1, MorphableType: "User"},
				{Name: "user_morph_2", MorphableID: 2, MorphableType: "User"},
				{Name: "post_morph_1", MorphableID: 1, MorphableType: "Post"},
				{Name: "post_morph_2", MorphableID: 2, MorphableType: "Post"},
			}

			for _, entity := range entities {
				err := query.Query().Create(&entity)
				s.NoError(err)
			}

			// Test querying by morph type
			var userMorphs []MorphableEntity
			err := query.Query().Where("morphable_type", "User").Find(&userMorphs)
			s.NoError(err)
			s.Len(userMorphs, 2)

			var postMorphs []MorphableEntity
			err = query.Query().Where("morphable_type", "Post").Find(&postMorphs)
			s.NoError(err)
			s.Len(postMorphs, 2)

			// Test querying by morph type and ID
			var specificMorph MorphableEntity
			err = query.Query().Where("morphable_type", "User").Where("morphable_id", 1).First(&specificMorph)
			s.NoError(err)
			s.Equal("user_morph_1", specificMorph.Name)

			// Test combined queries
			var combinedMorphs []MorphableEntity
			err = query.Query().Where("morphable_id", 1).Find(&combinedMorphs)
			s.NoError(err)
			s.Len(combinedMorphs, 2) // Should find both User and Post with ID 1
		})
	}
}

func (s *QueryTestSuite) TestWhereAny() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_any_user1", Avatar: "where_any_avatar1", Bio: convert.Pointer("bio1")},
				{Name: "where_any_user2", Avatar: "where_any_avatar2", Bio: convert.Pointer("bio2")},
				{Name: "where_any_user3", Avatar: "where_any_avatar3", Bio: convert.Pointer("bio3")},
				{Name: "where_any_user4", Avatar: "where_any_avatar4", Bio: convert.Pointer("bio4")},
			}
			s.Nil(query.Query().Create(&users))

			tests := []struct {
				name   string
				find   func(any, ...any) error
				assert func([]User)
			}{
				{
					name: "equals operator with single match",
					find: query.Query().WhereAny([]string{"name", "avatar"}, "=", "where_any_user1").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_any_user1", items[0].Name)
					},
				},
				{
					name: "equals operator with multiple matches",
					find: query.Query().WhereAny([]string{"name", "avatar"}, "=", "where_any_avatar2").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_any_user2", items[0].Name)
					},
				},
				{
					name: "combined with Where clause - simple",
					find: query.Query().Where("name", "where_any_user2").WhereAny([]string{"avatar"}, "=", "where_any_avatar2").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_any_user2", items[0].Name)
					},
				},
				{
					name: "Where before and after WhereAny",
					find: query.Query().Where("name LIKE ?", "where_any%").WhereAny([]string{"avatar"}, "=", "where_any_avatar1").Where("bio IS NOT NULL").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_any_user1", items[0].Name)
					},
				},
				{
					name: "no matches",
					find: query.Query().WhereAny([]string{"name", "avatar"}, "=", "nonexistent").Find,
					assert: func(items []User) {
						s.Len(items, 0)
					},
				},
				{
					name: "multiple WhereAny calls",
					find: query.Query().WhereAny([]string{"name"}, "IN", []string{"where_any_user1", "where_any_user2"}).WhereAny([]string{"avatar"}, "IN", []string{"where_any_avatar1", "where_any_avatar2"}).Find,
					assert: func(items []User) {
						s.Len(items, 2)
					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var items []User
					s.Nil(tt.find(&items))
					tt.assert(items)
				})
			}
		})
	}
}

func (s *QueryTestSuite) TestWhereAll() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_all_user1", Avatar: "where_all_avatar1", Bio: convert.Pointer("bio1")},
				{Name: "where_all_user1", Avatar: "where_all_avatar2", Bio: convert.Pointer("bio2")},
				{Name: "where_all_user2", Avatar: "where_all_avatar1", Bio: convert.Pointer("bio3")},
				{Name: "where_all_user2", Avatar: "where_all_avatar2", Bio: convert.Pointer("bio4")},
			}
			s.Nil(query.Query().Create(&users))

			tests := []struct {
				name   string
				find   func(any, ...any) error
				assert func([]User)
			}{
				{
					name: "equals operator - all columns match",
					find: query.Query().WhereAll([]string{"name", "avatar"}, "=", "where_all_user1").Find,
					assert: func(items []User) {
						s.Len(items, 0)
					},
				},
				{
					name: "single column match",
					find: query.Query().WhereAll([]string{"name"}, "=", "where_all_user1").Find,
					assert: func(items []User) {
						s.Len(items, 2)
						s.Equal("where_all_user1", items[0].Name)
						s.Equal("where_all_user1", items[1].Name)
					},
				},
				{
					name: "combined with Where clause - simple",
					find: query.Query().Where("name", "where_all_user2").WhereAll([]string{"avatar"}, "=", "where_all_avatar1").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_all_user2", items[0].Name)
						s.Equal("where_all_avatar1", items[0].Avatar)
					},
				},
				{
					name: "Where before and after WhereAll",
					find: query.Query().Where("name LIKE ?", "where_all%").WhereAll([]string{"avatar"}, "=", "where_all_avatar1").Where("bio IS NOT NULL").Find,
					assert: func(items []User) {
						s.Len(items, 2)
					},
				},
				{
					name: "no matches",
					find: query.Query().WhereAll([]string{"name", "avatar"}, "=", "nonexistent").Find,
					assert: func(items []User) {
						s.Len(items, 0)
					},
				},
				{
					name: "multiple WhereAll calls",
					find: query.Query().WhereAll([]string{"name"}, "=", "where_all_user1").WhereAll([]string{"avatar"}, "=", "where_all_avatar1").Find,
					assert: func(items []User) {
						s.Len(items, 1)
						s.Equal("where_all_user1", items[0].Name)
						s.Equal("where_all_avatar1", items[0].Avatar)
					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var items []User
					s.Nil(tt.find(&items))
					tt.assert(items)
				})
			}
		})
	}
}

func (s *QueryTestSuite) TestWhereNone() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			users := []User{
				{Name: "where_none_user1", Avatar: "where_none_avatar1", Bio: convert.Pointer("bio1")},
				{Name: "where_none_user2", Avatar: "where_none_avatar2", Bio: convert.Pointer("bio2")},
				{Name: "where_none_user3", Avatar: "where_none_avatar3", Bio: convert.Pointer("bio3")},
				{Name: "where_none_user4", Avatar: "where_none_avatar4", Bio: convert.Pointer("bio4")},
			}
			s.Nil(query.Query().Create(&users))

			tests := []struct {
				name   string
				find   func(any, ...any) error
				assert func([]User)
			}{
				{
					name: "equals operator - exclude single value",
					find: query.Query().WhereNone([]string{"name"}, "=", "where_none_user1").Find,
					assert: func(items []User) {
						s.Len(items, 3)
						s.Equal("where_none_user2", items[0].Name)
						s.Equal("where_none_user3", items[1].Name)
						s.Equal("where_none_user4", items[2].Name)
					},
				},
				{
					name: "equals operator - exclude from multiple columns",
					find: query.Query().WhereNone([]string{"name", "avatar"}, "=", "where_none_user1").Find,
					assert: func(items []User) {
						s.Len(items, 3)
					},
				},
				{
					name: "combined with Where clause - simple",
					find: query.Query().Where("name LIKE ?", "where_none%").WhereNone([]string{"avatar"}, "=", "where_none_avatar1").Find,
					assert: func(items []User) {
						s.Len(items, 3)
						s.Equal("where_none_user2", items[0].Name)
						s.Equal("where_none_user3", items[1].Name)
						s.Equal("where_none_user4", items[2].Name)
					},
				},
				{
					name: "Where before and after WhereNone",
					find: query.Query().Where("name LIKE ?", "where_none%").WhereNone([]string{"avatar"}, "=", "where_none_avatar1").Where("bio IS NOT NULL").Find,
					assert: func(items []User) {
						s.Len(items, 3)
					},
				},
				{
					name: "no matches - all excluded",
					find: query.Query().WhereNone([]string{"name"}, "LIKE", "where_none%").Find,
					assert: func(items []User) {
						s.Len(items, 0)
					},
				},
				{
					name: "all records match when excluding non-existent value",
					find: query.Query().WhereNone([]string{"name", "avatar"}, "=", "nonexistent").Find,
					assert: func(items []User) {
						s.Len(items, 4)
					},
				},
				{
					name: "multiple WhereNone calls",
					find: query.Query().WhereNone([]string{"name"}, "=", "where_none_user1").WhereNone([]string{"avatar"}, "=", "where_none_avatar4").Find,
					assert: func(items []User) {
						s.Len(items, 2)
						s.Equal("where_none_user2", items[0].Name)
						s.Equal("where_none_user3", items[1].Name)
					},
				},
			}

			for _, tt := range tests {
				s.Run(tt.name, func() {
					var items []User
					s.Nil(tt.find(&items))
					tt.assert(items)
				})
			}
		})
	}
}

func Benchmark_Orm(b *testing.B) {
	query := NewTestQueryBuilder().Postgres("", false)
	query.CreateTable(TestTableAuthors)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		author := Author{
			Name:   "benchmark",
			BookID: 1,
		}
		err := query.Query().Create(&author)
		if err != nil {
			b.Error(err)
		}

		var authors []Author
		err = query.Query().Limit(50).Find(&authors)
		if err != nil {
			b.Error(err)
		}
	}
}
