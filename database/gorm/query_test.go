package gorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	_ "gorm.io/driver/postgres"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
)

type contextKey int

const testContextKey contextKey = 0

type User struct {
	orm.Model
	orm.SoftDeletes
	Name    string
	Avatar  string
	Address *Address
	Books   []*Book
	House   *House   `gorm:"polymorphic:Houseable"`
	Phones  []*Phone `gorm:"polymorphic:Phoneable"`
	Roles   []*Role  `gorm:"many2many:role_user"`
	age     int
}

func (u *User) DispatchesEvents() map[ormcontract.EventType]func(ormcontract.Event) error {
	return map[ormcontract.EventType]func(ormcontract.Event) error{
		ormcontract.EventCreating: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_creating_name" {
					event.SetAttribute("avatar", "event_creating_avatar")
				}
				if name.(string) == "event_creating_FirstOrCreate_name" {
					event.SetAttribute("avatar", "event_creating_FirstOrCreate_avatar")
				}
				if name.(string) == "event_creating_IsDirty_name" {
					if event.IsDirty("name") {
						event.SetAttribute("avatar", "event_creating_IsDirty_avatar")
					}
				}
				if name.(string) == "event_context" {
					val := event.Context().Value(testContextKey)
					event.SetAttribute("avatar", val.(string))
				}
				if name.(string) == "event_query" {
					_ = event.Query().Create(&User{Name: "event_query1"})
				}
			}

			return nil
		},
		ormcontract.EventCreated: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_created_name" {
					event.SetAttribute("avatar", "event_created_avatar")
				}
				if name.(string) == "event_created_FirstOrCreate_name" {
					event.SetAttribute("avatar", "event_created_FirstOrCreate_avatar")
				}
			}

			return nil
		},
		ormcontract.EventSaving: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_saving_create_name" {
					event.SetAttribute("avatar", "event_saving_create_avatar")
				}
				if name.(string) == "event_saving_save_name" {
					event.SetAttribute("avatar", "event_saving_save_avatar")
				}
				if name.(string) == "event_saving_FirstOrCreate_name" {
					event.SetAttribute("avatar", "event_saving_FirstOrCreate_avatar")
				}
				if name.(string) == "event_save_without_name" {
					event.SetAttribute("avatar", "event_save_without_avatar")
				}
				if name.(string) == "event_save_quietly_name" {
					event.SetAttribute("avatar", "event_save_quietly_avatar")
				}
				if name.(string) == "event_saving_IsDirty_name" {
					if event.IsDirty("name") {
						event.SetAttribute("avatar", "event_saving_IsDirty_avatar")
					}
				}
			}

			avatar := event.GetAttribute("avatar")
			if avatar != nil && avatar.(string) == "event_saving_single_update_avatar" {
				event.SetAttribute("avatar", "event_saving_single_update_avatar1")
			}

			return nil
		},
		ormcontract.EventSaved: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_saved_create_name" {
					event.SetAttribute("avatar", "event_saved_create_avatar")
				}
				if name.(string) == "event_saved_save_name" {
					event.SetAttribute("avatar", "event_saved_save_avatar")
				}
				if name.(string) == "event_saved_FirstOrCreate_name" {
					event.SetAttribute("avatar", "event_saved_FirstOrCreate_avatar")
				}
				if name.(string) == "event_save_without_name" {
					event.SetAttribute("avatar", "event_saved_without_avatar")
				}
				if name.(string) == "event_save_quietly_name" {
					event.SetAttribute("avatar", "event_saved_quietly_avatar")
				}
			}

			avatar := event.GetAttribute("avatar")
			if avatar != nil && avatar.(string) == "event_saved_map_update_avatar" {
				event.SetAttribute("avatar", "event_saved_map_update_avatar1")
			}

			return nil
		},
		ormcontract.EventUpdating: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_updating_create_name" {
					event.SetAttribute("avatar", "event_updating_create_avatar")
				}
				if name.(string) == "event_updating_save_name" {
					event.SetAttribute("avatar", "event_updating_save_avatar")
				}
				if name.(string) == "event_updating_single_update_IsDirty_name1" {
					if event.IsDirty("name") {
						name := event.GetAttribute("name")
						if name != "event_updating_single_update_IsDirty_name1" {
							return errors.New("error")
						}

						event.SetAttribute("avatar", "event_updating_single_update_IsDirty_avatar")
					}
				}
				if name.(string) == "event_updating_map_update_IsDirty_name1" {
					if event.IsDirty("name") {
						name := event.GetAttribute("name")
						if name != "event_updating_map_update_IsDirty_name1" {
							return errors.New("error")
						}

						event.SetAttribute("avatar", "event_updating_map_update_IsDirty_avatar")
					}
				}
				if name.(string) == "event_updating_model_update_IsDirty_name1" {
					if event.IsDirty("name") {
						name := event.GetAttribute("name")
						if name != "event_updating_model_update_IsDirty_name1" {
							return errors.New("error")
						}
						event.SetAttribute("avatar", "event_updating_model_update_IsDirty_avatar")
					}
				}
			}

			avatar := event.GetAttribute("avatar")
			if avatar != nil {
				if avatar.(string) == "event_updating_save_avatar" {
					event.SetAttribute("avatar", "event_updating_save_avatar1")
				}
				if avatar.(string) == "event_updating_model_update_avatar" {
					event.SetAttribute("avatar", "event_updating_model_update_avatar1")
				}
			}

			return nil
		},
		ormcontract.EventUpdated: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil {
				if name.(string) == "event_updated_create_name" {
					event.SetAttribute("avatar", "event_updated_create_avatar")
				}
				if name.(string) == "event_updated_save_name" {
					event.SetAttribute("avatar", "event_updated_save_avatar")
				}
			}

			avatar := event.GetAttribute("avatar")
			if avatar != nil {
				if avatar.(string) == "event_updated_save_avatar" {
					event.SetAttribute("avatar", "event_updated_save_avatar1")
				}
				if avatar.(string) == "event_updated_model_update_avatar" {
					event.SetAttribute("avatar", "event_updated_model_update_avatar1")
				}
			}

			return nil
		},
		ormcontract.EventDeleting: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil && name.(string) == "event_deleting_name" {
				return errors.New("deleting error")
			}

			return nil
		},
		ormcontract.EventDeleted: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil && name.(string) == "event_deleted_name" {
				return errors.New("deleted error")
			}

			return nil
		},
		ormcontract.EventForceDeleting: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil && name.(string) == "event_force_deleting_name" {
				return errors.New("force deleting error")
			}

			return nil
		},
		ormcontract.EventForceDeleted: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil && name.(string) == "event_force_deleted_name" {
				return errors.New("force deleted error")
			}

			return nil
		},
		ormcontract.EventRetrieved: func(event ormcontract.Event) error {
			name := event.GetAttribute("name")
			if name != nil && name.(string) == "event_retrieved_name" {
				event.SetAttribute("name", "event_retrieved_name1")
			}

			return nil
		},
	}
}

