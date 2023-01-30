package gorm

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
	_ "gorm.io/driver/postgres"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
)

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

type Phone struct {
	orm.Model
	Name          string
	PhoneableID   uint
	PhoneableType string
}

type GormQueryTestSuite struct {
	suite.Suite
	dbs []ormcontract.DB
}

func TestGormQueryTestSuite(t *testing.T) {
	mysqlPool, mysqlDocker, mysqlDB, err := MysqlDocker()
	if err != nil {
		log.Fatalf("Get gorm mysql error: %s", err)
	}

	postgresqlPool, postgresqlDocker, postgresqlDB, err := PostgresqlDocker()
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

	_, _, sqliteDB, err := SqliteDocker()
	if err != nil {
		log.Fatalf("Get gorm sqlite error: %s", err)
	}

	sqlserverPool, sqlserverDocker, sqlserverDB, err := SqlserverDocker()
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

	suite.Run(t, &GormQueryTestSuite{
		dbs: []ormcontract.DB{
			mysqlDB,
			postgresqlDB,
			sqliteDB,
			sqlserverDB,
		},
	})

	file.Remove("goravel")

	if err := mysqlPool.Purge(mysqlDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := postgresqlPool.Purge(postgresqlDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := sqlserverPool.Purge(sqlserverDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *GormQueryTestSuite) SetupTest() {
}

func (s *GormQueryTestSuite) TestAssociation() {
	for _, db := range s.dbs {
		tests := []struct {
			description string
			setup       func(description string)
		}{
			{
				description: "Find",
				setup: func(description string) {
					user := &User{
						Name: "association_find_name",
						Address: &Address{
							Name: "association_find_address",
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)

					var userAddress Address
					s.Nil(db.Model(&user1).Association("Address").Find(&userAddress), description)
					s.True(userAddress.ID > 0, description)
					s.Equal("association_find_address", userAddress.Name, description)
				},
			},
			{
				description: "hasOne Append",
				setup: func(description string) {
					user := &User{
						Name: "association_has_one_append_name",
						Address: &Address{
							Name: "association_has_one_append_address",
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Address").Append(&Address{Name: "association_has_one_append_address1"}), description)

					s.Nil(db.Load(&user1, "Address"), description)
					s.True(user1.Address.ID > 0, description)
					s.Equal("association_has_one_append_address1", user1.Address.Name, description)
				},
			},
			{
				description: "hasMany Append",
				setup: func(description string) {
					user := &User{
						Name: "association_has_many_append_name",
						Books: []*Book{
							{Name: "association_has_many_append_address1"},
							{Name: "association_has_many_append_address2"},
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Books[0].ID > 0, description)
					s.True(user.Books[1].ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Books").Append(&Book{Name: "association_has_many_append_address3"}), description)

					s.Nil(db.Load(&user1, "Books"), description)
					s.Equal(3, len(user1.Books), description)
					s.Equal("association_has_many_append_address3", user1.Books[2].Name, description)
				},
			},
			{
				description: "hasOne Replace",
				setup: func(description string) {
					user := &User{
						Name: "association_has_one_append_name",
						Address: &Address{
							Name: "association_has_one_append_address",
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Address").Replace(&Address{Name: "association_has_one_append_address1"}), description)

					s.Nil(db.Load(&user1, "Address"), description)
					s.True(user1.Address.ID > 0, description)
					s.Equal("association_has_one_append_address1", user1.Address.Name, description)
				},
			},
			{
				description: "hasMany Replace",
				setup: func(description string) {
					user := &User{
						Name: "association_has_many_replace_name",
						Books: []*Book{
							{Name: "association_has_many_replace_address1"},
							{Name: "association_has_many_replace_address2"},
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Books[0].ID > 0, description)
					s.True(user.Books[1].ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Books").Replace(&Book{Name: "association_has_many_replace_address3"}), description)

					s.Nil(db.Load(&user1, "Books"), description)
					s.Equal(1, len(user1.Books), description)
					s.Equal("association_has_many_replace_address3", user1.Books[0].Name, description)
				},
			},
			{
				description: "Delete",
				setup: func(description string) {
					user := &User{
						Name: "association_delete_name",
						Address: &Address{
							Name: "association_delete_address",
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)

					// No ID when Delete
					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Address").Delete(&Address{Name: "association_delete_address"}), description)

					s.Nil(db.Load(&user1, "Address"), description)
					s.True(user1.Address.ID > 0, description)
					s.Equal("association_delete_address", user1.Address.Name, description)

					// Has ID when Delete
					var user2 User
					s.Nil(db.Find(&user2, user.ID), description)
					s.True(user2.ID > 0, description)
					var userAddress Address
					userAddress.ID = user1.Address.ID
					s.Nil(db.Model(&user2).Association("Address").Delete(&userAddress), description)

					s.Nil(db.Load(&user2, "Address"), description)
					s.Nil(user2.Address, description)
				},
			},
			{
				description: "Clear",
				setup: func(description string) {
					user := &User{
						Name: "association_clear_name",
						Address: &Address{
							Name: "association_clear_address",
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)

					// No ID when Delete
					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Nil(db.Model(&user1).Association("Address").Clear(), description)

					s.Nil(db.Load(&user1, "Address"), description)
					s.Nil(user1.Address, description)
				},
			},
			{
				description: "Count",
				setup: func(description string) {
					user := &User{
						Name: "association_count_name",
						Books: []*Book{
							{Name: "association_count_address1"},
							{Name: "association_count_address2"},
						},
					}

					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Books[0].ID > 0, description)
					s.True(user.Books[1].ID > 0, description)

					var user1 User
					s.Nil(db.Find(&user1, user.ID), description)
					s.True(user1.ID > 0, description)
					s.Equal(int64(2), db.Model(&user1).Association("Books").Count(), description)
				},
			},
		}

		for _, test := range tests {
			test.setup(test.description)
		}
	}
}

func (s *GormQueryTestSuite) TestCount() {
	for _, db := range s.dbs {
		user := User{Name: "count_user", Avatar: "count_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "count_user", Avatar: "count_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var count int64
		s.Nil(db.Model(&User{}).Where("name = ?", "count_user").Count(&count))
		s.True(count > 0)

		var count1 int64
		s.Nil(db.Table("users").Where("name = ?", "count_user").Count(&count1))
		s.True(count1 > 0)
	}
}

func (s *GormQueryTestSuite) TestCreate() {
	for _, db := range s.dbs {
		tests := []struct {
			description string
			setup       func(description string)
		}{
			{
				description: "success when create with no relationships",
				setup: func(description string) {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(db.Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID == 0, description)
					s.True(user.Books[0].ID == 0, description)
					s.True(user.Books[1].ID == 0, description)
				},
			},
			{
				description: "success when create with select orm.Associations",
				setup: func(description string) {
					user := User{Name: "create_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(db.Select(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)
					s.True(user.Books[0].ID > 0, description)
					s.True(user.Books[1].ID > 0, description)
				},
			},
			{
				description: "success when create with select fields",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(db.Select("Name", "Avatar", "Address").Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID > 0, description)
					s.True(user.Books[0].ID == 0, description)
					s.True(user.Books[1].ID == 0, description)
				},
			},
			{
				description: "success when create with omit fields",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(db.Omit("Address").Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID == 0, description)
					s.True(user.Books[0].ID > 0, description)
					s.True(user.Books[1].ID > 0, description)
				},
			},
			{
				description: "success create with omit orm.Associations",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.Nil(db.Omit(orm.Associations).Create(&user), description)
					s.True(user.ID > 0, description)
					s.True(user.Address.ID == 0, description)
					s.True(user.Books[0].ID == 0, description)
					s.True(user.Books[1].ID == 0, description)
				},
			},
			{
				description: "error when set select and omit at the same time",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(db.Omit(orm.Associations).Select("Name").Create(&user), "cannot set Select and Omits at the same time", description)
				},
			},
			{
				description: "error when select that set fields and orm.Associations at the same time",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(db.Select("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time", description)
				},
			},
			{
				description: "error when omit that set fields and orm.Associations at the same time",
				setup: func(description string) {
					user := User{Name: "create_user", Avatar: "create_avatar", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
					user.Address.Name = "create_address"
					user.Books[0].Name = "create_book0"
					user.Books[1].Name = "create_book1"
					s.EqualError(db.Omit("Name", orm.Associations).Create(&user), "cannot set orm.Associations and other fields at the same time", description)
				},
			},
		}
		for _, test := range tests {
			test.setup(test.description)
		}
	}
}

func (s *GormQueryTestSuite) TestDelete() {
	for _, db := range s.dbs {
		user := User{Name: "delete_user", Avatar: "delete_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		s.Nil(db.Delete(&user))

		var user1 User
		s.Nil(db.Find(&user1, user.ID))
		s.Equal(uint(0), user1.ID)

		user2 := User{Name: "delete_user", Avatar: "delete_avatar"}
		s.Nil(db.Create(&user2))
		s.True(user2.ID > 0)

		s.Nil(db.Delete(&User{}, user2.ID))

		var user3 User
		s.Nil(db.Find(&user3, user2.ID))
		s.Equal(uint(0), user3.ID)

		users := []User{{Name: "delete_user", Avatar: "delete_avatar"}, {Name: "delete_user1", Avatar: "delete_avatar1"}}
		s.Nil(db.Create(&users))
		s.True(users[0].ID > 0)
		s.True(users[1].ID > 0)

		s.Nil(db.Delete(&User{}, []uint{users[0].ID, users[1].ID}))

		var count int64
		s.Nil(db.Model(&User{}).Where("name", "delete_user").OrWhere("name", "delete_user1").Count(&count))
		s.True(count == 0)
	}
}

func (s *GormQueryTestSuite) TestDistinct() {
	for _, db := range s.dbs {
		user := User{Name: "distinct_user", Avatar: "distinct_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "distinct_user", Avatar: "distinct_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var users []User
		s.Nil(db.Distinct("name").Find(&users, []uint{user.ID, user1.ID}))
		s.Equal(1, len(users))
	}
}

func (s *GormQueryTestSuite) TestFirst() {
	for _, db := range s.dbs {
		user := User{Name: "first_user"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		var user1 User
		s.Nil(db.Where("name = ?", "first_user").First(&user1))
		s.True(user1.ID > 0)
	}
}

func (s *GormQueryTestSuite) TestFirstOrCreate() {
	for _, db := range s.dbs {
		var user User
		s.Nil(db.Where("avatar = ?", "first_or_create_avatar").FirstOrCreate(&user, User{Name: "first_or_create_user"}))
		s.True(user.ID > 0)

		var user1 User
		s.Nil(db.Where("avatar = ?", "first_or_create_avatar").FirstOrCreate(&user1, User{Name: "user"}, User{Avatar: "first_or_create_avatar1"}))
		s.True(user1.ID > 0)
		s.True(user1.Avatar == "first_or_create_avatar1")
	}
}

func (s *GormQueryTestSuite) TestJoin() {
	for _, db := range s.dbs {
		user := User{Name: "join_user", Avatar: "join_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		userAddress := Address{UserID: user.ID, Name: "join_address", Province: "join_province"}
		s.Nil(db.Create(&userAddress))
		s.True(userAddress.ID > 0)

		type Result struct {
			UserName        string
			UserAddressName string
		}
		var result []Result
		s.Nil(db.Model(&User{}).Where("users.id = ?", user.ID).Join("left join addresses ua on users.id = ua.user_id").
			Select("users.name user_name, ua.name user_address_name").Get(&result))
		s.Equal(1, len(result))
		s.Equal("join_user", result[0].UserName)
		s.Equal("join_address", result[0].UserAddressName)
	}
}

func (s *GormQueryTestSuite) TestOffset() {
	for _, db := range s.dbs {
		user := User{Name: "offset_user", Avatar: "offset_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "offset_user", Avatar: "offset_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var user2 []User
		s.Nil(db.Where("name = ?", "offset_user").Offset(1).Limit(1).Get(&user2))
		s.True(len(user2) > 0)
		s.True(user2[0].ID > 0)
	}
}

func (s *GormQueryTestSuite) TestOmit() {
	// todo
}

func (s *GormQueryTestSuite) TestOrder() {
	for _, db := range s.dbs {
		user := User{Name: "order_user", Avatar: "order_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "order_user", Avatar: "order_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var user2 []User
		s.Nil(db.Where("name = ?", "order_user").Order("id desc").Order("name asc").Get(&user2))
		s.True(len(user2) > 0)
		s.True(user2[0].ID > 0)
	}
}

func (s *GormQueryTestSuite) TestPluck() {
	for _, db := range s.dbs {
		user := User{Name: "pluck_user", Avatar: "pluck_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "pluck_user", Avatar: "pluck_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var avatars []string
		s.Nil(db.Model(&User{}).Where("name = ?", "pluck_user").Pluck("avatar", &avatars))
		s.True(len(avatars) > 0)
		s.True(avatars[0] == "pluck_avatar")
	}
}

func (s *GormQueryTestSuite) TestHasOne() {
	for _, db := range s.dbs {
		user := &User{
			Name: "has_one_name",
			Address: &Address{
				Name: "has_one_address",
			},
		}

		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Address.ID > 0)

		var user1 User
		s.Nil(db.With("Address").Where("name = ?", "has_one_name").First(&user1))
		s.True(user.ID > 0)
		s.True(user.Address.ID > 0)
	}
}

func (s *GormQueryTestSuite) TestHasOneMorph() {
	for _, db := range s.dbs {
		user := &User{
			Name: "has_one_morph_name",
			House: &House{
				Name: "has_one_morph_house",
			},
		}
		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.House.ID > 0)

		var user1 User
		s.Nil(db.With("House").Where("name = ?", "has_one_morph_name").First(&user1))
		s.True(user.ID > 0)
		s.True(user.Name == "has_one_morph_name")
		s.True(user.House.ID > 0)
		s.True(user.House.Name == "has_one_morph_house")

		var house House
		s.Nil(db.Where("name = ?", "has_one_morph_house").Where("houseable_type = ?", "users").Where("houseable_id = ?", user.ID).First(&house))
		s.True(house.ID > 0)
	}
}

func (s *GormQueryTestSuite) TestHasMany() {
	for _, db := range s.dbs {
		user := &User{
			Name: "has_many_name",
			Books: []*Book{
				{Name: "has_many_book1"},
				{Name: "has_many_book2"},
			},
		}

		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Books[0].ID > 0)
		s.True(user.Books[1].ID > 0)

		var user1 User
		s.Nil(db.With("Books").Where("name = ?", "has_many_name").First(&user1))
		s.True(user.ID > 0)
		s.True(len(user.Books) == 2)
	}
}

func (s *GormQueryTestSuite) TestHasManyMorph() {
	for _, db := range s.dbs {
		user := &User{
			Name: "has_many_morph_name",
			Phones: []*Phone{
				{Name: "has_many_morph_phone1"},
				{Name: "has_many_morph_phone2"},
			},
		}
		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Phones[0].ID > 0)
		s.True(user.Phones[1].ID > 0)

		var user1 User
		s.Nil(db.With("Phones").Where("name = ?", "has_many_morph_name").First(&user1))
		s.True(user.ID > 0)
		s.True(user.Name == "has_many_morph_name")
		s.True(len(user.Phones) == 2)
		s.True(user.Phones[0].Name == "has_many_morph_phone1")
		s.True(user.Phones[1].Name == "has_many_morph_phone2")

		var phones []Phone
		s.Nil(db.Where("name like ?", "has_many_morph_phone%").Where("phoneable_type = ?", "users").Where("phoneable_id = ?", user.ID).Find(&phones))
		s.True(len(phones) == 2)
	}
}

func (s *GormQueryTestSuite) TestBelongsTo() {
	for _, db := range s.dbs {
		user := &User{
			Name: "belongs_to_name",
			Address: &Address{
				Name: "belongs_to_address",
			},
		}

		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Address.ID > 0)

		var userAddress Address
		s.Nil(db.With("User").Where("name = ?", "belongs_to_address").First(&userAddress))
		s.True(userAddress.ID > 0)
		s.True(userAddress.User.ID > 0)
	}
}

func (s *GormQueryTestSuite) TestManyToMany() {
	for _, db := range s.dbs {
		user := &User{
			Name: "many_to_many_name",
			Roles: []*Role{
				{Name: "many_to_many_role1"},
				{Name: "many_to_many_role2"},
			},
		}

		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Roles[0].ID > 0)
		s.True(user.Roles[1].ID > 0)

		var user1 User
		s.Nil(db.With("Roles").Where("name = ?", "many_to_many_name").First(&user1))
		s.True(user.ID > 0)
		s.True(len(user.Roles) == 2)

		var role Role
		s.Nil(db.With("Users").Where("name = ?", "many_to_many_role1").First(&role))
		s.True(role.ID > 0)
		s.True(len(role.Users) == 1)
		s.Equal("many_to_many_name", role.Users[0].Name)
	}
}

func (s *GormQueryTestSuite) TestLimit() {
	for _, db := range s.dbs {
		user := User{Name: "limit_user", Avatar: "limit_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "limit_user", Avatar: "limit_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var user2 []User
		s.Nil(db.Where("name = ?", "limit_user").Limit(1).Get(&user2))
		s.True(len(user2) > 0)
		s.True(user2[0].ID > 0)
	}
}

func (s *GormQueryTestSuite) TestLoad() {
	for _, db := range s.dbs {
		user := User{Name: "load_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
		user.Address.Name = "load_address"
		user.Books[0].Name = "load_book0"
		user.Books[1].Name = "load_book1"
		s.Nil(db.Select(orm.Associations).Create(&user))
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
					s.Nil(db.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.True(len(user1.Books) == 0)
					s.Nil(db.Load(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 0)
					s.Nil(db.Load(&user1, "Books"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 2)
				},
			},
			{
				description: "load relationship with simple condition",
				setup: func(description string) {
					var user1 User
					s.Nil(db.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(db.Load(&user1, "Books", "name = ?", "load_book0"))
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
					s.Nil(db.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.Nil(db.Load(&user1, "Books", func(query ormcontract.Query) ormcontract.Query {
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
					s.Nil(db.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.Equal(0, len(user1.Books))
					s.EqualError(db.Load(&user1, ""), "relation cannot be empty")
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
					s.EqualError(db.Load(&userNoID, "Book"), "id cannot be empty")
				},
			},
		}
		for _, test := range tests {
			test.setup(test.description)
		}
	}
}

func (s *GormQueryTestSuite) TestLoadMissing() {
	for _, db := range s.dbs {
		user := User{Name: "load_missing_user", Address: &Address{}, Books: []*Book{&Book{}, &Book{}}}
		user.Address.Name = "load_missing_address"
		user.Books[0].Name = "load_missing_book0"
		user.Books[1].Name = "load_missing_book1"
		s.Nil(db.Select(orm.Associations).Create(&user))
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
					s.Nil(db.Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.True(len(user1.Books) == 0)
					s.Nil(db.LoadMissing(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 0)
					s.Nil(db.LoadMissing(&user1, "Books"))
					s.True(user1.Address.ID > 0)
					s.True(len(user1.Books) == 2)
				},
			},
			{
				description: "don't load when not missing",
				setup: func(description string) {
					var user1 User
					s.Nil(db.With("Books", "name = ?", "load_missing_book0").Find(&user1, user.ID))
					s.True(user1.ID > 0)
					s.Nil(user1.Address)
					s.True(len(user1.Books) == 1)
					s.Nil(db.LoadMissing(&user1, "Address"))
					s.True(user1.Address.ID > 0)
					s.Nil(db.LoadMissing(&user1, "Books"))
					s.True(len(user1.Books) == 1)
				},
			},
		}
		for _, test := range tests {
			test.setup(test.description)
		}
	}
}

func (s *GormQueryTestSuite) TestRaw() {
	for _, db := range s.dbs {
		user := User{Name: "raw_user", Avatar: "raw_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		var user1 User
		s.Nil(db.Raw("SELECT id, name FROM users WHERE name = ?", "raw_user").Scan(&user1))
		s.True(user1.ID > 0)
		s.Equal("raw_user", user1.Name)
		s.Equal("", user1.Avatar)
	}
}

func (s *GormQueryTestSuite) TestScope() {
	for _, db := range s.dbs {
		users := []User{{Name: "scope_user", Avatar: "scope_avatar"}, {Name: "scope_user1", Avatar: "scope_avatar1"}}
		s.Nil(db.Create(&users))
		s.True(users[0].ID > 0)
		s.True(users[1].ID > 0)

		var users1 []User
		s.Nil(db.Scopes(paginator("1", "1")).Find(&users1))

		s.Equal(1, len(users1))
		s.True(users1[0].ID > 0)
	}
}

func (s *GormQueryTestSuite) TestFind() {
	for _, db := range s.dbs {
		user := User{Name: "find_user"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		var user2 User
		s.Nil(db.Find(&user2, user.ID))
		s.True(user2.ID > 0)

		var user3 []User
		s.Nil(db.Find(&user3, []uint{user.ID}))
		s.Equal(1, len(user3))

		var user4 []User
		s.Nil(db.Where("id in ?", []uint{user.ID}).Find(&user4))
		s.Equal(1, len(user4))
	}
}

func (s *GormQueryTestSuite) TestGet() {
	for _, db := range s.dbs {
		user := User{Name: "get_user"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		var user5 []User
		s.Nil(db.Where("id in ?", []uint{user.ID}).Get(&user5))
		s.Equal(1, len(user5))
	}
}

func (s *GormQueryTestSuite) TestSelect() {
	for _, db := range s.dbs {
		user := User{Name: "select_user", Avatar: "select_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "select_user", Avatar: "select_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		user2 := User{Name: "select_user1", Avatar: "select_avatar1"}
		s.Nil(db.Create(&user2))
		s.True(user2.ID > 0)

		type Result struct {
			Name  string
			Count string
		}
		var result []Result
		s.Nil(db.Model(&User{}).Select("name, count(avatar) as count").Where("id in ?", []uint{user.ID, user1.ID, user2.ID}).Group("name").Get(&result))
		s.Equal(2, len(result))
		s.Equal("select_user", result[0].Name)
		s.Equal("2", result[0].Count)
		s.Equal("select_user1", result[1].Name)
		s.Equal("1", result[1].Count)

		var result1 []Result
		s.Nil(db.Model(&User{}).Select("name, count(avatar) as count").Group("name").Having("name = ?", "select_user").Get(&result1))

		s.Equal(1, len(result1))
		s.Equal("select_user", result1[0].Name)
		s.Equal("2", result1[0].Count)
	}
}

func (s *GormQueryTestSuite) TestSoftDelete() {
	for _, db := range s.dbs {
		user := User{Name: "soft_delete_user", Avatar: "soft_delete_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		s.Nil(db.Where("name = ?", "soft_delete_user").Delete(&User{}))

		var user1 User
		s.Nil(db.Find(&user1, user.ID))
		s.Equal(uint(0), user1.ID)

		var user2 User
		s.Nil(db.WithTrashed().Find(&user2, user.ID))
		s.True(user2.ID > 0)

		s.Nil(db.Where("name = ?", "soft_delete_user").ForceDelete(&User{}))

		var user3 User
		s.Nil(db.WithTrashed().Find(&user3, user.ID))
		s.Equal(uint(0), user3.ID)
	}
}

func (s *GormQueryTestSuite) TestTransactionSuccess() {
	for _, db := range s.dbs {
		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
		tx, err := db.Begin()
		s.Nil(err)
		s.Nil(tx.Create(&user))
		s.Nil(tx.Create(&user1))
		s.Nil(tx.Commit())

		var user2, user3 User
		s.Nil(db.Find(&user2, user.ID))
		s.Nil(db.Find(&user3, user1.ID))
	}
}

func (s *GormQueryTestSuite) TestTransactionError() {
	for _, db := range s.dbs {
		user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
		user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
		tx, err := db.Begin()
		s.Nil(err)
		s.Nil(tx.Create(&user))
		s.Nil(tx.Create(&user1))
		s.Nil(tx.Rollback())

		var users []User
		s.Nil(db.Where("name = ? or name = ?", "transaction_error_user", "transaction_error_user1").Find(&users))
		s.Equal(0, len(users))
	}
}

func (s *GormQueryTestSuite) TestUpdate() {
	for _, db := range s.dbs {
		user := User{Name: "update_user", Avatar: "update_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user.Name = "update_user1"
		s.Nil(db.Save(&user))
		s.Nil(db.Model(&User{}).Where("id = ?", user.ID).Update("avatar", "update_avatar1"))

		var user1 User
		s.Nil(db.Find(&user1, user.ID))
		s.Equal("update_user1", user1.Name)
		s.Equal("update_avatar1", user1.Avatar)
	}
}

func (s *GormQueryTestSuite) TestWhere() {
	for _, db := range s.dbs {
		user := User{Name: "where_user", Avatar: "where_avatar"}
		s.Nil(db.Create(&user))
		s.True(user.ID > 0)

		user1 := User{Name: "where_user1", Avatar: "where_avatar1"}
		s.Nil(db.Create(&user1))
		s.True(user1.ID > 0)

		var user2 []User
		s.Nil(db.Where("name = ?", "where_user").OrWhere("avatar = ?", "where_avatar1").Find(&user2))
		s.True(len(user2) > 0)

		var user3 User
		s.Nil(db.Where("name = 'where_user'").Find(&user3))
		s.True(user3.ID > 0)

		var user4 User
		s.Nil(db.Where("name", "where_user").Find(&user4))
		s.True(user4.ID > 0)
	}
}

func (s *GormQueryTestSuite) TestWith() {
	for _, db := range s.dbs {
		user := User{Name: "with_user", Address: &Address{
			Name: "with_address",
		}, Books: []*Book{{
			Name: "with_book0",
		}, {
			Name: "with_book1",
		}}}
		s.Nil(db.Select(orm.Associations).Create(&user))
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
					s.Nil(db.With("Address").With("Books").Find(&user1, user.ID))
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
					s.Nil(db.With("Books", "name = ?", "with_book0").Find(&user1, user.ID))
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
					s.Nil(db.With("Books", func(query ormcontract.Query) ormcontract.Query {
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
	}
}

func (s *GormQueryTestSuite) TestWithNesting() {
	for _, db := range s.dbs {
		user := User{Name: "with_nesting_user", Books: []*Book{{
			Name:   "with_nesting_book0",
			Author: &Author{Name: "with_nesting_author0"},
		}, {
			Name:   "with_nesting_book1",
			Author: &Author{Name: "with_nesting_author1"},
		}}}
		s.Nil(db.Select(orm.Associations).Create(&user))
		s.True(user.ID > 0)
		s.True(user.Books[0].ID > 0)
		s.True(user.Books[0].Author.ID > 0)
		s.True(user.Books[1].ID > 0)
		s.True(user.Books[1].Author.ID > 0)

		var user1 User
		s.Nil(db.With("Books.Author").Find(&user1, user.ID))
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
	}
}
