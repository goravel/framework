package gorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "gorm.io/driver/postgres"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/orm"
	configmocks "github.com/goravel/framework/mocks/config"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type QueryTestSuite struct {
	suite.Suite
	queries          map[ormcontract.Driver]ormcontract.Query
	mysqlDocker      *MysqlDocker
	mysqlDocker1     *MysqlDocker
	postgresqlDocker *PostgresqlDocker
	sqliteDocker     *SqliteDocker
	sqlserverDocker  *SqlserverDocker
}

func TestQueryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	testContext = context.Background()
	testContext = context.WithValue(testContext, testContextKey, "goravel")

	mysqlDocker := NewMysqlDocker(testDatabaseDocker)
	mysqlQuery, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	mysqlDocker1 := NewMysql1Docker(testDatabaseDocker)
	_, err = mysqlDocker1.New()
	if err != nil {
		log.Fatalf("Init mysql1 error: %s", err)
	}

	postgresqlDocker := NewPostgresqlDocker(testDatabaseDocker)
	postgresqlQuery, err := postgresqlDocker.New()
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	sqliteDocker := NewSqliteDocker(dbDatabase)
	sqliteQuery, err := sqliteDocker.New()
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverDocker := NewSqlserverDocker(testDatabaseDocker)
	sqlserverQuery, err := sqlserverDocker.New()
	if err != nil {
		log.Fatalf("Init sqlserver error: %s", err)
	}

	suite.Run(t, &QueryTestSuite{
		queries: map[ormcontract.Driver]ormcontract.Query{
			ormcontract.DriverMysql:      mysqlQuery,
			ormcontract.DriverPostgresql: postgresqlQuery,
			ormcontract.DriverSqlite:     sqliteQuery,
			ormcontract.DriverSqlserver:  sqlserverQuery,
		},
		mysqlDocker:      mysqlDocker,
		mysqlDocker1:     mysqlDocker1,
		postgresqlDocker: postgresqlDocker,
		sqliteDocker:     sqliteDocker,
		sqlserverDocker:  sqlserverDocker,
	})
}

func (s *QueryTestSuite) SetupTest() {}

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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)

					var userAddress Address
					s.Nil(query.Model(&user1).Association("Address").Find(&userAddress))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID), driver)
					s.True(user1.ID > 0, driver)
					s.Nil(query.Model(&user1).Association("Address").Append(&Address{Name: "association_has_one_append_address1"}), driver)

					s.Nil(query.Load(&user1, "Address"), driver)
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Model(&user1).Association("Books").Append(&Book{Name: "association_has_many_append_address3"}))

					s.Nil(query.Load(&user1, "Books"))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Model(&user1).Association("Address").Replace(&Address{Name: "association_has_one_append_address1"}))

					s.Nil(query.Load(&user1, "Address"))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Model(&user1).Association("Books").Replace(&Book{Name: "association_has_many_replace_address3"}))

					s.Nil(query.Load(&user1, "Books"))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					// No ID when Delete
					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Model(&user1).Association("Address").Delete(&Address{Name: "association_delete_address"}))

					s.Nil(query.Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.Equal("association_delete_address", user1.Address.Name)

					// Has ID when Delete
					var user2 User
					s.Nil(query.Find(&user2, user.ID))
					s.True(user2.ID > 0)
					var userAddress Address
					userAddress.ID = user1.Address.ID
					s.Nil(query.Model(&user2).Association("Address").Delete(&userAddress))

					s.Nil(query.Load(&user2, "Address"))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)

					// No ID when Delete
					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(query.Model(&user1).Association("Address").Clear())

					s.Nil(query.Load(&user1, "Address"))
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

					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Equal(int64(2), query.Model(&user1).Association("Books").Count())
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
		s.Run(driver.String(), func() {
			user := &User{
				Name: "belongs_to_name",
				Address: &Address{
					Name: "belongs_to_address",
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var userAddress Address
			s.Nil(query.With("User").Where("name = ?", "belongs_to_address").First(&userAddress))
			s.True(userAddress.ID > 0)
			s.True(userAddress.User.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestCount() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var count int64
			s.Nil(query.Model(&User{}).Where("name = ?", "count_user").Count(&count))
			s.True(count > 0)

			var count1 int64
			s.Nil(query.Table("users").Where("name = ?", "count_user").Count(&count1))
			s.True(count1 > 0)
		})
	}
}

func (s *QueryTestSuite) TestCreate() {
	for driver, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success when refresh connection",
				setup: func() {
					s.mockDummyConnection(driver)

					people := People{Body: "create_people"}
					s.Nil(query.Create(&people))
					s.True(people.ID > 0)

					people1 := People{Body: "create_people1"}
					s.Nil(query.Model(&People{}).Create(&people1))
					s.True(people1.ID > 0)
				},
			},
			{
				name: "success when create with no relationships",
				setup: func() {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "success when create with select orm.Associations",
				setup: func() {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Select(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "success when create with select fields",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Select("Name", "Avatar", "Address").Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID > 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "success when create with omit fields",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Omit("Address").Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID > 0)
					s.True(user.Books[1].ID > 0)
				},
			},
			{
				name: "success create with omit orm.Associations",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(query.Omit(orm.Associations).Create(&user))
					s.True(user.ID > 0)
					s.True(user.Address.ID == 0)
					s.True(user.Books[0].ID == 0)
					s.True(user.Books[1].ID == 0)
				},
			},
			{
				name: "error when set select and omit at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Omit(orm.Associations).Select("Name").Create(&user), "cannot set Select and Omits at the same time")
				},
			},
			{
				name: "error when select that set fields and orm.Associations at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Select("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time")
				},
			},
			{
				name: "error when omit that set fields and orm.Associations at the same time",
				setup: func() {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(query.Omit("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time")
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
		s.Run(driver.String(), func() {
			user := User{Name: "cursor_user", Avatar: "cursor_avatar", Address: &Address{Name: "cursor_address"}, Books: []*Book{
				{Name: "cursor_book"},
			}}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "cursor_user", Avatar: "cursor_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "cursor_user", Avatar: "cursor_avatar2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)
			res, err := query.Delete(&user2)
			s.Nil(err)
			s.Equal(int64(1), res.RowsAffected)

			users, err := query.Model(&User{}).Where("name = ?", "cursor_user").WithTrashed().With("Address").With("Books").Cursor()
			s.Nil(err)
			var size int
			var addressNum int
			var bookNum int
			for row := range users {
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
		})
	}
}

func (s *QueryTestSuite) TestDBRaw() {
	userName := "db_raw"
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: userName}

			s.Nil(query.Create(&user))
			s.True(user.ID > 0)
			switch driver {
			case ormcontract.DriverSqlserver, ormcontract.DriverMysql:
				res, err := query.Model(&user).Update("Name", databasedb.Raw("concat(name, ?)", driver.String()))
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)
			default:
				res, err := query.Model(&user).Update("Name", databasedb.Raw("name || ?", driver.String()))
				s.Nil(err)
				s.Equal(int64(1), res.RowsAffected)
			}

			var user1 User
			s.Nil(query.Find(&user1, user.ID))
			s.True(user1.ID > 0)
			s.True(user1.Name == userName+driver.String())
		})
	}
}