type Role struct {
	orm.Model
	Name  string
	Users []*User `gorm:"many2many:role_user"`
}

type Address struct {
	orm.Model
	UserID   uint
	Name     string
	Province string
	User     *User
}

type Book struct {
	orm.Model
	UserID uint
	Name   string
	User   *User
	Author *Author
}

type Author struct {
	orm.Model
	BookID uint
	Name   string
}

type House struct {
	orm.Model
	Name          string
	HouseableID   uint
	HouseableType string
}

func (h *House) Factory() string {
	return "house"
}

type Phone struct {
	orm.Model
	Name          string
	PhoneableID   uint
	PhoneableType string
}

type Product struct {
	orm.Model
	orm.SoftDeletes
	Name string
}

func (p *Product) Connection() string {
	return "postgresql"
}

type Review struct {
	orm.Model
	orm.SoftDeletes
	Body string
}

func (r *Review) Connection() string {
	return ""
}

type Person struct {
	orm.Model
	orm.SoftDeletes
	Name string
}

func (p *Person) Connection() string {
	return "dummy"
}

type QueryTestSuite struct {
	suite.Suite
	queries map[ormcontract.Driver]ormcontract.Query
}

func TestQueryTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	testContext = context.Background()
	testContext = context.WithValue(testContext, testContextKey, "goravel")

	mysqlDocker := NewMysqlDocker()
	mysqlPool, mysqlResource, mysqlQuery, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	postgresqlDocker := NewPostgresqlDocker()
	postgresqlPool, postgresqlResource, postgresqlQuery, err := postgresqlDocker.New()
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	sqliteDocker := NewSqliteDocker(dbDatabase)
	_, _, sqliteQuery, err := sqliteDocker.New()
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverDocker := NewSqlserverDocker()
	sqlserverPool, sqlserverResource, sqlserverQuery, err := sqlserverDocker.New()
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
	})

	assert.Nil(t, file.Remove(dbDatabase))
	assert.Nil(t, mysqlPool.Purge(mysqlResource))
	assert.Nil(t, postgresqlPool.Purge(postgresqlResource))
	assert.Nil(t, sqlserverPool.Purge(sqlserverResource))
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
					user := &User{
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
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
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
					s.Equal("event_created_avatar", user.Avatar)

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
					s.Equal("event_created_FirstOrCreate_avatar", user.Avatar)

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
					s.Equal("event_updating_model_update_avatar1", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_updating_model_update_name", user1.Name)
					s.Equal("event_updating_model_update_avatar1", user1.Avatar)
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

func (s *QueryTestSuite) TestFind() {
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
					s.Nil(query.Find(&user2, user.ID))
					s.True(user2.ID > 0)

					var user3 []User
					s.Nil(query.Find(&user3, []uint{user.ID}))
					s.Equal(1, len(user3))

					var user4 []User
					s.Nil(query.Where("id in ?", []uint{user.ID}).Find(&user4))
					s.Equal(1, len(user4))
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
	for _, query := range s.queries {
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "first_user"}
					s.Nil(query.Create(&user))
					s.True(user.ID > 0)

					var user1 User
					s.Nil(query.Where("name", "first_user").First(&user1))
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
		})
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
			var total int64
			s.Nil(query.Where("name = ?", "paginate_user").Paginate(1, 3, &users, nil))
			s.Equal(3, len(users))

			s.Nil(query.Where("name = ?", "paginate_user").Paginate(2, 3, &users, &total))
			s.Equal(1, len(users))
			s.Equal(int64(4), total)

			s.Nil(query.Model(User{}).Where("name = ?", "paginate_user").Paginate(1, 3, &users, &total))
			s.Equal(3, len(users))
			s.Equal(int64(4), total)

			s.Nil(query.Table("users").Where("name = ?", "paginate_user").Paginate(1, 3, &users, &total))
			s.Equal(3, len(users))
			s.Equal(int64(4), total)
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
	for driver, query := range s.queries {
		s.Run(driver.String(), func() {
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
				test.setup(test.description)
			}
		})
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
		tests := []struct {
			name  string
			setup func()
		}{
			{
				name: "success",
				setup: func() {
					user := User{Name: "event_save_quietly_name", Avatar: "save_quietly_avatar"}
					s.Nil(query.SaveQuietly(&user))
					s.True(user.ID > 0)
					s.Equal("event_save_quietly_name", user.Name)
					s.Equal("save_quietly_avatar", user.Avatar)

					var user1 User
					s.Nil(query.Find(&user1, user.ID))
					s.Equal("event_save_quietly_name", user1.Name)
					s.Equal("save_quietly_avatar", user1.Avatar)
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
			err = query.Where("name", "update_or_create_user").UpdateOrCreate(&user2, User{Name: "update_or_create_user"}, User{Avatar: "update_or_create_avatar1"})
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
			s.True(len(user2) > 0)

			var user3 User
			s.Nil(query.Where("name = 'where_user'").Find(&user3))
			s.True(user3.ID > 0)

			var user4 User
			s.Nil(query.Where("name", "where_user").Find(&user4))
			s.True(user4.ID > 0)
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

func TestCustomConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	mysqlDocker := NewMysqlDocker()
	mysqlPool, mysqlResource, query, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}
	postgresqlDocker := NewPostgresqlDocker()
	postgresqlPool, postgresqlResource, _, err := postgresqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	review := Review{Body: "create_review"}
	assert.Nil(t, query.Create(&review))
	assert.True(t, review.ID > 0)

	var review1 Review
	assert.Nil(t, query.Where("body", "create_review").First(&review1))
	assert.True(t, review1.ID > 0)

	mysqlDocker.MockConfig.On("Get", "database.connections.postgresql.read").Return(nil)
	mysqlDocker.MockConfig.On("Get", "database.connections.postgresql.write").Return(nil)
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.host").Return("localhost")
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.username").Return(DbUser)
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.password").Return(DbPassword)
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.driver").Return(ormcontract.DriverPostgresql.String())
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.database").Return("postgres")
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	mysqlDocker.MockConfig.On("GetString", "database.connections.postgresql.prefix").Return("")
	mysqlDocker.MockConfig.On("GetBool", "database.connections.postgresql.singular").Return(false)
	mysqlDocker.MockConfig.On("GetInt", "database.connections.postgresql.port").Return(cast.ToInt(postgresqlResource.GetPort("5432/tcp")))

	product := Product{Name: "create_product"}
	assert.Nil(t, query.Create(&product))
	assert.True(t, product.ID > 0)

	var product1 Product
	assert.Nil(t, query.Where("name", "create_product").First(&product1))
	assert.True(t, product1.ID > 0)

	var product2 Product
	assert.Nil(t, query.Where("name", "create_product1").First(&product2))
	assert.True(t, product2.ID == 0)

	mysqlDocker.MockConfig.On("GetString", "database.connections.dummy.driver").Return("")

	person := Person{Name: "create_person"}
	assert.NotNil(t, query.Create(&person))
	assert.True(t, person.ID == 0)

	assert.Nil(t, mysqlPool.Purge(mysqlResource))
	assert.Nil(t, postgresqlPool.Purge(postgresqlResource))
}

func TestReadWriteSeparate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	readMysqlDocker := NewMysqlDocker()
	readMysqlPool, readMysqlResource, readMysqlQuery, err := readMysqlDocker.New()
	if err != nil {
		log.Fatalf("Get read mysql error: %s", err)
	}

	writeMysqlDocker := NewMysqlDocker()
	writeMysqlPool, writeMysqlResource, writeMysqlQuery, err := writeMysqlDocker.New()
	if err != nil {
		log.Fatalf("Get write mysql error: %s", err)
	}

	writeMysqlDocker.MockReadWrite(readMysqlDocker.Port, writeMysqlDocker.Port)
	mysqlQuery, err := writeMysqlDocker.Query(false)
	if err != nil {
		log.Fatalf("Get mysql gorm error: %s", err)
	}

	readPostgresqlDocker := NewPostgresqlDocker()
	readPostgresqlPool, readPostgresqlResource, readPostgresqlQuery, err := readPostgresqlDocker.New()
	if err != nil {
		log.Fatalf("Get read postgresql error: %s", err)
	}

	writePostgresqlDocker := NewPostgresqlDocker()
	writePostgresqlPool, writePostgresqlResource, writePostgresqlQuery, err := writePostgresqlDocker.New()
	if err != nil {
		log.Fatalf("Get write postgresql error: %s", err)
	}

	writePostgresqlDocker.MockReadWrite(readPostgresqlDocker.Port, writePostgresqlDocker.Port)
	postgresqlQuery, err := writePostgresqlDocker.Query(false)
	if err != nil {
		log.Fatalf("Get postgresql gorm error: %s", err)
	}

	readSqliteDocker := NewSqliteDocker(dbDatabase)
	_, _, readSqliteQuery, err := readSqliteDocker.New()
	if err != nil {
		log.Fatalf("Get read sqlite error: %s", err)
	}

	writeSqliteDocker := NewSqliteDocker(dbDatabase1)
	_, _, writeSqliteQuery, err := writeSqliteDocker.New()
	if err != nil {
		log.Fatalf("Get write sqlite error: %s", err)
	}

	writeSqliteDocker.MockReadWrite()
	sqliteDB, err := writeSqliteDocker.Query(false)
	if err != nil {
		log.Fatalf("Get sqlite gorm error: %s", err)
	}

	readSqlserverDocker := NewSqlserverDocker()
	readSqlserverPool, readSqlserverResource, readSqlserverQuery, err := readSqlserverDocker.New()
	if err != nil {
		log.Fatalf("Get read sqlserver error: %s", err)
	}

	writeSqlserverDocker := NewSqlserverDocker()
	writeSqlserverPool, writeSqlserverResource, writeSqlserverQuery, err := writeSqlserverDocker.New()
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

	assert.Nil(t, file.Remove(dbDatabase))
	assert.Nil(t, file.Remove(dbDatabase1))

	if err := readMysqlPool.Purge(readMysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writeMysqlPool.Purge(writeMysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := readPostgresqlPool.Purge(readPostgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writePostgresqlPool.Purge(writePostgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := readSqlserverPool.Purge(readSqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := writeSqlserverPool.Purge(writeSqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func TestTablePrefixAndSingular(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	mysqlDocker := NewMysqlDocker()
	mysqlPool, mysqlResource, err := mysqlDocker.Init()
	if err != nil {
		log.Fatalf("Init mysql docker error: %s", err)
	}
	mysqlDocker.mockWithPrefixAndSingular()
	mysqlQuery, err := mysqlDocker.QueryWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	postgresqlDocker := NewPostgresqlDocker()
	postgresqlPool, postgresqlResource, err := postgresqlDocker.Init()
	if err != nil {
		log.Fatalf("Init postgresql docker error: %s", err)
	}
	postgresqlDocker.mockWithPrefixAndSingular()
	postgresqlQuery, err := postgresqlDocker.QueryWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init postgresql error: %s", err)
	}

	sqliteDocker := NewSqliteDocker(dbDatabase)
	_, _, err = sqliteDocker.Init()
	if err != nil {
		log.Fatalf("Init sqlite docker error: %s", err)
	}
	sqliteDocker.mockWithPrefixAndSingular()
	sqliteDB, err := sqliteDocker.QueryWithPrefixAndSingular()
	if err != nil {
		log.Fatalf("Init sqlite error: %s", err)
	}

	sqlserverDocker := NewSqlserverDocker()
	sqlserverPool, sqlserverResource, err := sqlserverDocker.Init()
	if err != nil {
		log.Fatalf("Init sqlserver docker error: %s", err)
	}
	sqlserverDocker.mockWithPrefixAndSingular()
	sqlserverDB, err := sqlserverDocker.QueryWithPrefixAndSingular()
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

	assert.Nil(t, file.Remove(dbDatabase))

	if err := mysqlPool.Purge(mysqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := postgresqlPool.Purge(postgresqlResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := sqlserverPool.Purge(sqlserverResource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
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
