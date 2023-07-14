package gorm

import (
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/factory"
	ormmocks "github.com/goravel/framework/contracts/database/orm/mocks"
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
	factory   *FactoryImpl
	mockQuery *ormmocks.Query
}

func TestFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryTestSuite))
}

func (s *FactoryTestSuite) SetupTest() {
	s.mockQuery = ormmocks.NewQuery(s.T())
	s.factory = NewFactoryImpl(s.mockQuery)
}

func (s *FactoryTestSuite) TestTimes() {
	factInstance := s.factory.Times(2)
	s.NotNil(factInstance)
	s.NotNil(reflect.TypeOf(factInstance) == reflect.TypeOf((*FactoryImpl)(nil)))
	s.mockQuery.AssertExpectations(s.T())
}

func (s *FactoryTestSuite) TestCreate() {
	var person []Person
	s.mockQuery.On("Create", &person).Return(nil).Once()
	s.Nil(s.factory.Create(&person))
	s.True(len(person) > 0)
	s.True(reflect.TypeOf(person) == reflect.TypeOf([]Person{}))
	s.mockQuery.AssertExpectations(s.T())

	var person1 Person
	s.mockQuery.On("Create", &person1).Return(nil).Once()
	s.Nil(s.factory.Create(&person1))
	s.NotNil(person1)
	s.True(reflect.TypeOf(person1) == reflect.TypeOf(Person{}))
	s.mockQuery.AssertExpectations(s.T())
}

func (s *FactoryTestSuite) TestMake() {
	var person []Person
	s.Nil(s.factory.Make(&person))
	s.True(len(person) > 0)
	s.True(reflect.TypeOf(person) == reflect.TypeOf([]Person{}))
	s.mockQuery.AssertExpectations(s.T())
}
