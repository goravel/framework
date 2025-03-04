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

	s.query = NewQuery(s.ctx, s.mockDriver, s.mockBuilder, s.mockLogger, "users", nil)
}

func (s *QueryTestSuite) TestCount() {
	var count int64

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Get(&count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(dest any, query string, args ...any) {
		destCount := dest.(*int64)
		*destCount = 1
	}).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

	count, err := s.query.Where("name", "John").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *QueryTestSuite) TestDecrement() {
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Exec("UPDATE users SET age = age - ? WHERE name = ?", uint64(1), "John").Return(mockResult, nil).Once()
	s.mockDriver.EXPECT().Explain("UPDATE users SET age = age - ? WHERE name = ?", uint64(1), "John").Return("UPDATE users SET age = age - 1 WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = age - 1 WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Decrement("age")
	s.NoError(err)

	mockResult.AssertExpectations(s.T())
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

func (s *QueryTestSuite) TestDistinct() {
	var users TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Get(&users, "SELECT DISTINCT * FROM users WHERE name = ?", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT DISTINCT * FROM users WHERE name = ?", "John").Return("SELECT DISTINCT * FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT DISTINCT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Distinct().First(&users)
	s.NoError(err)
}

func (s *QueryTestSuite) TestExists() {
	var count int64

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Get(&count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(dest any, query string, args ...any) {
		destCount := dest.(*int64)
		*destCount = 1
	}).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

	exists, err := s.query.Where("name", "John").Exists()
	s.NoError(err)
	s.True(exists)
}

func (s *QueryTestSuite) TestFind() {
	s.Run("single ID", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("SELECT * FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Find(&user, 1)

		s.NoError(err)
	})

	s.Run("multiple ID", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Run(func(dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Return("SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))", int64(2), nil).Return().Once()

		err := s.query.Where("name", "John").Find(&users, []int{1, 2})

		s.NoError(err)
	})

	s.Run("primary key is not id", func() {
		var users TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&users, "SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return("SELECT * FROM users WHERE (name = \"John\" AND uuid = \"123\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND uuid = \"123\")", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Find(&users, "uuid", "123")

		s.NoError(err)
	})

	s.Run("invalid argument number", func() {
		var users []TestUser

		err := s.query.Where("name", "John").Find(&users, 1, 2, 3)
		s.Equal(errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
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

func (s *QueryTestSuite) TestFirstOr() {
	var user TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").FirstOr(&user, func() error {
		return errors.New("no rows")
	})

	s.Equal(errors.New("no rows"), err)
}

func (s *QueryTestSuite) TestFirstOrFail() {
	s.Run("success", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Nil(err)
	})

	s.Run("failed to get", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(assert.AnError).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Equal(assert.AnError, err)
	})

	s.Run("no rows", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), databasesql.ErrNoRows).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Equal(databasesql.ErrNoRows, err)
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

func (s *QueryTestSuite) TestIncrement() {
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Exec("UPDATE users SET age = age + ? WHERE name = ?", uint64(1), "John").Return(mockResult, nil).Once()
	s.mockDriver.EXPECT().Explain("UPDATE users SET age = age + ? WHERE name = ?", uint64(1), "John").Return("UPDATE users SET age = age + 1 WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = age + 1 WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Increment("age")
	s.NoError(err)

	mockResult.AssertExpectations(s.T())
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

func (s *QueryTestSuite) TestInsertGetId() {
	s.Run("empty", func() {
		id, err := s.query.InsertGetId(nil)
		s.Equal(errors.DatabaseUnsupportedType.Args("nil", "struct, map[string]any").SetModule("DB"), err)
		s.Equal(int64(0), id)
	})

	s.Run("success", func() {
		user := map[string]any{
			"name": "John",
			"age":  25,
		}

		mockResult := &MockResult{}
		mockResult.On("LastInsertId").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("INSERT INTO users (age,name) VALUES (?,?)", 25, "John").Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("INSERT INTO users (age,name) VALUES (?,?)", 25, "John").Return("INSERT INTO users (age,name) VALUES (25,\"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (age,name) VALUES (25,\"John\")", int64(1), nil).Return().Once()

		id, err := s.query.InsertGetId(user)
		s.Nil(err)
		s.Equal(int64(1), id)

		mockResult.AssertExpectations(s.T())
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

// func (s *QueryTestSuite) TestLimit() {
// 	var users []TestUser

// 	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
// 	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ? LIMIT 1", 25).Return(nil).Once()
// 	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? LIMIT 1", 25).Return("SELECT * FROM users WHERE age = 25 LIMIT 1").Once()
// 	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 LIMIT 1", int64(0), nil).Return().Once()

// 	err := s.query.Where("age", 25).Limit(1).Get(&users)
// 	s.Nil(err)
// }

func (s *QueryTestSuite) TestLatest() {
	s.Run("default column", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE age = ? ORDER BY created_at DESC", 25).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY created_at DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY created_at DESC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY created_at DESC", int64(1), nil).Return().Once()

		err := s.query.Where("age", 25).Latest(&user)
		s.Nil(err)
	})

	s.Run("custom column", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE age = ? ORDER BY name DESC", 25).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY name DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY name DESC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY name DESC", int64(1), nil).Return().Once()

		err := s.query.Where("age", 25).Latest(&user, "name")
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestOrderBy() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ? ORDER BY age ASC, id ASC", 25).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY age ASC, id ASC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id ASC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id ASC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("age").OrderBy("id").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrderByDesc() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ? ORDER BY age ASC, id DESC", 25).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY age ASC, id DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id DESC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id DESC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("age").OrderByDesc("id").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrderByRaw() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ? ORDER BY name ASC, age DESC, id ASC", 25).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY name ASC, age DESC, id ASC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY name ASC, age DESC, id ASC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY name ASC, age DESC, id ASC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("name").OrderByRaw("age DESC, id ASC").Get(&users)
	s.Nil(err)
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

		err := s.query.Where("name", "John").OrWhere(func(query db.Query) db.Query {
			return query.Where("age", []int{25, 30}).Where("name", "Tom").OrWhere("age", 40)
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestOrWhereBetween() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age BETWEEN ? AND ?)", "John", 18, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age BETWEEN ? AND ?)", "John", 18, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age BETWEEN 18 AND 30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age BETWEEN 18 AND 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereColumn() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR height = weight)", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR height = weight)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR height = weight)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR height = weight)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereColumn("height", "weight").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereIn() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age IN (?,?))", "John", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IN (?,?))", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age IN (25,30))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IN (25,30))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereLike() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR name LIKE ?)", "John", "%John%").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR name LIKE ?)", "John", "%John%").Return("SELECT * FROM users WHERE (name = \"John\" OR name LIKE \"%John%\")").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR name LIKE \"%John%\")", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNot() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR NOT (name = ?))", "John", "Jane").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR NOT (name = ?))", "John", "Jane").Return("SELECT * FROM users WHERE (name = \"John\" OR NOT (name = \"Jane\"))")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR NOT (name = \"Jane\"))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNot("name", "Jane").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotBetween() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age NOT BETWEEN ? AND ?)", "John", 18, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age NOT BETWEEN ? AND ?)", "John", 18, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age NOT BETWEEN 18 AND 30)")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age NOT BETWEEN 18 AND 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotIn() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age NOT IN (?,?))", "John", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age NOT IN (?,?))", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age NOT IN (25,30))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age NOT IN (25,30))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotLike() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR name NOT LIKE ?)", "John", "%John%").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR name NOT LIKE ?)", "John", "%John%").Return("SELECT * FROM users WHERE (name = \"John\" OR name NOT LIKE \"%John%\")").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR name NOT LIKE \"%John%\")", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotNull() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age IS NOT NULL)", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IS NOT NULL)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR age IS NOT NULL)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IS NOT NULL)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNull() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age IS NULL)", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IS NULL)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR age IS NULL)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IS NULL)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereRaw() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? OR age = ? or age = ?)", "John", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age = ? or age = ?)", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age = 25 OR age = 30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age = 25 OR age = 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereRaw("age = ? or age = ?", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestPluck() {
	var names []string

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&names, "SELECT name FROM users WHERE name = ?", "John").Run(func(dest any, query string, args ...any) {
		destNames := dest.(*[]string)
		*destNames = []string{"John"}
	}).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT name FROM users WHERE name = ?", "John").Return("SELECT name FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Pluck("name", &names)
	s.NoError(err)
	s.Equal([]string{"John"}, names)
}

