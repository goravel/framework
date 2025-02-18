package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
)

// TestUser is a test model
type TestUser struct {
	ID   uint   `db:"id"`
	Name string `db:"-"`
	Age  int
}

type QueryTestSuite struct {
	suite.Suite
	mockBuilder *mocksdb.Builder
	query       *Query
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, &QueryTestSuite{})
}

func (s *QueryTestSuite) SetupTest() {
	s.mockBuilder = mocksdb.NewBuilder(s.T())
	s.query = NewQuery(database.Config{}, s.mockBuilder, "users")
}

func (s *QueryTestSuite) TestFirst() {
	var user TestUser
	s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()

	err := s.query.Where("name", "John").First(&user)
	s.Nil(err)
}

func (s *QueryTestSuite) TestGet() {
	var users []TestUser
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ?", 25).Return(nil).Once()

	err := s.query.Where("age", 25).Get(&users)
	s.Nil(err)
	s.mockBuilder.AssertExpectations(s.T())
}

func (s *QueryTestSuite) TestInsert() {
	s.Run("single struct", func() {
		user := TestUser{
			ID:   1,
			Name: "John",
			Age:  25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?)", uint(1)).Return(mockResult, nil).Once()

		result, err := s.query.Insert(user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("multiple structs", func() {
		users := []TestUser{
			{ID: 1, Name: "John", Age: 25},
			{ID: 2, Name: "Jane", Age: 30},
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(2), nil)
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?),(?)", uint(1), uint(2)).Return(mockResult, nil).Once()

		result, err := s.query.Insert(users)
		s.Nil(err)
		s.Equal(int64(2), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("single map", func() {
		user := map[string]any{
			"id":   1,
			"name": "John",
			"age":  25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (age,id,name) VALUES (?,?,?)", 25, 1, "John").Return(mockResult, nil).Once()

		result, err := s.query.Insert(user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("multiple maps", func() {
		users := []map[string]any{
			{"id": 1, "name": "John", "age": 25},
			{"id": 2, "name": "Jane", "age": 30},
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(2), nil)
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (age,id,name) VALUES (?,?,?),(?,?,?)", 25, 1, "John", 30, 2, "Jane").Return(mockResult, nil).Once()

		result, err := s.query.Insert(users)
		s.Nil(err)
		s.Equal(int64(2), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("unknown type", func() {
		user := "unknown"

		_, err := s.query.Insert(user)
		s.Equal(errors.DatabaseUnsupportedType.Args("string", "struct, []struct, map[string]any, []map[string]any").SetModule("DB"), err)
	})

	s.Run("failed to exec", func() {
		user := TestUser{
			ID:   1,
			Name: "John",
			Age:  25,
		}

		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?)", uint(1)).Return(nil, assert.AnError).Once()

		result, err := s.query.Insert(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestWhere() {
	s.Run("simple where condition", func() {
		var user TestUser
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()

		err := s.query.Where("name", "John").First(&user)
		s.Nil(err)
	})

	s.Run("where with multiple arguments", func() {
		var users []TestUser
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age IN (?,?)", 25, 30).Return(nil).Once()

		err := s.query.Where("age", []int{25, 30}).Get(&users)
		s.Nil(err)
	})

	s.Run("where with raw query", func() {
		var users []TestUser
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age > ?", 18).Return(nil).Once()

		err := s.query.Where("age > ?", 18).Get(&users)
		s.Nil(err)
	})
}

// MockResult implements sql.Result interface for testing
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	arguments := m.Called()
	return arguments.Get(0).(int64), arguments.Error(1)
}
