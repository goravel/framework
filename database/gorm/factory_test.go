package gorm

import (
	"log"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type UserFactory struct {
}

func (u *UserFactory) Definition() any {
	faker := gofakeit.New(0)
	return map[string]interface{}{
		"name":       faker.Name(),
		"avatar":     faker.Email(),
		"created_at": faker.Date(),
		"updated_at": faker.Date(),
	}
}

type FactoryTestSuite struct {
	suite.Suite
	factory *FactoryImpl
	query   ormcontract.Query
}

func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) SetupTest() {
	mysqlDocker := NewMysqlDocker()
	_, _, mysqlQuery, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}
	s.query = mysqlQuery
	s.factory = NewFactoryImpl(s.query)
}

func (s *FactoryTestSuite) TestTimes() {
	var user []User
	factInstance := s.factory.Times(2)
	s.Nil(factInstance.Make(&user))
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

	var user1 User
	s.Nil(s.factory.Create(&user1))
	s.NotNil(user1)
	s.True(user1.ID > 0)

	var user3 []User
	factInstance := s.factory.Times(2)
	s.Nil(factInstance.Create(&user3))
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

	var user1 User
	s.Nil(s.factory.CreateQuietly(&user1))
	s.NotNil(user1)
	s.True(user1.ID > 0)

	var user3 []User
	factInstance := s.factory.Times(2)
	s.Nil(factInstance.CreateQuietly(&user3))
	s.True(len(user3) == 2)
	s.True(user3[0].ID > 0)
	s.True(user3[1].ID > 0)
	s.True(len(user3[0].Name) > 0)
	s.True(len(user3[1].Name) > 0)
}

func (s *FactoryTestSuite) TestMake() {
	var user []User
	s.Nil(s.factory.Make(&user))
	s.True(len(user) == 1)
	s.True(len(user[0].Name) > 0)
}

func (s *FactoryTestSuite) TestGetRawAttributes() {
	var author Author
	attributes, err := s.factory.getRawAttributes(&author)
	s.NotNil(err)
	s.Nil(attributes)

	var house House
	attributes, err = s.factory.getRawAttributes(&house)
	s.NotNil(err)
	s.Nil(attributes)

	var user User
	attributes, err = s.factory.getRawAttributes(&user)
	s.Nil(err)
	s.NotNil(attributes)
}
