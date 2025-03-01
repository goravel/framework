package db

import (
	"context"
	databasesql "database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
	"github.com/goravel/framework/support/carbon"
)

// TestUser is a test model
type TestUser struct {
	ID    uint   `db:"id"`
	Phone string `db:"phone"`
	Name  string `db:"-"`
	Age   int
}

type QueryTestSuite struct {
	suite.Suite
	ctx         context.Context
	mockBuilder *mocksdb.Builder
	mockDriver  *mocksdriver.Driver
	mockLogger  *mockslogger.Logger
	now         carbon.Carbon
	query       *Query
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, &QueryTestSuite{})
}

func (s *QueryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockBuilder = mocksdb.NewBuilder(s.T())
	s.mockDriver = mocksdriver.NewDriver(s.T())
	s.mockLogger = mockslogger.NewLogger(s.T())
	s.now = carbon.Now()
	carbon.SetTestNow(s.now)

	s.query = NewQuery(s.ctx, s.mockDriver, s.mockBuilder, s.mockLogger, "users")
}

func (s *QueryTestSuite) TestDelete() {
	s.Run("success", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("failed to exec", func() {
		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(nil, assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		_, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Equal(assert.AnError, err)
	})

	s.Run("failed to get rows affected", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(0), assert.AnError).Once()

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		_, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestFirst() {
	s.Run("success", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Nil(err)
	})

	s.Run("failed to get", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Equal(assert.AnError, err)
	})

	s.Run("no rows", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestGet() {
	s.Run("success", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ?", 25).Run(func(dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ?", 25).Return("SELECT * FROM users WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25", int64(2), nil).Return().Once()

		err := s.query.Where("age", 25).Get(&users)
		s.Nil(err)
		s.mockBuilder.AssertExpectations(s.T())
	})

	s.Run("failed to get", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ?", 25).Return(assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ?", 25).Return("SELECT * FROM users WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("age", 25).Get(&users)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestInsert() {
	s.Run("empty", func() {
		result, err := s.query.Insert(nil)
		s.Nil(err)
		s.Equal(int64(0), result.RowsAffected)
	})

	s.Run("single struct", func() {
		user := TestUser{
			ID:   1,
			Name: "John",
			Age:  25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?)", uint(1)).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (id) VALUES (?)", uint(1)).Return("INSERT INTO users (id) VALUES (1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (id) VALUES (1)", int64(1), nil).Return().Once()

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

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?),(?)", uint(1), uint(2)).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (id) VALUES (?),(?)", uint(1), uint(2)).Return("INSERT INTO users (id) VALUES (1),(2)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (id) VALUES (1),(2)", int64(2), nil).Return().Once()

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

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (age,id,name) VALUES (?,?,?)", 25, 1, "John").Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (age,id,name) VALUES (?,?,?)", 25, 1, "John").Return("INSERT INTO users (age,id,name) VALUES (25,1,\"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (age,id,name) VALUES (25,1,\"John\")", int64(1), nil).Return().Once()

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

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (age,id,name) VALUES (?,?,?),(?,?,?)", 25, 1, "John", 30, 2, "Jane").Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (age,id,name) VALUES (?,?,?),(?,?,?)", 25, 1, "John", 30, 2, "Jane").Return("INSERT INTO users (age,id,name) VALUES (25,1,\"John\"),(30,2,\"Jane\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (age,id,name) VALUES (25,1,\"John\"),(30,2,\"Jane\")", int64(2), nil).Return().Once()

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

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (id) VALUES (?)", uint(1)).Return(nil, assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (id) VALUES (?)", uint(1)).Return("INSERT INTO users (id) VALUES (1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (id) VALUES (1)", int64(-1), assert.AnError).Return().Once()

		result, err := s.query.Insert(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestUpdate() {
	s.Run("single struct", func() {
		user := TestUser{
			Phone: "1234567890",
			Name:  "John",
			Age:   25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("single map", func() {
		user := map[string]any{
			"phone": "1234567890",
			"name":  "John",
			"age":   25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("UPDATE users SET age = ?, name = ?, phone = ? WHERE (name = ? AND id = ?)", 25, "John", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("UPDATE users SET age = ?, name = ?, phone = ? WHERE (name = ? AND id = ?)", 25, "John", "1234567890", "John", 1).Return("UPDATE users SET age = 25, name = \"John\", phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = 25, name = \"John\", phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("failed to exec", func() {
		user := TestUser{
			Phone: "1234567890",
			Name:  "John",
			Age:   25,
		}

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(nil, assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})

	s.Run("failed to get rows affected", func() {
		user := TestUser{
			Phone: "1234567890",
			Name:  "John",
			Age:   25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(0), assert.AnError).Once()

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestWhere() {
	now := carbon.Now()
	carbon.SetTestNow(now)

	s.Run("simple condition", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, now, "SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Where("age", 25).Where("age", []int{25, 30}).First(&user)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age > ?", 18).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age > ?", 18).Return("SELECT * FROM users WHERE age > 18").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age > 18", int64(0), nil).Return().Once()

		err := s.query.Where("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND (age IN (?,?) AND name = ?))", "John", 25, 30, "Tom").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND (age IN (?,?) AND name = ?))", "John", 25, 30, "Tom").Return("SELECT * FROM users WHERE (name = \"John\" AND (age IN (25,30) AND name = \"Tom\"))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND (age IN (25,30) AND name = \"Tom\"))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").Where(func(query db.Query) {
			query.Where("age", []int{25, 30}).Where("name", "Tom")
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestOrWhere() {
	now := carbon.Now()
	carbon.SetTestNow(now)

	s.Run("simple condition", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE (((name = ? AND age = ?) OR age IN (?,?)) OR name = ?)", "John", 25, 30, 40, "Jane").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (((name = ? AND age = ?) OR age IN (?,?)) OR name = ?)", "John", 25, 30, 40, "Jane").Return("SELECT * FROM users WHERE (((name = \"John\" AND age = 25) OR age IN (30,40)) OR name = \"Jane\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, now, "SELECT * FROM users WHERE (((name = \"John\" AND age = 25) OR age IN (30,40)) OR name = \"Jane\")", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Where("age", 25).OrWhere("age", []int{30, 40}).OrWhere("name", "Jane").First(&user)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age > ?)", "John", 18).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age > ?)", "John", 18).Return("SELECT * FROM users WHERE (name = \"John\" OR age > 18)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age > 18)", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").OrWhere("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR ((age IN (?,?) AND name = ?) OR age = ?))", "John", 25, 30, "Tom", 40).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR ((age IN (?,?) AND name = ?) OR age = ?))", "John", 25, 30, "Tom", 40).Return("SELECT * FROM users WHERE (name = \"John\" OR ((age IN (25,30) AND name = \"Tom\") OR age = 40))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR ((age IN (25,30) AND name = \"Tom\") OR age = 40))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").OrWhere(func(query db.Query) {
			query.Where("age", []int{25, 30}).Where("name", "Tom").OrWhere("age", 40)
		}).Get(&users)
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