func (s *QueryTestSuite) TestSelect() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT id, name FROM users WHERE name = ?", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT id, name FROM users WHERE name = ?", "John").Return("SELECT id, name FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT id, name FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

	err := s.query.Select("id", "name").Where("name", "John").Get(&users)
	s.Nil(err)
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

	s.Run("single column", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Exec("UPDATE users SET phone = ? WHERE name = ?", "1234567890", "John").Return(mockResult, nil).Once()
		s.mockDriver.EXPECT().Explain("UPDATE users SET phone = ? WHERE name = ?", "1234567890", "John").Return("UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"", int64(1), nil).Return().Once()

		result, err := s.query.Where("name", "John").Update("phone", "1234567890")
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("failed to update single column with wrong number of arguments", func() {
		_, err := s.query.Where("name", "John").Update("phone", "1234567890", "1234567890")
		s.Equal(errors.DatabaseInvalidArgumentNumber.Args(2, "1"), err)
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

// func (s *QueryTestSuite) TestValue() {
// 	var name string

// 	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
// 	s.mockBuilder.EXPECT().Get(&name, "SELECT name FROM users WHERE name = ? LIMIT 1", "John").Run(func(dest any, query string, args ...any) {
// 		destName := dest.(*string)
// 		*destName = "John"
// 	}).Return(nil).Once()
// 	s.mockDriver.EXPECT().Explain("SELECT name FROM users WHERE name = ? LIMIT 1", "John").Return("SELECT name FROM users WHERE name = \"John\" LIMIT 1").Once()
// 	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name FROM users WHERE name = \"John\" LIMIT 1", int64(-1), nil).Return().Once()

// 	err := s.query.Where("name", "John").Value("name", &name)
// 	s.NoError(err)
// 	s.Equal("John", name)
// }

func (s *QueryTestSuite) TestWhen() {
	s.Run("when condition is true", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 25).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 25).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 25)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND age = 25)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").When(true, func(query db.Query) db.Query {
			return query.Where("age", 25)
		}).First(&user)
		s.Nil(err)
	})

	s.Run("when condition is false", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").When(false, func(query db.Query) db.Query {
			return query.Where("age", 25)
		}).First(&user)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhere() {
	s.Run("simple condition", func() {
		var user TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Get(&user, "SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))", int64(1), nil).Return().Once()

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

		err := s.query.Where("name", "John").Where(func(query db.Query) db.Query {
			return query.Where("age", []int{25, 30}).Where("name", "Tom")
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhereBetween() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age BETWEEN ? AND ?", 18, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age BETWEEN ? AND ?", 18, 30).Return("SELECT * FROM users WHERE age BETWEEN 18 AND 30").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age BETWEEN 18 AND 30", int64(0), nil).Return().Once()

	err := s.query.WhereBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereColumn() {
	var users []TestUser

	s.Run("simple condition", func() {
		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (age = height AND name = ?)", "John").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (age = height AND name = ?)", "John").Return("SELECT * FROM users WHERE (age = height AND name = \"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (age = height AND name = \"John\")", int64(0), nil).Return().Once()

		err := s.query.WhereColumn("age", "height").Where("name", "John").Get(&users)
		s.Nil(err)
	})

	s.Run("with operator", func() {
		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (age > height AND name = ?)", "John").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (age > height AND name = ?)", "John").Return("SELECT * FROM users WHERE (age > height AND name = \"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (age > height AND name = \"John\")", int64(0), nil).Return().Once()

		err := s.query.WhereColumn("age", ">", "height").Where("name", "John").Get(&users)
		s.Nil(err)
	})

	s.Run("with multiple columns", func() {
		err := s.query.WhereColumn("age", ">", "height", "age").WhereColumn("name", "=", "John").Get(&users)
		s.Equal(errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
	})

	s.Run("with not enough arguments", func() {
		err := s.query.WhereColumn("age").WhereColumn("name", "=", "John").Get(&users)
		s.Equal(errors.DatabaseInvalidArgumentNumber.Args(2, "1 or 2"), err)
	})
}

func (s *QueryTestSuite) TestWhereExists() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Twice()
	s.mockDriver.EXPECT().Explain("SELECT * FROM agents WHERE age = ?", 25).Return("SELECT * FROM agents WHERE age = 25").Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND EXISTS (SELECT * FROM agents WHERE age = 25))", "John").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND EXISTS (SELECT * FROM agents WHERE age = 25))", "John").Return("SELECT * FROM users WHERE (name = \"John\" AND EXISTS (SELECT * FROM agents WHERE age = 25))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND EXISTS (SELECT * FROM agents WHERE age = 25))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").WhereExists(func() db.Query {
		return NewQuery(s.ctx, s.mockDriver, s.mockBuilder, s.mockLogger, "agents", nil).Where("age", 25)
	}).Get(&users)
	s.Nil(err)

}

