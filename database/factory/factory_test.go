package factory

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/suite"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/factory"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Timestamps
}

type SoftDeletes struct {
	DeletedAt gormio.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

type Timestamps struct {
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type User struct {
	Model
	SoftDeletes
	Name   string
	Avatar string
}

func (u *User) Factory() factory.Factory {
	return &UserFactory{}
}

type UserFactory struct {
}

func (u *UserFactory) Definition() map[string]any {
	faker := gofakeit.New(0)
	return map[string]any{
		"Name":      faker.Name(),
		"Avatar":    faker.Email(),
		"CreatedAt": carbon.NewDateTime(carbon.Now()),
		"UpdatedAt": carbon.NewDateTime(carbon.Now()),
	}
}

type Author struct {
	Model
	BookID uint
	Name   string
	SoftDeletes
}

func (a *Author) Factory() factory.Factory {
	return &AuthorFactory{}
}

type AuthorFactory struct {
}

func (a *AuthorFactory) Definition() map[string]any {
	faker := gofakeit.New(0)
	return map[string]any{
		"ID":        1,
		"BookID":    2,
		"Name":      faker.Name(),
		"CreatedAt": carbon.NewDateTime(carbon.Now()),
		"UpdatedAt": carbon.NewDateTime(carbon.Now()),
		"DeletedAt": gormio.DeletedAt{Time: time.Now(), Valid: true},
	}
}

type House struct {
	Model
	Name          string
	HouseableID   uint
	HouseableType string
}

type FactoryTestSuite struct {
	suite.Suite
	factory *FactoryImpl
	query   ormcontract.Query
}

func TestFactoryTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &FactoryTestSuite{})
}

func (s *FactoryTestSuite) SetupSuite() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	postgresQuery := gorm.NewTestQuery(postgresDocker)
	postgresQuery.CreateTable(gorm.TestTableHouses, gorm.TestTableUsers)

	s.query = postgresQuery.Query()
}

func (s *FactoryTestSuite) SetupTest() {
	s.factory = NewFactoryImpl(s.query)
}

func (s *FactoryTestSuite) TestTimes() {
	var user []User
	s.Nil(s.factory.Count(2).Make(&user))
	s.True(len(user) == 2)
	s.True(len(user[0].Name) > 0)
	s.True(len(user[1].Name) > 0)
}

func (s *FactoryTestSuite) TestCreate() {
	var user []User
	s.Nil(s.factory.Create(&user))
	s.True(len(user) == 1)
	s.True(user[0].ID > 0)
	s.True(len(user[0].Name) > 0)
	s.True(len(user[0].Avatar) > 0)
	s.NotEmpty(user[0].CreatedAt.String())
	s.NotEmpty(user[0].UpdatedAt.String())

	var user1 User
	s.Nil(s.factory.Create(&user1))
	s.NotNil(user1)
	s.True(user1.ID > 0)
	s.True(len(user1.Avatar) > 0)
	s.NotEmpty(user1.CreatedAt.String())
	s.NotEmpty(user1.UpdatedAt.String())

	var user2 User
	s.Nil(s.factory.Create(&user2, map[string]any{
		"Avatar": "avatar",
	}))
	s.NotNil(user2)
	s.True(user2.ID > 0)
	s.Equal("avatar", user2.Avatar)
	s.NotEmpty(user2.CreatedAt.String())
	s.NotEmpty(user2.UpdatedAt.String())

	var user3 []User
	s.Nil(s.factory.Count(2).Create(&user3))
	s.True(len(user3) == 2)
	s.True(user3[0].ID > 0)
	s.True(user3[1].ID > 0)
	s.True(len(user3[0].Name) > 0)
	s.True(len(user3[1].Name) > 0)
}

func (s *FactoryTestSuite) TestCreateQuietly() {
	var user []User
	s.Nil(s.factory.CreateQuietly(&user))
	s.True(len(user) == 1)
	s.True(user[0].ID > 0)
	s.True(len(user[0].Name) > 0)
	s.True(len(user[0].Avatar) > 0)
	s.NotEmpty(user[0].CreatedAt.String())
	s.NotEmpty(user[0].UpdatedAt.String())

	var user1 User
	s.Nil(s.factory.CreateQuietly(&user1))
	s.NotNil(user1)
	s.True(user1.ID > 0)
	s.True(len(user1.Avatar) > 0)
	s.NotEmpty(user1.CreatedAt.String())
	s.NotEmpty(user1.UpdatedAt.String())

	var user2 User
	s.Nil(s.factory.CreateQuietly(&user2, map[string]any{
		"Avatar": "avatar",
	}))
	s.NotNil(user2)
	s.True(user2.ID > 0)
	s.Equal("avatar", user2.Avatar)
	s.NotEmpty(user2.CreatedAt.String())
	s.NotEmpty(user2.UpdatedAt.String())

	var user3 []User
	s.Nil(s.factory.Count(2).CreateQuietly(&user3))
	s.True(len(user3) == 2)
	s.True(user3[0].ID > 0)
	s.True(user3[1].ID > 0)
	s.True(len(user3[0].Name) > 0)
	s.True(len(user3[1].Name) > 0)
}

func (s *FactoryTestSuite) TestMake() {
	var user User
	s.Nil(s.factory.Make(&user))
	s.True(user.ID == 0)
	s.True(len(user.Name) > 0)
	s.True(len(user.Avatar) > 0)
	s.NotEmpty(user.CreatedAt.String())
	s.NotEmpty(user.UpdatedAt.String())

	var user1 User
	s.Nil(s.factory.Make(&user1, map[string]any{
		"Avatar": "avatar",
	}))
	s.True(user1.ID == 0)
	s.True(len(user1.Name) > 0)
	s.Equal("avatar", user1.Avatar)
	s.NotEmpty(user1.CreatedAt.String())
	s.NotEmpty(user1.UpdatedAt.String())

	var users []User
	s.Nil(s.factory.Make(&users))
	s.True(len(users) == 1)
	s.True(users[0].ID == 0)
	s.True(len(users[0].Name) > 0)
	s.True(len(users[0].Avatar) > 0)
	s.NotEmpty(users[0].CreatedAt.String())
	s.NotEmpty(users[0].UpdatedAt.String())

	var author Author
	s.Nil(s.factory.Make(&author))
	s.True(author.ID == 1)
	s.True(len(author.Name) > 0)
	s.True(author.BookID == 2)
	s.NotEmpty(author.CreatedAt.String())
	s.NotEmpty(author.UpdatedAt.String())
	s.True(author.DeletedAt.Valid)
}

func (s *FactoryTestSuite) TestGetRawAttributes() {
	var house House
	attributes, err := s.factory.getRawAttributes(&house)
	s.NotNil(err)
	s.Nil(attributes)

	var user User
	attributes, err = s.factory.getRawAttributes(&user)
	s.Nil(err)
	s.NotNil(attributes)

	var user1 User
	attributes, err = s.factory.getRawAttributes(&user1, map[string]any{
		"Avatar": "avatar",
	})
	s.Nil(err)
	s.NotNil(attributes)
	s.True(len(attributes["Name"].(string)) > 0)
	s.Equal("avatar", attributes["Avatar"].(string))
	s.NotNil(attributes["CreatedAt"])
	s.NotNil(attributes["UpdatedAt"])
}