func (s *QueryTestSuite) TestDelete() {
	for driver, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Delete(&user)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success when refresh connection",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Delete(&user)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)

					// refresh connection
					s.mockDummyConnection(driver)

					people := People{Body: "delete_people"}
					s.Nil(query.Create(&people))
					s.True(people.ID > 0)

					res, err = query.Delete(&people)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var people1 People
					s.Nil(query.Find(&people1, people.ID))
					s.Equal(uint(0), people1.ID)
				},
			},
			{
				name: "success by id",
				setup: func() {
					user := User{Name: "delete_user", Avatar: "delete_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Delete(&User{}, user.ID)
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal(uint(0), user1.ID)
				},
			},
			{
				name: "success by multiple",
				setup: func() {
					users := []User{{Name: "delete_user", Avatar: "delete_avatar"}, {Name: "delete_user1", Avatar: "delete_avatar1"}}
					s.Nil(query.Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Delete(&User{}, []uint{users[0].ID, users[1].ID})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var count int64
					s.Nil(query.Model(&User{}).Where("name", "delete_user").OrWhere("name", "delete_user1").Count(&count))
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
		s.Run(driver.String(), func() {
			user := User{Name: "distinct_user", Avatar: "distinct_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "distinct_user", Avatar: "distinct_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Distinct("name").Find(&users, []uint{user.ID, user1.ID}))
			s.Equal(1, len(users))
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
				name: "trigger when create",
				setup: func() {
					user := User{Name: "event_creating_name"}
					s.Nil(query.Create(&user))
					s.Equal("event_creating_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_creating_name", user1.Name)
					s.Equal("event_creating_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.FirstOrCreate(&user, User{Name: "event_creating_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_creating_FirstOrCreate_name", user.Name)
					s.Equal("event_creating_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_creating_FirstOrCreate_name", user1.Name)
					s.Equal("event_creating_FirstOrCreate_avatar", user1.Avatar)
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
				name: "trigger when create",
				setup: func() {
					user := User{Name: "event_created_name", Avatar: "avatar"}
					s.Nil(query.Create(&user))
					s.Equal(fmt.Sprintf("event_created_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_created_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.FirstOrCreate(&user, User{Name: "event_created_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_created_FirstOrCreate_name", user.Name)
					s.Equal(fmt.Sprintf("event_created_FirstOrCreate_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_created_FirstOrCreate_name", user1.Name)
					s.Equal("", user1.Avatar)
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
				name: "trigger when create",
				setup: func() {
					user := User{Name: "event_saving_create_name"}
					s.Nil(query.Create(&user))
					s.Equal("event_saving_create_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saving_create_name", user1.Name)
					s.Equal("event_saving_create_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.FirstOrCreate(&user, User{Name: "event_saving_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_saving_FirstOrCreate_name", user.Name)
					s.Equal("event_saving_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saving_FirstOrCreate_name", user1.Name)
					s.Equal("event_saving_FirstOrCreate_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_saving_save_name"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_save_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saving_save_name", user1.Name)
					s.Equal("event_saving_save_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by single column",
				setup: func() {
					user := User{Name: "event_saving_single_update_name", Avatar: "avatar"}
					s.Nil(query.Create(&user))

					res, err := query.Model(&user).Update("avatar", "event_saving_single_update_avatar")
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_saving_single_update_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saving_single_update_name", user1.Name)
					s.Equal("event_saving_single_update_avatar1", user1.Avatar)
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
					s.Nil(query.Create(&user))
					s.Equal("event_saved_create_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saved_create_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user User
					s.Nil(query.FirstOrCreate(&user, User{Name: "event_saved_FirstOrCreate_name"}))
					s.True(user.ID > 0)
					s.Equal("event_saved_FirstOrCreate_name", user.Name)
					s.Equal("event_saved_FirstOrCreate_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saved_FirstOrCreate_name", user1.Name)
					s.Equal("", user1.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_saved_save_name", Avatar: "avatar"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saved_save_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saved_save_name", user1.Name)
					s.Equal("avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by map",
				setup: func() {
					user := User{Name: "event_saved_map_update_name", Avatar: "avatar"}
					s.Nil(query.Create(&user))

					res, err := query.Model(&user).Update(map[string]any{
						"avatar": "event_saved_map_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_saved_map_update_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_saved_map_update_name", user1.Name)
					s.Equal("event_saved_map_update_avatar", user1.Avatar)
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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "not trigger when create by save",
				setup: func() {
					user := User{Name: "event_updating_save_name", Avatar: "avatar"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_updating_save_name", Avatar: "avatar"}
					s.Nil(query.Save(&user))

					user.Avatar = "event_updating_save_avatar"
					s.Nil(query.Save(&user))
					s.Equal("event_updating_save_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_updating_save_name", user1.Name)
					s.Equal("event_updating_save_avatar1", user1.Avatar)
				},
			},
			{
				name: "trigger when update by model",
				setup: func() {
					user := User{Name: "event_updating_model_update_name", Avatar: "avatar"}
					s.Nil(query.Create(&user))

					res, err := query.Model(&user).Update(User{
						Avatar: "event_updating_model_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal(fmt.Sprintf("event_updating_model_update_avatar_%d", user.ID), user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "not trigger when create by save",
				setup: func() {
					user := User{Name: "event_updated_save_name", Avatar: "avatar"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)
					s.Equal("avatar", user.Avatar)
				},
			},
			{
				name: "trigger when save",
				setup: func() {
					user := User{Name: "event_updated_save_name", Avatar: "avatar"}
					s.Nil(query.Save(&user))

					user.Avatar = "event_updated_save_avatar"
					s.Nil(query.Save(&user))
					s.Equal("event_updated_save_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_updated_save_name", user1.Name)
					s.Equal("event_updated_save_avatar", user1.Avatar)
				},
			},
			{
				name: "trigger when update by model",
				setup: func() {
					user := User{Name: "event_updated_model_update_name", Avatar: "avatar"}
					s.Nil(query.Create(&user))

					res, err := query.Model(&user).Update(User{
						Avatar: "event_updated_model_update_avatar",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_updated_model_update_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
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
		user := User{Name: "event_deleting_name", Avatar: "event_deleting_avatar"}
		s.Nil(query.Create(&user))

		res, err := query.Delete(&user)
		s.EqualError(err, "deleting error")
		s.Nil(res)

		var user1 User
		s.Nil(query.Find(&user1, user.ID))
		s.True(user1.ID > 0)
	}
}

func (s *QueryTestSuite) TestEvent_Deleted() {
	for _, query := range s.queries {
		user := User{Name: "event_deleted_name", Avatar: "event_deleted_avatar"}
		s.Nil(query.Create(&user))

		res, err := query.Delete(&user)
		s.EqualError(err, "deleted error")
		s.Nil(res)

		var user1 User
		s.Nil(query.Find(&user1, user.ID))
		s.True(user1.ID == 0)
	}
}

func (s *QueryTestSuite) TestEvent_ForceDeleting() {
	for _, query := range s.queries {
		user := User{Name: "event_force_deleting_name", Avatar: "event_force_deleting_avatar"}
		s.Nil(query.Create(&user))

		res, err := query.ForceDelete(&user)
		s.EqualError(err, "force deleting error")
		s.Nil(res)

		var user1 User
		s.Nil(query.Find(&user1, user.ID))
		s.True(user1.ID > 0)
	}
}

func (s *QueryTestSuite) TestEvent_ForceDeleted() {
	for _, query := range s.queries {
		user := User{Name: "event_force_deleted_name", Avatar: "event_force_deleted_avatar"}
		s.Nil(query.Create(&user))

		res, err := query.ForceDelete(&user)
		s.EqualError(err, "force deleted error")
		s.Nil(res)

		var user1 User
		s.Nil(query.Find(&user1, user.ID))
		s.True(user1.ID == 0)
	}
}

func (s *QueryTestSuite) TestEvent_Retrieved() {
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "trigger when Find",
				setup: func() {
					var user1 User
					s.Nil(query.Where("name", "event_retrieved_name").Find(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when First",
				setup: func() {
					var user1 User
					s.Nil(query.Where("name", "event_retrieved_name").First(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)

					var user2 User
					s.Nil(query.Where("name", "event_retrieved_name1").First(&user2))
					s.True(user2.ID == 0)
					s.Equal("", user2.Name)
				},
			},
			{
				name: "trigger when FirstOr",
				setup: func() {
					var user1 User
					s.Nil(query.Where("name", "event_retrieved_name").Find(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrCreate",
				setup: func() {
					var user1 User
					s.Nil(query.FirstOrCreate(&user1, User{Name: "event_retrieved_name"}))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrFail",
				setup: func() {
					var user1 User
					s.Nil(query.Where("name", "event_retrieved_name").FirstOrFail(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrNew",
				setup: func() {
					var user1 User
					s.Nil(query.FirstOrNew(&user1, User{Name: "event_retrieved_name"}))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
			{
				name: "trigger when FirstOrFail",
				setup: func() {
					var user1 User
					s.Nil(query.Where("name", "event_retrieved_name").FirstOrFail(&user1))
					s.True(user1.ID > 0)
					s.Equal("event_retrieved_name1", user1.Name)
				},
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				user := User{Name: "event_retrieved_name"}
				s.Nil(query.Create(&user))
				s.True(user.ID > 0)

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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("event_creating_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "save",
				setup: func() {
					user := User{Name: "event_saving_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)
					s.Equal("event_saving_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by single column",
				setup: func() {
					user := User{Name: "event_updating_single_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Model(&user).Update("name", "event_updating_single_update_IsDirty_name1")
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("event_updating_single_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_single_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_updating_single_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_single_update_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by map",
				setup: func() {
					user := User{Name: "event_updating_map_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Model(&user).Update(map[string]any{
						"name": "event_updating_map_update_IsDirty_name1",
					})
					s.Nil(err)
					s.Equal(int64(1), res.RowsAffected)
					s.Equal("event_updating_map_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_map_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_updating_map_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_map_update_IsDirty_avatar", user.Avatar)
				},
			},
			{
				name: "update by model",
				setup: func() {
					user := User{Name: "event_updating_model_update_IsDirty_name", Avatar: "is_dirty_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					res, err := query.Model(&user).Update(User{
						Name: "event_updating_model_update_IsDirty_name1",
					})
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("event_updating_model_update_IsDirty_name1", user.Name)
					s.Equal("event_updating_model_update_IsDirty_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
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
		s.Nil(query.Create(&user))
		s.Equal("goravel", user.Avatar)
	}
}

func (s *QueryTestSuite) TestEvent_Query() {
	for _, query := range s.queries {
		user := User{Name: "event_query"}
		s.Nil(query.Create(&user))
		s.True(user.ID > 0)
		s.Equal("event_query", user.Name)

		var user1 User
		s.Nil(query.Where("name", "event_query1").Find(&user1))
		s.True(user1.ID > 0)
	}
}

func (s *QueryTestSuite) TestExec() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			res, err := query.Exec("INSERT INTO users (name, avatar, created_at, updated_at) VALUES ('exec_user', 'exec_avatar', '2023-03-09 18:56:33', '2023-03-09 18:56:35');")
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user User
			err = query.Where("name", "exec_user").First(&user)
			s.Nil(err)
			s.True(user.ID > 0)

			res, err = query.Exec(fmt.Sprintf("UPDATE users set name = 'exec_user1' where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			res, err = query.Exec(fmt.Sprintf("DELETE FROM users where id = %d", user.ID))
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)
		})
	}
}

func (s *QueryTestSuite) TestExists() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "exists_user", Avatar: "exists_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "exists_user", Avatar: "exists_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var t bool
			s.Nil(query.Model(&User{}).Where("name = ?", "exists_user").Exists(&t))
			s.True(t)

			var f bool
			s.Nil(query.Model(&User{}).Where("name = ?", "no_exists_user").Exists(&f))
			s.False(f)
		})
	}
}

func (s *QueryTestSuite) TestFind() {
	for _, query := range s.queries {
		user := User{Name: "find_user"}
		s.Nil(query.Create(&user))
		s.True(user.ID > 0)

		var user2 User
		s.Nil(query.Find(&user2, user.ID))
		s.True(user2.ID > 0)

		var user3 []User
		s.Nil(query.Find(&user3, []uint{user.ID}))
		s.Equal(1, len(user3))

		var user4 []User
		s.Nil(query.Where("id in ?", []uint{user.ID}).Find(&user4))
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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					var user2 User
					s.Nil(query.FindOrFail(&user2, user.ID))
					s.True(user2.ID > 0)
				},
			},
			{
				name: "error",
				setup: func() {
					var user User
					s.ErrorIs(query.FindOrFail(&user, 10000), orm.ErrRecordNotFound)
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
	for driver, query := range s.queries {
		user := User{Name: "first_user"}
		s.Nil(query.Create(&user))
		s.True(user.ID > 0)

		var user1 User
		s.Nil(query.Where("name", "first_user").First(&user1))
		s.True(user1.ID > 0)

		// refresh connection
		s.mockDummyConnection(driver)

		people := People{Body: "first_people"}
		s.Nil(query.Create(&people))
		s.True(people.ID > 0)

		var people1 People
		s.Nil(query.Where("id in ?", []uint{people.ID}).First(&people1))
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
					s.Nil(query.Where("name", "first_or_user").FirstOr(&user, func() error {
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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					var user1 User
					s.Nil(query.Where("name", "first_or_name").Find(&user1))
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
					s.EqualError(query.FirstOrCreate(&user), "query condition is require")
					s.True(user.ID == 0)
				},
			},
			{
				name: "success",
				setup: func() {
					var user User
					s.Nil(query.FirstOrCreate(&user, User{Name: "first_or_create_user"}))
					s.True(user.ID > 0)
					s.Equal("first_or_create_user", user.Name)

					var user1 User
					s.Nil(query.FirstOrCreate(&user1, User{Name: "first_or_create_user"}))
					s.Equal(user.ID, user1.ID)

					var user2 User
					s.Nil(query.Where("avatar", "first_or_create_avatar").FirstOrCreate(&user2, User{Name: "user"}, User{Avatar: "first_or_create_avatar2"}))
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
					s.Equal(orm.ErrRecordNotFound, query.Where("name", "first_or_fail_user").FirstOrFail(&user))
					s.Equal(uint(0), user.ID)
				},
			},
			{
				name: "success",
				setup: func() {
					user := User{Name: "first_or_fail_name"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("first_or_fail_name", user.Name)

					var user1 User
					s.Nil(query.Where("name", "first_or_fail_name").FirstOrFail(&user1))
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
					s.Nil(query.FirstOrNew(&user, User{Name: "first_or_new_name"}))
					s.Equal(uint(0), user.ID)
					s.Equal("first_or_new_name", user.Name)
					s.Equal("", user.Avatar)

					var user1 User
					s.Nil(query.FirstOrNew(&user1, User{Name: "first_or_new_name"}, User{Avatar: "first_or_new_avatar"}))
					s.Equal(uint(0), user1.ID)
					s.Equal("first_or_new_name", user1.Name)
					s.Equal("first_or_new_avatar", user1.Avatar)
				},
			},
			{
				name: "found",
				setup: func() {
					user := User{Name: "first_or_new_name"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("first_or_new_name", user.Name)

					var user1 User
					s.Nil(query.FirstOrNew(&user1, User{Name: "first_or_new_name"}))
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
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)
					s.Equal("force_delete_name", user.Name)

					res, err := query.Where("name = ?", "force_delete_name").ForceDelete(&User{})
					s.Equal(int64(1), res.RowsAffected)
					s.Nil(err)
					s.Equal("force_delete_name", user.Name)

					var user1 User
					s.Nil(query.WithTrashed().Find(&user1, user.ID))
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
		s.Run(driver.String(), func() {
			user := User{Name: "get_user"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user1 []User
			s.Nil(query.Where("id in ?", []uint{user.ID}).Get(&user1))
			s.Equal(1, len(user1))

			// refresh connection
			s.mockDummyConnection(driver)

			people := People{Body: "get_people"}
			s.Nil(query.Create(&people))
			s.True(people.ID > 0)

			var people1 []People
			s.Nil(query.Where("id in ?", []uint{people.ID}).Get(&people1))
			s.Equal(1, len(people1))

			var user2 []User
			s.Nil(query.Where("id in ?", []uint{user.ID}).Get(&user2))
			s.Equal(1, len(user2))
		})
		break
	}
}

func (s *QueryTestSuite) TestJoin() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "join_user", Avatar: "join_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			userAddress := Address{UserID: user.ID, Name: "join_address", Province: "join_province"}
			s.Nil(query.Create(&userAddress))
			s.True(userAddress.ID > 0)

			type Result struct {
				UserName        string
				UserAddressName string
			}
			var result []Result
			s.Nil(query.Model(&User{}).Where("users.id = ?", user.ID).Join("left join addresses ua on users.id = ua.user_id").
				Select("users.name user_name, ua.name user_address_name").Get(&result))
			s.Equal(1, len(result))
			s.Equal("join_user", result[0].UserName)
			s.Equal("join_address", result[0].UserAddressName)
		})
	}
}

func (s *QueryTestSuite) TestLockForUpdate() {
	for driver, query := range s.queries {
		if driver != ormcontract.DriverSqlite {
			s.Run(driver.String(), func() {
				user := User{Name: "lock_for_update_user"}
				s.Nil(query.Create(&user))
				s.True(user.ID > 0)

				for i := 0; i < 10; i++ {
					go func() {
						tx, err := query.Begin()
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
				s.Nil(query.Find(&user2, user.ID))
				s.Equal("lock_for_update_user1111111111", user2.Name)
			})
		}
	}
}

func (s *QueryTestSuite) TestOffset() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "offset_user", Avatar: "offset_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "offset_user", Avatar: "offset_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "offset_user").Offset(1).Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestOrder() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "order_user", Avatar: "order_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_user", Avatar: "order_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "order_user").Order("id desc").Order("name asc").Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestOrderBy() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "order_asc_user", Avatar: "order_asc_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_asc_user", Avatar: "order_asc_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users1 []User
			s.Nil(query.Where("name = ?", "order_asc_user").OrderBy("id").Get(&users1))
			s.True(len(users1) == 2)
			s.True(users1[0].ID == user.ID)

			var users2 []User
			s.Nil(query.Where("name = ?", "order_asc_user").OrderBy("id", "DESC").Get(&users2))
			s.True(len(users2) == 2)
			s.True(users2[0].ID == user1.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrderByDesc() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "order_desc_user", Avatar: "order_desc_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "order_desc_user", Avatar: "order_desc_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "order_desc_user").OrderByDesc("id").Get(&users))
			usersLength := len(users)
			s.True(usersLength == 2)
			s.True(users[usersLength-1].ID == user.ID)
		})
	}
}

func (s *QueryTestSuite) TestInRandomOrder() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			for i := 0; i < 30; i++ {
				user := User{Name: "random_order_user", Avatar: "random_order_avatar"}
				s.Nil(query.Create(&user))
				s.True(user.ID > 0)
			}

			var users1 []User
			s.Nil(query.Where("name = ?", "random_order_user").InRandomOrder().Find(&users1))
			s.True(len(users1) == 30)

			var users2 []User
			s.Nil(query.Where("name = ?", "random_order_user").InRandomOrder().Find(&users2))
			s.True(len(users2) == 30)

			s.True(users1[0].ID != users2[0].ID || users1[14].ID != users2[14].ID || users1[29].ID != users2[29].ID)
		})
	}
}

func (s *QueryTestSuite) TestPaginate() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "paginate_user", Avatar: "paginate_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "paginate_user", Avatar: "paginate_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "paginate_user", Avatar: "paginate_avatar2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "paginate_user", Avatar: "paginate_avatar3"}
			s.Nil(query.Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "paginate_user").Paginate(1, 3, &users, nil))
			s.Equal(3, len(users))

			var users1 []User
			var total1 int64
			s.Nil(query.Where("name = ?", "paginate_user").Paginate(2, 3, &users1, &total1))
			s.Equal(1, len(users1))
			s.Equal(int64(4), total1)

			var users2 []User
			var total2 int64
			s.Nil(query.Model(User{}).Where("name = ?", "paginate_user").Paginate(1, 3, &users2, &total2))
			s.Equal(3, len(users2))
			s.Equal(int64(4), total2)

			var users3 []User
			var total3 int64
			s.Nil(query.Table("users").Where("name = ?", "paginate_user").Paginate(1, 3, &users3, &total3))
			s.Equal(3, len(users3))
			s.Equal(int64(4), total3)
		})
	}
}

func (s *QueryTestSuite) TestPluck() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "pluck_user", Avatar: "pluck_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "pluck_user", Avatar: "pluck_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var avatars []string
			s.Nil(query.Model(&User{}).Where("name = ?", "pluck_user").Pluck("avatar", &avatars))
			s.Equal(2, len(avatars))
			s.Equal("pluck_avatar", avatars[0])
			s.Equal("pluck_avatar1", avatars[1])
		})
	}
}

func (s *QueryTestSuite) TestHasOne() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_one_name",
				Address: &Address{
					Name: "has_one_address",
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)

			var user1 User
			s.Nil(query.With("Address").Where("name = ?", "has_one_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Address.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestHasOneMorph() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_one_morph_name",
				House: &House{
					Name: "has_one_morph_house",
				},
			}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.House.ID > 0)

			var user1 User
			s.Nil(query.With("House").Where("name = ?", "has_one_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_one_morph_name")
			s.True(user.House.ID > 0)
			s.True(user.House.Name == "has_one_morph_house")

			var house House
			s.Nil(query.Where("name = ?", "has_one_morph_house").Where("houseable_type = ?", "users").Where("houseable_id = ?", user.ID).First(&house))
			s.True(house.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestHasMany() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_many_name",
				Books: []*Book{
					{Name: "has_many_book1"},
					{Name: "has_many_book2"},
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[1].ID > 0)

			var user1 User
			s.Nil(query.With("Books").Where("name = ?", "has_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Books) == 2)
		})
	}
}

func (s *QueryTestSuite) TestHasManyMorph() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "has_many_morph_name",
				Phones: []*Phone{
					{Name: "has_many_morph_phone1"},
					{Name: "has_many_morph_phone2"},
				},
			}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Phones[0].ID > 0)
			s.True(user.Phones[1].ID > 0)

			var user1 User
			s.Nil(query.With("Phones").Where("name = ?", "has_many_morph_name").First(&user1))
			s.True(user.ID > 0)
			s.True(user.Name == "has_many_morph_name")
			s.True(len(user.Phones) == 2)
			s.True(user.Phones[0].Name == "has_many_morph_phone1")
			s.True(user.Phones[1].Name == "has_many_morph_phone2")

			var phones []Phone
			s.Nil(query.Where("name like ?", "has_many_morph_phone%").Where("phoneable_type = ?", "users").Where("phoneable_id = ?", user.ID).Find(&phones))
			s.True(len(phones) == 2)
		})
	}
}

func (s *QueryTestSuite) TestManyToMany() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := &User{
				Name: "many_to_many_name",
				Roles: []*Role{
					{Name: "many_to_many_role1"},
					{Name: "many_to_many_role2"},
				},
			}

			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Roles[0].ID > 0)
			s.True(user.Roles[1].ID > 0)

			var user1 User
			s.Nil(query.With("Roles").Where("name = ?", "many_to_many_name").First(&user1))
			s.True(user.ID > 0)
			s.True(len(user.Roles) == 2)

			var role Role
			s.Nil(query.With("Users").Where("name = ?", "many_to_many_role1").First(&role))
			s.True(role.ID > 0)
			s.True(len(role.Users) == 1)
			s.Equal("many_to_many_name", role.Users[0].Name)
		})
	}
}

func (s *QueryTestSuite) TestLimit() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "limit_user", Avatar: "limit_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "limit_user", Avatar: "limit_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "limit_user").Limit(1).Get(&user2))
			s.True(len(user2) > 0)
			s.True(user2[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestLoad() {
	for _, query := range s.queries {
		user := User{Name: "load_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
		user.Address.Name = "load_address"
		user.Books[0].Name = "load_book0"
		user.Books[1].Name = "load_book1"
		s.Nil(query.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Address.ID > 0)
		s.True(user.Books[0].ID > 0)
		s.True(user.Books[1].ID > 0)

		tests := []struct {
			description string
			setup       func(description string)
		}{
			{
				description: "simple load relationship",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.True(len(user1.Books) == 0)
					s.Nil(query.Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 0)
					s.Nil(query.Load(&user1, "Books"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 2)
				},
			},
			{
				description: "load relationship with simple condition",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(query.Load(&user1, "Books", "name = ?", "load_book0"))
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
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(query.Load(&user1, "Books", func(query ormcontract.Query) ormcontract.Query {
						return query.Where("name = ?", "load_book0")
					}))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(1, len(user1.Books))
					s.Equal("load_book0", user.Books[0].Name)
				},
			},
			{
				description: "error when relation is empty",
				setup: func(description string) {
					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.EqualError(query.Load(&user1, ""), "relation cannot be empty")
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
					s.EqualError(query.Load(&userNoID, "Book"), "id cannot be empty")
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
		s.Run(driver.String(), func() {
			user := User{Name: "load_missing_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
			user.Address.Name = "load_missing_address"
			user.Books[0].Name = "load_missing_book0"
			user.Books[1].Name = "load_missing_book1"
			s.Nil(query.Select(orm.Associations).Create(&user))
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
						s.Nil(query.Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 0)
						s.Nil(query.LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 0)
						s.Nil(query.LoadMissing(&user1, "Books"))
						s.True(user1.Address.ID > 0)
						s.True(len(user1.Books) == 2)
					},
				},
				{
					description: "don't load when not missing",
					setup: func(description string) {
						var user1 User
						s.Nil(query.With("Books", "name = ?", "load_missing_book0").Find(&user1, user.ID))
						s.True(user1.ID > 0)
						s.Nil(user1.Address)
						s.True(len(user1.Books) == 1)
						s.Nil(query.LoadMissing(&user1, "Address"))
						s.True(user1.Address.ID > 0)
						s.Nil(query.LoadMissing(&user1, "Books"))
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

func (s *QueryTestSuite) TestRaw() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "raw_user", Avatar: "raw_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			var user1 User
			s.Nil(query.Raw("SELECT id, name FROM users WHERE name = ?", "raw_user").Scan(&user1))
			s.True(user1.ID > 0)
			s.Equal("raw_user", user1.Name)
			s.Equal("", user1.Avatar)
		})
	}
}

func (s *QueryTestSuite) TestReuse() {
	for _, query := range s.queries {
		users := []User{{Name: "reuse_user", Avatar: "reuse_avatar"}, {Name: "reuse_user1", Avatar: "reuse_avatar1"}}
		s.Nil(query.Create(&users))
		s.True(users[0].ID > 0)
		s.True(users[1].ID > 0)

		q := query.Where("name", "reuse_user")

		var users1 User
		s.Nil(q.Where("avatar", "reuse_avatar").Find(&users1))
		s.True(users1.ID > 0)

		var users2 User
		s.Nil(q.Where("avatar", "reuse_avatar1").Find(&users2))
		s.True(users2.ID == 0)

		var users3 User
		s.Nil(query.Where("avatar", "reuse_avatar1").Find(&users3))
		s.True(users3.ID > 0)
	}
}

func (s *QueryTestSuite) TestRefreshConnection() {
	tests := []struct {
		name             string
		model            any
		setup            func()
		expectConnection string
		expectErr        string
	}{
		{
			name: "invalid model",
			model: func() any {
				var product string
				return product
			}(),
			setup:     func() {},
			expectErr: "invalid model",
		},
		{
			name: "the connection of model is empty",
			model: func() any {
				var review Review
				return review
			}(),
			setup:            func() {},
			expectConnection: "mysql",
		},
		{
			name: "the connection of model is same as current connection",
			model: func() any {
				var box Box
				return box
			}(),
			setup:            func() {},
			expectConnection: "mysql",
		},
		{
			name: "connections are different, but drivers are same",
			model: func() any {
				var people People
				return people
			}(),
			setup: func() {
				mockDummyConnection(s.mysqlDocker.MockConfig, testDatabaseDocker.Mysql1.Config())
			},
			expectConnection: "dummy",
		},
		{
			name: "connections and drivers are different",
			model: func() any {
				var product Product
				return product
			}(),
			setup: func() {
				mockPostgresqlConnection(s.mysqlDocker.MockConfig, testDatabaseDocker.Postgresql.Config())
			},
			expectConnection: "postgresql",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			test.setup()
			queryImpl := s.queries[ormcontract.DriverMysql].(*QueryImpl)
			query, err := queryImpl.refreshConnection(test.model)
			if test.expectErr != "" {
				s.EqualError(err, test.expectErr)
			} else {
				s.Nil(err)
			}
			if test.expectConnection == "" {
				s.Nil(query)
			} else {
				s.Equal(test.expectConnection, query.connection)
			}
		})
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
					user := User{Name: "save_create_user", Avatar: "save_create_avatar"}
					s.Nil(query.Save(&user))
					s.True(user.ID > 0)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("save_create_user", user1.Name)
				},
			},
			{
				name: "success when update",
				setup: func() {
					user := User{Name: "save_update_user", Avatar: "save_update_avatar"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					user.Name = "save_update_user1"
					s.Nil(query.Save(&user))

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("save_update_user1", user1.Name)
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
		s.Nil(query.SaveQuietly(&user))
		s.True(user.ID > 0)
		s.Equal("event_save_quietly_name", user.Name)
		s.Equal("save_quietly_avatar", user.Avatar)

		var user1 User
		s.Nil(query.Find(&user1, user.ID))
		s.Equal("event_save_quietly_name", user1.Name)
		s.Equal("save_quietly_avatar", user1.Avatar)
	}
}

func (s *QueryTestSuite) TestScope() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			users := []User{{Name: "scope_user", Avatar: "scope_avatar"}, {Name: "scope_user1", Avatar: "scope_avatar1"}}
			s.Nil(query.Create(&users))
			s.True(users[0].ID > 0)
			s.True(users[1].ID > 0)

			var users1 []User
			s.Nil(query.Scopes(paginator("1", "1")).Find(&users1))

			s.Equal(1, len(users1))
			s.True(users1[0].ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestSelect() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "select_user", Avatar: "select_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "select_user", Avatar: "select_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "select_user1", Avatar: "select_avatar1"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			type Result struct {
				Name  string
				Count string
			}
			var result []Result
			s.Nil(query.Model(&User{}).Select("name, count(avatar) as count").Where("id in ?", []uint{user.ID, user1.ID, user2.ID}).Group("name").Get(&result))
			s.Equal(2, len(result))
			s.Equal("select_user", result[0].Name)
			s.Equal("2", result[0].Count)
			s.Equal("select_user1", result[1].Name)
			s.Equal("1", result[1].Count)

			var result1 []Result
			s.Nil(query.Model(&User{}).Select("name, count(avatar) as count").Group("name").Having("name = ?", "select_user").Get(&result1))

			s.Equal(1, len(result1))
			s.Equal("select_user", result1[0].Name)
			s.Equal("2", result1[0].Count)
		})
	}
}

func (s *QueryTestSuite) TestSharedLock() {
	for driver, query := range s.queries {
		if driver != ormcontract.DriverSqlite {
			s.Run(driver.String(), func() {
				user := User{Name: "shared_lock_user"}
				s.Nil(query.Create(&user))
				s.True(user.ID > 0)

				tx, err := query.Begin()
				s.Nil(err)
				var user1 User
				s.Nil(tx.SharedLock().Find(&user1, user.ID))
				s.True(user1.ID > 0)

				var user2 User
				s.Nil(query.SharedLock().Find(&user2, user.ID))
				s.True(user2.ID > 0)

				user1.Name += "1"
				s.Nil(tx.Save(&user1))

				s.Nil(tx.Commit())

				var user3 User
				s.Nil(query.Find(&user3, user.ID))
				s.Equal("shared_lock_user1", user3.Name)
			})
		}
	}
}

func (s *QueryTestSuite) TestSoftDelete() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "soft_delete_user", Avatar: "soft_delete_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			res, err := query.Where("name = ?", "soft_delete_user").Delete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user1 User
			s.Nil(query.Find(&user1, user.ID))
			s.Equal(uint(0), user1.ID)

			var user2 User
			s.Nil(query.WithTrashed().Find(&user2, user.ID))
			s.True(user2.ID > 0)

			res, err = query.Where("name = ?", "soft_delete_user").ForceDelete(&User{})
			s.Equal(int64(1), res.RowsAffected)
			s.Nil(err)

			var user3 User
			s.Nil(query.WithTrashed().Find(&user3, user.ID))
			s.Equal(uint(0), user3.ID)
		})
	}
}

func (s *QueryTestSuite) TestSum() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "count_user", Avatar: "count_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "count_user", Avatar: "count_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var value float64
			err := query.Table("users").Sum("id", &value)
			s.Nil(err)
			s.True(value > 0)
		})
	}
}

func (s *QueryTestSuite) TestToSql() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			switch driver {
			case ormcontract.DriverPostgresql:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", query.Where("id", 1).ToSql().Find(User{}))
			case ormcontract.DriverSqlserver:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = @p1 AND \"users\".\"deleted_at\" IS NULL", query.Where("id", 1).ToSql().Find(User{}))
			default:
				s.Equal("SELECT * FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", query.Where("id", 1).ToSql().Find(User{}))
			}
		})
	}
}

