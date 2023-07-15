package gorm

import (
	"log"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/factory"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
)

type Person struct {
	orm.Model
	orm.SoftDeletes
	Name   string
	Avatar string
}

func (p *Person) Factory() factory.Factory {
	return &PersonFactory{}
}

type PersonFactory struct {
}

func (p *PersonFactory) Definition() any {
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
	factory ormcontract.Factory
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
	var person []Person
	factInstance := s.factory.Times(2)
	s.Nil(factInstance.Make(&person))
	s.True(len(person) == 2)
	s.True(len(person[0].Name) > 0)
	s.True(len(person[1].Name) > 0)
}

func (s *FactoryTestSuite) TestCreate() {
	var person []Person
	s.Nil(s.factory.Create(&person))
	s.True(len(person) > 0)
	s.True(person[0].ID > 0)

	var person1 Person
	s.Nil(s.factory.Create(&person1))
	s.NotNil(person1)
	s.True(person1.ID > 0)
}

func (s *FactoryTestSuite) TestMake() {
	var person []Person
	s.Nil(s.factory.Make(&person))
	s.True(len(person) > 0)
	s.True(len(person[0].Name) > 0)
}