func (s *QueryTestSuite) TestWhereIn() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age IN (?,?)", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age IN (?,?)", 25, 30).Return("SELECT * FROM users WHERE age IN (25,30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IN (25,30)", int64(0), nil).Return().Once()

	err := s.query.WhereIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereLike() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE name LIKE ?", "%John%").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name LIKE ?", "%John%").Return("SELECT * FROM users WHERE name LIKE \"%John%\"")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name LIKE \"%John%\"", int64(0), nil).Return().Once()

	err := s.query.WhereLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNot() {
	s.Run("simple condition", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND NOT (name = ?))", "John", "Jane").Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT (name = ?))", "John", "Jane").Return("SELECT * FROM users WHERE (name = \"John\" AND NOT (name = \"Jane\"))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT (name = \"Jane\"))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot("name", "Jane").Get(&users)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND NOT (age > ?))", "John", 18).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT (age > ?))", "John", 18).Return("SELECT * FROM users WHERE (name = \"John\" AND NOT (age > 18))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT (age > 18))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
		s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE (name = ? AND NOT ((name = ? AND age IN (?,?))))", "John", "Jane", 25, 30).Return(nil).Once()
		s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT ((name = ? AND age IN (?,?))))", "John", "Jane", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" AND NOT ((name = \"Jane\" AND age IN (25,30))))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT ((name = \"Jane\" AND age IN (25,30))))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot(func(query db.Query) db.Query {
			return query.Where("name", "Jane").Where("age", []int{25, 30})
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhereNotBetween() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age NOT BETWEEN ? AND ?", 18, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age NOT BETWEEN ? AND ?", 18, 30).Return("SELECT * FROM users WHERE age NOT BETWEEN 18 AND 30")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age NOT BETWEEN 18 AND 30", int64(0), nil).Return().Once()

	err := s.query.WhereNotBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotIn() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age NOT IN (?,?)", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age NOT IN (?,?)", 25, 30).Return("SELECT * FROM users WHERE age NOT IN (25,30)")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age NOT IN (25,30)", int64(0), nil).Return().Once()

	err := s.query.WhereNotIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotLike() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE name NOT LIKE ?", "%John%").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE name NOT LIKE ?", "%John%").Return("SELECT * FROM users WHERE name NOT LIKE \"%John%\"")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name NOT LIKE \"%John%\"", int64(0), nil).Return().Once()

	err := s.query.WhereNotLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotNull() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age IS NOT NULL").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age IS NOT NULL").Return("SELECT * FROM users WHERE age IS NOT NULL")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IS NOT NULL", int64(0), nil).Return().Once()

	err := s.query.WhereNotNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNull() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age IS NULL").Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age IS NULL").Return("SELECT * FROM users WHERE age IS NULL")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IS NULL", int64(0), nil).Return().Once()

	err := s.query.WhereNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereRaw() {
	var users []TestUser

	s.mockDriver.EXPECT().Config().Return(database.Config{}).Once()
	s.mockBuilder.EXPECT().Select(&users, "SELECT * FROM users WHERE age = ? or age = ?", 25, 30).Return(nil).Once()
	s.mockDriver.EXPECT().Explain("SELECT * FROM users WHERE age = ? or age = ?", 25, 30).Return("SELECT * FROM users WHERE age = 25 or age = 30")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 or age = 30", int64(0), nil).Return().Once()

	err := s.query.WhereRaw("age = ? or age = ?", []any{25, 30}).Get(&users)
	s.Nil(err)
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