func (s *QueryTestSuite) TestToRawSql() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			switch driver {
			case ormcontract.DriverPostgresql:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", query.Where("id", 1).ToRawSql().Find(User{}))
			case ormcontract.DriverSqlserver:
				s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1$ AND \"users\".\"deleted_at\" IS NULL", query.Where("id", 1).ToRawSql().Find(User{}))
			default:
				s.Equal("SELECT * FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", query.Where("id", 1).ToRawSql().Find(User{}))
			}
		})
	}
}

func (s *QueryTestSuite) TestTransactionSuccess() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
			user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
			tx, err := query.Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Commit())

			var user2, user3 User
			s.Nil(query.Find(&user2, user.ID))
			s.Nil(query.Find(&user3, user1.ID))
		})
	}
}

func (s *QueryTestSuite) TestTransactionError() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			tx, err := query.Begin()
			s.Nil(err)
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))
			s.Nil(tx.Rollback())

			var users []User
			s.Nil(query.Where("name = ? or name = ?", "transaction_error_user", "transaction_error_user1").Find(&users))
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
					s.Nil(query.Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Model(&User{}).Where("name = ?", "updates_single_name").Update("avatar", "update_single_avatar2")
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Find(&user2, users[0].ID))
					s.Equal("update_single_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Find(&user3, users[1].ID))
					s.Equal("update_single_avatar2", user3.Avatar)
				},
			},
			{
				name: "update columns by map, success",
				setup: func() {
					users := []User{{Name: "update_map_name", Avatar: "update_map_avatar"}, {Name: "update_map_name", Avatar: "update_map_avatar1"}}
					s.Nil(query.Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Model(&User{}).Where("name = ?", "update_map_name").Update(map[string]any{
						"avatar": "update_map_avatar2",
					})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Find(&user2, users[0].ID))
					s.Equal("update_map_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Find(&user3, users[0].ID))
					s.Equal("update_map_avatar2", user3.Avatar)
				},
			},
			{
				name: "update columns by model, success",
				setup: func() {
					users := []User{{Name: "update_model_name", Avatar: "update_model_avatar"}, {Name: "update_model_name", Avatar: "update_model_avatar1"}}
					s.Nil(query.Create(&users))
					s.True(users[0].ID > 0)
					s.True(users[1].ID > 0)

					res, err := query.Model(&User{}).Where("name = ?", "update_model_name").Update(User{Avatar: "update_model_avatar2"})
					s.Equal(int64(2), res.RowsAffected)
					s.Nil(err)

					var user2 User
					s.Nil(query.Find(&user2, users[0].ID))
					s.Equal("update_model_avatar2", user2.Avatar)
					var user3 User
					s.Nil(query.Find(&user3, users[0].ID))
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
		s.Run(driver.String(), func() {
			var user User
			err := query.UpdateOrCreate(&user, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar"})
			s.Nil(err)
			s.True(user.ID > 0)

			var user1 User
			err = query.Where("name", "update_or_create_user").Find(&user1)
			s.Nil(err)
			s.True(user1.ID > 0)

			var user2 User
			err = query.UpdateOrCreate(&user2, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar1"})
			s.Nil(err)
			s.True(user2.ID > 0)
			s.Equal("update_or_create_avatar1", user2.Avatar)

			var user3 User
			err = query.Where("avatar", "update_or_create_avatar1").Find(&user3)
			s.Nil(err)
			s.True(user3.ID > 0)

			var count int64
			err = query.Model(User{}).Where("name", "update_or_create_user").Count(&count)
			s.Nil(err)
			s.Equal(int64(1), count)
		})
	}
}

func (s *QueryTestSuite) TestWhere() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_user", Avatar: "where_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_user1", Avatar: "where_avatar1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var user2 []User
			s.Nil(query.Where("name = ?", "where_user").OrWhere("avatar = ?", "where_avatar1").Find(&user2))
			s.Equal(2, len(user2))

			var user3 User
			s.Nil(query.Where("name = 'where_user'").Find(&user3))
			s.True(user3.ID > 0)

			var user4 User
			s.Nil(query.Where("name", "where_user").Find(&user4))
			s.True(user4.ID > 0)
		})
	}
}

func (s *QueryTestSuite) TestWhereIn() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.WhereIn("id", []any{user.ID, user1.ID}).Find(&users))
			s.True(len(users) == 2)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereIn() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Where("id = ?", -1).OrWhereIn("id", []any{user.ID, user1.ID}).Find(&users))
			s.True(len(users) == 2)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotIn() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_in_user_2", Avatar: "where_in_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			var user3 User
			s.Nil(query.Where("id = ?", user2.ID).WhereNotIn("id", []any{user.ID, user1.ID}).First(&user3))
			s.True(user3.ID == user2.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNotIn() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_in_user", Avatar: "where_in_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_in_user_1", Avatar: "where_in_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_in_user_2", Avatar: "where_in_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			var users []User
			s.Nil(query.Where("id = ?", -1).OrWhereNotIn("id", []any{user.ID, user1.ID}).Find(&users))
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
		s.Run(driver.String(), func() {
			user := User{Name: "where_between_user", Avatar: "where_between_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_between_user_1", Avatar: "where_between_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_between_user_2", Avatar: "where_between_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			var users []User
			s.Nil(query.WhereBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 3)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotBetween() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "where_not_between_user", Avatar: "where_not_between_avatar_2"}
			s.Nil(query.Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "where_not_between_user").WhereNotBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 1)
			s.True(users[0].ID == user3.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereBetween() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "or_where_between_user", Avatar: "or_where_between_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_between_user_1", Avatar: "or_where_between_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "or_where_between_user_2", Avatar: "or_where_between_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "or_where_between_user_3", Avatar: "or_where_between_avatar_3"}
			s.Nil(query.Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "or_where_between_user_3").OrWhereBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) == 4)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNotBetween() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			user := User{Name: "or_where_between_user", Avatar: "or_where_between_avatar"}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_between_user_1", Avatar: "or_where_between_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			user2 := User{Name: "or_where_between_user_2", Avatar: "or_where_between_avatar_2"}
			s.Nil(query.Create(&user2))
			s.True(user2.ID > 0)

			user3 := User{Name: "or_where_between_user_3", Avatar: "or_where_between_avatar_3"}
			s.Nil(query.Create(&user3))
			s.True(user3.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "or_where_between_user_3").OrWhereNotBetween("id", user.ID, user2.ID).Find(&users))
			s.True(len(users) >= 1)
		})
	}
}

func (s *QueryTestSuite) TestWhereNull() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			bio := "where_null_bio"
			user := User{Name: "where_null_user", Avatar: "where_null_avatar", Bio: &bio}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_null_user", Avatar: "where_null_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "where_null_user").WhereNull("bio").Find(&users))
			s.True(len(users) == 1)
			s.True(users[0].ID == user1.ID)
		})
	}
}

func (s *QueryTestSuite) TestOrWhereNull() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			bio := "or_where_null_bio"
			user := User{Name: "or_where_null_user", Avatar: "or_where_null_avatar", Bio: &bio}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "or_where_null_user_1", Avatar: "or_where_null_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "or_where_null_user").OrWhereNull("bio").Find(&users))
			s.True(len(users) >= 2)
		})
	}
}

func (s *QueryTestSuite) TestWhereNotNull() {
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
			bio := "where_not_null_bio"
			user := User{Name: "where_not_null_user", Avatar: "where_not_null_avatar", Bio: &bio}
			s.Nil(query.Create(&user))
			s.True(user.ID > 0)

			user1 := User{Name: "where_not_null_user", Avatar: "where_not_null_avatar_1"}
			s.Nil(query.Create(&user1))
			s.True(user1.ID > 0)

			var users []User
			s.Nil(query.Where("name = ?", "where_not_null_user").WhereNotNull("bio").Find(&users))
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
					s.Nil(query.WithoutEvents().Save(&user))
					s.True(user.ID > 0)
					s.Equal("without_events_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
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
		s.Run(driver.String(), func() {
			user := User{Name: "with_user", Address: &Address{
				Name: "with_address",
			}, Books: []*Book{{
				Name: "with_book0",
			}, {
				Name: "with_book1",
			}}}
			s.Nil(query.Select(orm.Associations).Create(&user))
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
						s.Nil(query.With("Address").With("Books").Find(&user1, user.ID))
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
						s.Nil(query.With("Books", "name = ?", "with_book0").Find(&user1, user.ID))
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
						s.Nil(query.With("Books", func(query ormcontract.Query) ormcontract.Query {
							return query.Where("name = ?", "with_book0")
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
		s.Run(driver.String(), func() {
			user := User{Name: "with_nesting_user", Books: []*Book{{
				Name:   "with_nesting_book0",
				Author: &Author{Name: "with_nesting_author0"},
			}, {
				Name:   "with_nesting_book1",
				Author: &Author{Name: "with_nesting_author1"},
			}}}
			s.Nil(query.Select(orm.Associations).Create(&user))
			s.True(user.ID > 0)
			s.True(user.Books[0].ID > 0)
			s.True(user.Books[0].Author.ID > 0)
			s.True(user.Books[1].ID > 0)
			s.True(user.Books[1].Author.ID > 0)

			var user1 User
			s.Nil(query.With("Books.Author").Find(&user1, user.ID))
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

func (s *QueryTestSuite) mockDummyConnection(driver ormcontract.Driver) {
	switch driver {
	case ormcontract.DriverMysql:
		mockDummyConnection(s.mysqlDocker.MockConfig, testDatabaseDocker.Mysql1.Config())
	case ormcontract.DriverPostgresql:
		mockDummyConnection(s.postgresqlDocker.MockConfig, testDatabaseDocker.Mysql1.Config())
	case ormcontract.DriverSqlite:
		mockDummyConnection(s.sqliteDocker.MockConfig, testDatabaseDocker.Mysql1.Config())
	case ormcontract.DriverSqlserver:
		mockDummyConnection(s.sqlserverDocker.MockConfig, testDatabaseDocker.Mysql1.Config())
	}
}

func TestCustomConnection(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	mysqlDocker := NewMysqlDocker(testDatabaseDocker)
	query, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}
	postgresqlDocker := NewPostgresqlDocker(testDatabaseDocker)
	_, err = postgresqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	review := Review{Body: "create_review"}
	assert.Nil(t, query.Create(&review))
	assert.True(t, review.ID > 0)

	var review1 Review
	assert.Nil(t, query.Where("body", "create_review").First(&review1))
	assert.True(t, review1.ID > 0)

	mockPostgresqlConnection(mysqlDocker.MockConfig, testDatabaseDocker.Postgresql.Config())

	product := Product{Name: "create_product"}
	assert.Nil(t, query.Create(&product))
	assert.True(t, product.ID > 0)

	var product1 Product
	assert.Nil(t, query.Where("name", "create_product").First(&product1))
	assert.True(t, product1.ID > 0)

	var product2 Product
	assert.Nil(t, query.Where("name", "create_product1").First(&product2))
	assert.True(t, product2.ID == 0)

	mockDummyConnection(mysqlDocker.MockConfig, testDatabaseDocker.Mysql.Config())

	person := Person{Name: "create_person"}
	assert.NotNil(t, query.Create(&person))
	assert.True(t, person.ID == 0)
}

func TestFilterFindConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []any
		expectErr  error
	}{
		{
			name: "condition is empty",
		},
		{
			name:       "condition is empty string",
			conditions: []any{""},
			expectErr:  ErrorMissingWhereClause,
		},
		{
			name:       "condition is empty slice",
			conditions: []any{[]string{}},
			expectErr:  ErrorMissingWhereClause,
		},
		{
			name:       "condition has value",
			conditions: []any{"name = ?", "test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := filterFindConditions(test.conditions...)
			if test.expectErr != nil {
				assert.Equal(t, err, test.expectErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetModelConnection(t *testing.T) {
	tests := []struct {
		name             string
		model            any
		expectErr        string
		expectConnection string
	}{
		{
			name: "invalid model",
			model: func() any {
				var product string
				return product
			}(),
			expectErr: "invalid model",
		},
		{
			name: "not ConnectionModel",
			model: func() any {
				var phone Phone
				return phone
			}(),
		},
		{
			name: "the connection of model is empty",
			model: func() any {
				var review Review
				return review
			}(),
		},
		{
			name: "the connection of model is not empty",
			model: func() any {
				var product Product
				return product
			}(),
			expectConnection: "postgresql",
		},
		{
			name: "the connection of model is not empty and model is slice",
			model: func() any {
				var products []Product
				return products
			}(),
			expectConnection: "postgresql",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connection, err := getModelConnection(test.model)
			if test.expectErr != "" {
				assert.EqualError(t, err, test.expectErr)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.expectConnection, connection)
		})
	}
}

func TestObserver(t *testing.T) {
	orm.Observers = append(orm.Observers, orm.Observer{
		Model:    User{},
		Observer: &UserObserver{},
	})

	assert.Nil(t, observer(Product{}))
	assert.Equal(t, &UserObserver{}, observer(User{}))
}

func TestObserverEvent(t *testing.T) {
	assert.EqualError(t, observerEvent(ormcontract.EventRetrieved, &UserObserver{})(nil), "retrieved")
	assert.EqualError(t, observerEvent(ormcontract.EventCreating, &UserObserver{})(nil), "creating")
	assert.EqualError(t, observerEvent(ormcontract.EventCreated, &UserObserver{})(nil), "created")
	assert.EqualError(t, observerEvent(ormcontract.EventUpdating, &UserObserver{})(nil), "updating")
	assert.EqualError(t, observerEvent(ormcontract.EventUpdated, &UserObserver{})(nil), "updated")
	assert.EqualError(t, observerEvent(ormcontract.EventSaving, &UserObserver{})(nil), "saving")
	assert.EqualError(t, observerEvent(ormcontract.EventSaved, &UserObserver{})(nil), "saved")
	assert.EqualError(t, observerEvent(ormcontract.EventDeleting, &UserObserver{})(nil), "deleting")
	assert.EqualError(t, observerEvent(ormcontract.EventDeleted, &UserObserver{})(nil), "deleted")
	assert.EqualError(t, observerEvent(ormcontract.EventForceDeleting, &UserObserver{})(nil), "forceDeleting")
	assert.EqualError(t, observerEvent(ormcontract.EventForceDeleted, &UserObserver{})(nil), "forceDeleted")
	assert.Nil(t, observerEvent("error", &UserObserver{}))
}

func TestReadWriteSeparate(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	writeDatabaseDocker, err := supportdocker.InitDatabase()
	if err != nil {
		log.Fatalf("Init docker error: %s", err)
	}

	readMysqlDocker := NewMysqlDocker(testDatabaseDocker)
	readMysqlQuery, err := readMysqlDocker.New()
	if err != nil {
		log.Fatalf("Get read mysql error: %s", err)
	}

	writeMysqlDocker := NewMysqlDocker(writeDatabaseDocker)
	writeMysqlQuery, err := writeMysqlDocker.New()
	if err != nil {
		log.Fatalf("Get write mysql error: %s", err)
	}

	writeMysqlDocker.MockReadWrite(readMysqlDocker.Port, writeMysqlDocker.Port)
	mysqlQuery, err := writeMysqlDocker.Query(false)
	if err != nil {
		log.Fatalf("Get mysql gorm error: %s", err)
	}

	readPostgresqlDocker := NewPostgresqlDocker(testDatabaseDocker)
	readPostgresqlQuery, err := readPostgresqlDocker.New()
	if err != nil {
		log.Fatalf("Get read postgresql error: %s", err)
	}

	writePostgresqlDocker := NewPostgresqlDocker(writeDatabaseDocker)
	writePostgresqlQuery, err := writePostgresqlDocker.New()
	if err != nil {
		log.Fatalf("Get write postgresql error: %s", err)
	}

	writePostgresqlDocker.MockReadWrite(readPostgresqlDocker.Port, writePostgresqlDocker.Port)
	postgresqlQuery, err := writePostgresqlDocker.Query(false)
	if err != nil {
		log.Fatalf("Get postgresql gorm error: %s", err)
	}

	readSqliteDocker := NewSqliteDocker(dbDatabase)
	readSqliteQuery, err := readSqliteDocker.New()
	if err != nil {
		log.Fatalf("Get read sqlite error: %s", err)
	}

	writeSqliteDocker := NewSqliteDocker(dbDatabase1)
	writeSqliteQuery, err := writeSqliteDocker.New()
	if err != nil {
		log.Fatalf("Get write sqlite error: %s", err)
	}

	writeSqliteDocker.MockReadWrite()
	sqliteDB, err := writeSqliteDocker.Query(false)
	if err != nil {
		log.Fatalf("Get sqlite gorm error: %s", err)
	}

	readSqlserverDocker := NewSqlserverDocker(testDatabaseDocker)
	readSqlserverQuery, err := readSqlserverDocker.New()
	if err != nil {
		log.Fatalf("Get read sqlserver error: %s", err)
	}

	writeSqlserverDocker := NewSqlserverDocker(writeDatabaseDocker)
	writeSqlserverQuery, err := writeSqlserverDocker.New()
	if err != nil {
		log.Fatalf("Get write sqlserver error: %s", err)
	}
	writeSqlserverDocker.MockReadWrite(readSqlserverDocker.Port, writeSqlserverDocker.Port)
	sqlserverDB, err := writeSqlserverDocker.Query(false)
	if err != nil {
		log.Fatalf("Get sqlserver gorm error: %s", err)
	}

	dbs := map[ormcontract.Driver]map[string]ormcontract.Query{
		ormcontract.DriverMysql: {
			"mix":   mysqlQuery,
			"read":  readMysqlQuery,
			"write": writeMysqlQuery,
		},
		ormcontract.DriverPostgresql: {
			"mix":   postgresqlQuery,
			"read":  readPostgresqlQuery,
			"write": writePostgresqlQuery,
		},
		ormcontract.DriverSqlite: {
			"mix":   sqliteDB,
			"read":  readSqliteQuery,
			"write": writeSqliteQuery,
		},
		ormcontract.DriverSqlserver: {
			"mix":   sqlserverDB,
			"read":  readSqlserverQuery,
			"write": writeSqlserverQuery,
		},
	}

	for drive, db := range dbs {
		t.Run(drive.String(), func(t *testing.T) {
			user := User{Name: "user"}
			assert.Nil(t, db["mix"].Create(&user))
			assert.True(t, user.ID > 0)

			var user2 User
			assert.Nil(t, db["mix"].Find(&user2, user.ID))
			assert.True(t, user2.ID == 0)

			var user3 User
			assert.Nil(t, db["read"].Find(&user3, user.ID))
			assert.True(t, user3.ID == 0)

			var user4 User
			assert.Nil(t, db["write"].Find(&user4, user.ID))
			assert.True(t, user4.ID > 0)
		})
	}

	defer assert.Nil(t, writeDatabaseDocker.Stop())
	defer assert.Nil(t, file.Remove(dbDatabase1))
}

func TestTablePrefixAndSingular(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	mysqlDocker := NewMysqlDocker(testDatabaseDocker)
	mysqlQuery, err := mysqlDocker.NewWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	postgresqlDocker := NewPostgresqlDocker(testDatabaseDocker)
	postgresqlQuery, err := postgresqlDocker.NewWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	sqliteDocker := NewSqliteDocker(dbDatabase)
	sqliteDB, err := sqliteDocker.NewWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverDocker := NewSqlserverDocker(testDatabaseDocker)
	sqlserverDB, err := sqlserverDocker.NewWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init sqlserver error: %s", err)
	}

	dbs := map[ormcontract.Driver]ormcontract.Query{
		ormcontract.DriverMysql:      mysqlQuery,
		ormcontract.DriverPostgresql: postgresqlQuery,
		ormcontract.DriverSqlite:     sqliteDB,
		ormcontract.DriverSqlserver:  sqlserverDB,
	}

	for drive, db := range dbs {
		t.Run(drive.String(), func(t *testing.T) {
			user := User{Name: "user"}
			assert.Nil(t, db.Create(&user))
			assert.True(t, user.ID > 0)

			var user1 User
			assert.Nil(t, db.Find(&user1, user.ID))
			assert.True(t, user1.ID > 0)
		})
	}
}

func paginator(page string, limit string) func(methods ormcontract.Query) ormcontract.Query {
	return func(query ormcontract.Query) ormcontract.Query {
		page, _ := strconv.Atoi(page)
		limit, _ := strconv.Atoi(limit)
		offset := (page - 1) * limit

		return query.Offset(offset).Limit(limit)
	}
}

func mockDummyConnection(mockConfig *configmocks.Config, databaseConfig contractstesting.DatabaseConfig) {
	mockConfig.On("GetString", "database.connections.dummy.prefix").Return("")
	mockConfig.On("GetBool", "database.connections.dummy.singular").Return(false)
	mockConfig.On("Get", "database.connections.dummy.read").Return(nil)
	mockConfig.On("Get", "database.connections.dummy.write").Return(nil)
	mockConfig.On("GetString", "database.connections.dummy.host").Return("127.0.0.1")
	mockConfig.On("GetString", "database.connections.dummy.username").Return(databaseConfig.Username)
	mockConfig.On("GetString", "database.connections.dummy.password").Return(databaseConfig.Password)
	mockConfig.On("GetInt", "database.connections.dummy.port").Return(databaseConfig.Port)
	mockConfig.On("GetString", "database.connections.dummy.driver").Return(ormcontract.DriverMysql.String())
	mockConfig.On("GetString", "database.connections.dummy.charset").Return("utf8mb4")
	mockConfig.On("GetString", "database.connections.dummy.loc").Return("Local")
	mockConfig.On("GetString", "database.connections.dummy.database").Return(databaseConfig.Database)
}

func mockPostgresqlConnection(mockConfig *configmocks.Config, databaseConfig contractstesting.DatabaseConfig) {
	mockConfig.On("GetString", "database.connections.postgresql.prefix").Return("")
	mockConfig.On("GetBool", "database.connections.postgresql.singular").Return(false)
	mockConfig.On("Get", "database.connections.postgresql.read").Return(nil)
	mockConfig.On("Get", "database.connections.postgresql.write").Return(nil)
	mockConfig.On("GetString", "database.connections.postgresql.host").Return("127.0.0.1")
	mockConfig.On("GetString", "database.connections.postgresql.username").Return(databaseConfig.Username)
	mockConfig.On("GetString", "database.connections.postgresql.password").Return(databaseConfig.Password)
	mockConfig.On("GetInt", "database.connections.postgresql.port").Return(databaseConfig.Port)
	mockConfig.On("GetString", "database.connections.postgresql.driver").Return(ormcontract.DriverPostgresql.String())
	mockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	mockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	mockConfig.On("GetString", "database.connections.postgresql.database").Return(databaseConfig.Database)
}

type UserObserver struct{}

func (u *UserObserver) Retrieved(event ormcontract.Event) error {
	return errors.New("retrieved")
}

func (u *UserObserver) Creating(event ormcontract.Event) error {
	return errors.New("creating")
}

func (u *UserObserver) Created(event ormcontract.Event) error {
	return errors.New("created")
}

func (u *UserObserver) Updating(event ormcontract.Event) error {
	return errors.New("updating")
}

func (u *UserObserver) Updated(event ormcontract.Event) error {
	return errors.New("updated")
}

func (u *UserObserver) Saving(event ormcontract.Event) error {
	return errors.New("saving")
}

func (u *UserObserver) Saved(event ormcontract.Event) error {
	return errors.New("saved")
}

func (u *UserObserver) Deleting(event ormcontract.Event) error {
	return errors.New("deleting")
}

func (u *UserObserver) Deleted(event ormcontract.Event) error {
	return errors.New("deleted")
}

func (u *UserObserver) ForceDeleting(event ormcontract.Event) error {
	return errors.New("forceDeleting")
}

func (u *UserObserver) ForceDeleted(event ormcontract.Event) error {
	return errors.New("forceDeleted")
}
