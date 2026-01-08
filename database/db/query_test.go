package db

import (
	"context"
	databasesql "database/sql"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
	"github.com/goravel/framework/support/carbon"
)

// TestUser is a test model
type TestUser struct {
	ID    uint `db:"id"`
	Phone string
	Email string
	Name  string `db:"-"`
	Age   int    `db:"-"`
}

type QueryTestSuite struct {
	suite.Suite
	ctx              context.Context
	mockGrammar      *mocksdriver.Grammar
	mockLogger       *mockslogger.Logger
	mockReadBuilder  *mocksdb.Builder
	mockWriteBuilder *mocksdb.Builder
	now              *carbon.Carbon
	query            *Query
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, &QueryTestSuite{})
}

func (s *QueryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockGrammar = mocksdriver.NewGrammar(s.T())
	s.mockLogger = mockslogger.NewLogger(s.T())
	s.mockReadBuilder = mocksdb.NewBuilder(s.T())
	s.mockWriteBuilder = mocksdb.NewBuilder(s.T())
	s.now = carbon.Now()
	carbon.SetTestNow(s.now)

	s.query = NewQuery(s.ctx, s.mockReadBuilder, s.mockWriteBuilder, s.mockGrammar, s.mockLogger, "users", nil)
}

func (s *QueryTestSuite) TestAddWhere() {
	query := &Query{}
	query = query.addWhere(driver.Where{
		Query: "name",
		Args:  []any{"test"},
	}).(*Query)
	query = query.addWhere(driver.Where{
		Query: "name1",
		Args:  []any{"test1"},
	}).(*Query)
	query = query.addWhere(driver.Where{
		Query: "name2",
		Args:  []any{"test2"},
	}).(*Query)
	query1 := query.addWhere(driver.Where{
		Query: "name3",
		Args:  []any{"test3"},
	}).(*Query)

	s.Equal([]driver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
	}, query.conditions.Where)

	s.Equal([]driver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name3", Args: []any{"test3"}},
	}, query1.conditions.Where)

	query2 := query.addWhere(driver.Where{
		Query: "name4",
		Args:  []any{"test4"},
	}).(*Query)

	s.Equal([]driver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name4", Args: []any{"test4"}},
	}, query2.conditions.Where)

	s.Equal([]driver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
	}, query.conditions.Where)

	s.Equal([]driver.Where{
		{Query: "name", Args: []any{"test"}},
		{Query: "name1", Args: []any{"test1"}},
		{Query: "name2", Args: []any{"test2"}},
		{Query: "name3", Args: []any{"test3"}},
	}, query1.conditions.Where)
}

func (s *QueryTestSuite) TestCount() {
	s.Run("without select", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		count, err := s.query.Where("name", "John").Count()
		s.NoError(err)
		s.Equal(int64(1), count)
	})

	s.Run("with select - one column", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(name) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(name) FROM users WHERE name = ?", "John").Return("SELECT COUNT(name) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(name) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		count, err := s.query.Select("name").Where("name", "John").Count()
		s.NoError(err)
		s.Equal(int64(1), count)
	})

	s.Run("with select - one column with rename", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		count, err := s.query.Select("name as name").Where("name", "John").Count()
		s.NoError(err)
		s.Equal(int64(1), count)
	})

	s.Run("with select - multiple columns", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		count, err := s.query.Select("name", "avatar").Where("name", "John").Count()
		s.NoError(err)
		s.Equal(int64(1), count)
	})
}

func (s *QueryTestSuite) TestCrossJoin() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users CROSS JOIN posts as p WHERE age = ?", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users CROSS JOIN posts as p WHERE age = ?", 25).Return("SELECT * FROM users CROSS JOIN posts as p WHERE age = 25").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users CROSS JOIN posts as p WHERE age = 25", int64(0), nil).Return().Once()

	err := s.query.CrossJoin("posts as p").Where("age", 25).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestDecrement() {
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	update := map[string]any{"age": sq.Expr("age - ?", uint64(1))}

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()
	s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET age = age - ? WHERE name = ?", uint64(1), "John").Return(mockResult, nil).Once()
	s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET age = age - ? WHERE name = ?", uint64(1), "John").Return("UPDATE users SET age = age - 1 WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = age - 1 WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Decrement("age")
	s.NoError(err)

	mockResult.AssertExpectations(s.T())
}

func (s *QueryTestSuite) TestDelete() {
	s.Run("success", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("failed to exec", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(nil, assert.AnError).Once()
		s.mockWriteBuilder.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		_, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Equal(assert.AnError, err)
	})

	s.Run("failed to get rows affected", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(0), assert.AnError).Once()

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("DELETE FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("DELETE FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "DELETE FROM users WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()

		_, err := s.query.Where("name", "John").Where("id", 1).Delete()
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestDistinct() {
	s.Run("without column", func() {
		var users TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &users, "SELECT DISTINCT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT DISTINCT * FROM users WHERE name = ?", "John").Return("SELECT DISTINCT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT DISTINCT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Distinct().First(&users)
		s.NoError(err)
	})

	s.Run("with one column", func() {
		var users TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &users, "SELECT DISTINCT name FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT DISTINCT name FROM users WHERE name = ?", "John").Return("SELECT DISTINCT name FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT DISTINCT name FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Distinct("name").First(&users)
		s.NoError(err)
	})

	s.Run("with multiple columns", func() {
		var users TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &users, "SELECT DISTINCT name, age FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT DISTINCT name, age FROM users WHERE name = ?", "John").Return("SELECT DISTINCT name, age FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT DISTINCT name, age FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Distinct("name", "age").First(&users)
		s.NoError(err)
	})

	s.Run("Count - without column", func() {
		count, err := s.query.Where("name", "John").Distinct().Count()
		s.Equal(errors.DatabaseCountDistinctWithoutColumns, err)
		s.Equal(int64(0), count)
	})

	s.Run("Count - with one column", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(DISTINCT name) FROM users WHERE name = ?", "John").RunAndReturn(func(ctx context.Context, i1 interface{}, s string, i2 ...interface{}) error {
			destCount := i1.(*int64)
			*destCount = 1
			return nil
		}).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(DISTINCT name) FROM users WHERE name = ?", "John").Return("SELECT COUNT(DISTINCT name) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(DISTINCT name) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		res, err := s.query.Where("name", "John").Distinct("name").Count()
		s.NoError(err)
		s.Equal(int64(1), res)
	})

	s.Run("Count - with one column and rename", func() {
		res, err := s.query.Where("name", "John").Distinct("name as name").Count()
		s.Equal(errors.DatabaseCountDistinctWithoutColumns, err)
		s.Equal(int64(0), res)
	})

	s.Run("Count - with multiple columns", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").RunAndReturn(func(ctx context.Context, i1 interface{}, s string, i2 ...interface{}) error {
			destCount := i1.(*int64)
			*destCount = 1
			return nil
		}).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		res, err := s.query.Where("name", "John").Distinct("name", "age").Count()
		s.NoError(err)
		s.Equal(int64(1), res)
	})
}

func (s *QueryTestSuite) TestExists() {
	var count int64

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
		destCount := dest.(*int64)
		*destCount = 1
	}).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

	exists, err := s.query.Where("name", "John").Exists()
	s.NoError(err)
	s.True(exists)
}

func (s *QueryTestSuite) TestFind() {
	s.Run("single ID", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("SELECT * FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Find(&user, 1)

		s.NoError(err)
	})

	s.Run("multiple ID", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Return("SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))", int64(2), nil).Return().Once()

		err := s.query.Where("name", "John").Find(&users, []int{1, 2})

		s.NoError(err)
	})

	s.Run("primary key is not id", func() {
		var users TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return("SELECT * FROM users WHERE (name = \"John\" AND uuid = \"123\")").Once()
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

func (s *QueryTestSuite) TestFindOrFail() {
	s.Run("single ID", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("SELECT * FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").FindOrFail(&user, 1)

		s.NoError(err)
	})

	s.Run("multiple ID", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id IN (?,?))", "John", 1, 2).Return("SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id IN (1,2))", int64(2), nil).Return().Once()

		err := s.query.Where("name", "John").FindOrFail(&users, []int{1, 2})

		s.NoError(err)
	})

	s.Run("primary key is not id", func() {
		var users TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND uuid = ?)", "John", "123").Return("SELECT * FROM users WHERE (name = \"John\" AND uuid = \"123\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND uuid = \"123\")", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").FindOrFail(&users, "uuid", "123")

		s.NoError(err)
	})

	s.Run("invalid argument number", func() {
		var users []TestUser

		err := s.query.Where("name", "John").FindOrFail(&users, 1, 2, 3)
		s.Equal(errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
	})

	s.Run("record not found", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return(databasesql.ErrNoRows).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND id = ?)", "John", 1).Return("SELECT * FROM users WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND id = 1)", int64(-1), databasesql.ErrNoRows).Return().Once()

		err := s.query.Where("name", "John").FindOrFail(&user, 1)

		s.Equal(databasesql.ErrNoRows, err)
	})
}

func (s *QueryTestSuite) TestFirst() {
	s.Run("success", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Nil(err)
	})

	s.Run("failed to get", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(assert.AnError).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Equal(assert.AnError, err)
	})

	s.Run("no rows", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").First(&user)

		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestFirstOr() {
	var user TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").FirstOr(&user, func() error {
		return errors.New("no rows")
	})

	s.Equal(errors.New("no rows"), err)
}

func (s *QueryTestSuite) TestFirstOrFail() {
	s.Run("success", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Nil(err)
	})

	s.Run("failed to get", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(assert.AnError).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Equal(assert.AnError, err)
	})

	s.Run("no rows", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(databasesql.ErrNoRows).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(-1), databasesql.ErrNoRows).Return().Once()

		err := s.query.Where("name", "John").FirstOrFail(&user)

		s.Equal(databasesql.ErrNoRows, err)
	})
}

func (s *QueryTestSuite) TestGet() {
	s.Run("success", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ?", 25).Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ?", 25).Return("SELECT * FROM users WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25", int64(2), nil).Return().Once()

		err := s.query.Where("age", 25).Get(&users)
		s.Nil(err)
		s.mockReadBuilder.AssertExpectations(s.T())
	})

	s.Run("failed to get", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ?", 25).Return(assert.AnError).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ?", 25).Return("SELECT * FROM users WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25", int64(-1), assert.AnError).Return().Once()

		err := s.query.Where("age", 25).Get(&users)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestGroupBy_Having() {
	s.Run("With GroupBy and Having", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? GROUP BY name HAVING name = ? ORDER BY name ASC", 25, "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? GROUP BY name HAVING name = ? ORDER BY name ASC", 25, "John").Return("SELECT * FROM users WHERE age = 25 GROUP BY name HAVING name = \"John\" ORDER BY name ASC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 GROUP BY name HAVING name = \"John\" ORDER BY name ASC", int64(2), nil).Return().Once()

		err := s.query.Where("age", 25).GroupBy("name").Having("name = ?", "John").OrderBy("name").Get(&users)
		s.Nil(err)
		s.mockReadBuilder.AssertExpectations(s.T())
	})

	s.Run("Only GroupBy", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? GROUP BY name ORDER BY name ASC", 25).Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? GROUP BY name ORDER BY name ASC", 25).Return("SELECT * FROM users WHERE age = 25 GROUP BY name ORDER BY name ASC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 GROUP BY name ORDER BY name ASC", int64(2), nil).Return().Once()

		err := s.query.Where("age", 25).GroupBy("name").OrderBy("name").Get(&users)
		s.Nil(err)
	})

	s.Run("Only Having", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? ORDER BY name ASC", 25).Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY name ASC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY name ASC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY name ASC", int64(2), nil).Return().Once()

		err := s.query.Where("age", 25).Having("name = ?", "John").OrderBy("name").Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestIncrement() {
	mockResult := &MockResult{}
	mockResult.On("RowsAffected").Return(int64(1), nil)

	update := map[string]any{"age": sq.Expr("age + ?", uint64(1))}

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET age = age + ? WHERE name = ?", uint64(1), "John").Return(mockResult, nil).Once()
	s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET age = age + ? WHERE name = ?", uint64(1), "John").Return("UPDATE users SET age = age + 1 WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = age + 1 WHERE name = \"John\"", int64(1), nil).Return().Once()
	s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

	err := s.query.Where("name", "John").Increment("age")
	s.NoError(err)

	mockResult.AssertExpectations(s.T())
}

func (s *QueryTestSuite) TestInRandomOrder() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockGrammar.EXPECT().CompileInRandomOrder(mock.Anything, mock.Anything).RunAndReturn(func(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
		conditions.OrderBy = []string{"RAND()"}
		return builder
	}).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users ORDER BY RAND()").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users ORDER BY RAND()").Return("SELECT * FROM users ORDER BY RAND()").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users ORDER BY RAND()", int64(0), nil).Return().Once()

	err := s.query.InRandomOrder().Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestInsert() {
	s.Run("empty", func() {
		result, err := s.query.Insert(nil)
		s.Equal(errors.DatabaseDataIsEmpty, err)
		s.Nil(result)
	})

	s.Run("single struct", func() {
		user := TestUser{
			ID:   1,
			Name: "John",
			Age:  25,
		}

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (id) VALUES (?)", uint(1)).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (id) VALUES (?)", uint(1)).Return("INSERT INTO users (id) VALUES (1)").Once()
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (id) VALUES (?),(?)", uint(1), uint(2)).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (id) VALUES (?),(?)", uint(1), uint(2)).Return("INSERT INTO users (id) VALUES (1),(2)").Once()
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (age,id,name) VALUES (?,?,?)", 25, 1, "John").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (age,id,name) VALUES (?,?,?)", 25, 1, "John").Return("INSERT INTO users (age,id,name) VALUES (25,1,\"John\")").Once()
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (age,id,name) VALUES (?,?,?),(?,?,?)", 25, 1, "John", 30, 2, "Jane").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (age,id,name) VALUES (?,?,?),(?,?,?)", 25, 1, "John", 30, 2, "Jane").Return("INSERT INTO users (age,id,name) VALUES (25,1,\"John\"),(30,2,\"Jane\")").Once()
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (id) VALUES (?)", uint(1)).Return(nil, assert.AnError).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (id) VALUES (?)", uint(1)).Return("INSERT INTO users (id) VALUES (1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (id) VALUES (1)", int64(-1), assert.AnError).Return().Once()

		result, err := s.query.Insert(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestInsertGetID() {
	s.Run("empty", func() {
		id, err := s.query.InsertGetID(nil)
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (age,name) VALUES (?,?)", 25, "John").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (age,name) VALUES (?,?)", 25, "John").Return("INSERT INTO users (age,name) VALUES (25,\"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (age,name) VALUES (25,\"John\")", int64(1), nil).Return().Once()

		id, err := s.query.InsertGetID(user)
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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (id) VALUES (?)", uint(1)).Return(nil, assert.AnError).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (id) VALUES (?)", uint(1)).Return("INSERT INTO users (id) VALUES (1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (id) VALUES (1)", int64(-1), assert.AnError).Return().Once()

		id, err := s.query.InsertGetID(user)
		s.Equal(int64(0), id)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestJoin() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return("SELECT * FROM users JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25", int64(0), nil).Return().Once()

	err := s.query.Join("posts as p ON users.id = p.user_id AND p.id = ?", 1).Where("age", 25).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestLatest() {
	s.Run("default column", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE age = ? ORDER BY created_at DESC", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY created_at DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY created_at DESC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY created_at DESC", int64(1), nil).Return().Once()

		err := s.query.Where("age", 25).Latest().First(&user)
		s.Nil(err)
	})

	s.Run("custom column", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE age = ? ORDER BY name DESC", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY name DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY name DESC").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY name DESC", int64(1), nil).Return().Once()

		err := s.query.Where("age", 25).Latest("name").First(&user)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestLeftJoin() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users LEFT JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users LEFT JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return("SELECT * FROM users LEFT JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users LEFT JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25", int64(0), nil).Return().Once()

	err := s.query.LeftJoin("posts as p ON users.id = p.user_id AND p.id = ?", 1).Where("age", 25).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestLockForUpdate() {
	s.Run("FOR UPDATE", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		s.mockGrammar.EXPECT().CompileLockForUpdate(mock.Anything, mock.Anything).RunAndReturn(func(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
			return builder.Suffix("FOR UPDATE")
		}).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? FOR UPDATE", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? FOR UPDATE", 25).Return("SELECT * FROM users WHERE age = 25 FOR UPDATE").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 FOR UPDATE", int64(0), nil).Return().Once()

		err := s.query.Where("age", 25).LockForUpdate().Get(&users)
		s.Nil(err)
	})

	s.Run("WITH (ROWLOCK, UPDLOCK, HOLDLOCK)", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		s.mockGrammar.EXPECT().CompileLockForUpdate(mock.Anything, mock.Anything).RunAndReturn(func(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
			return builder.From(conditions.Table + " WITH (ROWLOCK, UPDLOCK, HOLDLOCK)")
		}).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WITH (ROWLOCK, UPDLOCK, HOLDLOCK) WHERE age = ?", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WITH (ROWLOCK, UPDLOCK, HOLDLOCK) WHERE age = ?", 25).Return("SELECT * FROM users WITH (ROWLOCK, UPDLOCK, HOLDLOCK) WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WITH (ROWLOCK, UPDLOCK, HOLDLOCK) WHERE age = 25", int64(0), nil).Return().Once()

		err := s.query.Where("age", 25).LockForUpdate().Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestLimit() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? LIMIT 1", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? LIMIT 1", 25).Return("SELECT * FROM users WHERE age = 25 LIMIT 1").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 LIMIT 1", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).Limit(1).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOffset() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? OFFSET 1", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? OFFSET 1", 25).Return("SELECT * FROM users WHERE age = 25 OFFSET 1").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 OFFSET 1", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).Offset(1).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrderBy() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? ORDER BY age ASC, id ASC", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY age ASC, id ASC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id ASC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id ASC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("age").OrderBy("id").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrderByDesc() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? ORDER BY age ASC, id DESC", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY age ASC, id DESC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id DESC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY age ASC, id DESC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("age").OrderByDesc("id").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrderByRaw() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? ORDER BY name ASC, age DESC, id ASC", 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? ORDER BY name ASC, age DESC, id ASC", 25).Return("SELECT * FROM users WHERE age = 25 ORDER BY name ASC, age DESC, id ASC").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 ORDER BY name ASC, age DESC, id ASC", int64(0), nil).Return().Once()

	err := s.query.Where("age", 25).OrderBy("name").OrderByRaw("age DESC, id ASC").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhere() {
	now := carbon.Now()
	carbon.SetTestNow(now)

	s.Run("simple condition", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (((name = ? AND age = ?) OR age IN (?,?)) OR name = ?)", "John", 25, 30, 40, "Jane").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (((name = ? AND age = ?) OR age IN (?,?)) OR name = ?)", "John", 25, 30, 40, "Jane").Return("SELECT * FROM users WHERE (((name = \"John\" AND age = 25) OR age IN (30,40)) OR name = \"Jane\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, now, "SELECT * FROM users WHERE (((name = \"John\" AND age = 25) OR age IN (30,40)) OR name = \"Jane\")", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Where("age", 25).OrWhere("age", []int{30, 40}).OrWhere("name", "Jane").First(&user)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age > ?)", "John", 18).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age > ?)", "John", 18).Return("SELECT * FROM users WHERE (name = \"John\" OR age > 18)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age > 18)", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").OrWhere("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR ((age IN (?,?) AND name = ?) OR age = ?))", "John", 25, 30, "Tom", 40).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR ((age IN (?,?) AND name = ?) OR age = ?))", "John", 25, 30, "Tom", 40).Return("SELECT * FROM users WHERE (name = \"John\" OR ((age IN (25,30) AND name = \"Tom\") OR age = 40))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR ((age IN (25,30) AND name = \"Tom\") OR age = 40))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").OrWhere(func(query db.Query) db.Query {
			return query.Where("age", []int{25, 30}).Where("name", "Tom").OrWhere("age", 40)
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestOrWhereBetween() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age BETWEEN ? AND ?)", "John", 18, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age BETWEEN ? AND ?)", "John", 18, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age BETWEEN 18 AND 30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age BETWEEN 18 AND 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereColumn() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR height = weight)", "John").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR height = weight)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR height = weight)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR height = weight)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereColumn("height", "weight").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereIn() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age IN (?,?))", "John", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IN (?,?))", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age IN (25,30))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IN (25,30))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereLike() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR name LIKE ?)", "John", "%John%").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR name LIKE ?)", "John", "%John%").Return("SELECT * FROM users WHERE (name = \"John\" OR name LIKE \"%John%\")").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR name LIKE \"%John%\")", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNot() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR NOT (name = ?))", "John", "Jane").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR NOT (name = ?))", "John", "Jane").Return("SELECT * FROM users WHERE (name = \"John\" OR NOT (name = \"Jane\"))")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR NOT (name = \"Jane\"))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNot("name", "Jane").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotBetween() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age NOT BETWEEN ? AND ?)", "John", 18, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age NOT BETWEEN ? AND ?)", "John", 18, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age NOT BETWEEN 18 AND 30)")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age NOT BETWEEN 18 AND 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotIn() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age NOT IN (?,?))", "John", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age NOT IN (?,?))", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age NOT IN (25,30))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age NOT IN (25,30))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotLike() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR name NOT LIKE ?)", "John", "%John%").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR name NOT LIKE ?)", "John", "%John%").Return("SELECT * FROM users WHERE (name = \"John\" OR name NOT LIKE \"%John%\")").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR name NOT LIKE \"%John%\")", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNotNull() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age IS NOT NULL)", "John").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IS NOT NULL)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR age IS NOT NULL)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IS NOT NULL)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNotNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereNull() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age IS NULL)", "John").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age IS NULL)", "John").Return("SELECT * FROM users WHERE (name = \"John\" OR age IS NULL)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age IS NULL)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestOrWhereRaw() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? OR age = ? or age = ?)", "John", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? OR age = ? or age = ?)", "John", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" OR age = 25 OR age = 30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" OR age = 25 OR age = 30)", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").OrWhereRaw("age = ? or age = ?", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestPaginate() {
	s.Run("without Select", func() {
		var users []TestUser
		var total int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Twice()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &total, "SELECT COUNT(*) FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destTotal := dest.(*int64)
			*destTotal = 2
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destUsers := dest.(*[]TestUser)
			*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").Return("SELECT * FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0", int64(2), nil).Return().Once()

		err := s.query.Where("name", "John").Paginate(1, 10, &users, &total)
		s.Nil(err)
		s.Equal(int64(2), total)
		s.Equal(2, len(users))
	})

	s.Run("with Select - one column", func() {
		var users []TestUser
		var total int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Twice()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &total, "SELECT COUNT(name) FROM users WHERE name = ?", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destTotal := dest.(*int64)
				*destTotal = 2
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(name) FROM users WHERE name = ?", "John").Return("SELECT COUNT(name) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(name) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT name FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destUsers := dest.(*[]TestUser)
				*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT name FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").Return("SELECT name FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0", int64(2), nil).Return().Once()

		err := s.query.Select("name").Where("name", "John").Paginate(1, 10, &users, &total)
		s.Nil(err)
		s.Equal(int64(2), total)
		s.Equal(2, len(users))
	})

	s.Run("with Select - one column with rename", func() {
		var users []TestUser
		var total int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Twice()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &total, "SELECT COUNT(*) FROM users WHERE name = ?", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destTotal := dest.(*int64)
				*destTotal = 2
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT name as name FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destUsers := dest.(*[]TestUser)
				*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT name as name FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").Return("SELECT name as name FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name as name FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0", int64(2), nil).Return().Once()

		err := s.query.Select("name as name").Where("name", "John").Paginate(1, 10, &users, &total)
		s.Nil(err)
		s.Equal(int64(2), total)
		s.Equal(2, len(users))
	})

	s.Run("with Select - multiple columns", func() {
		var users []TestUser
		var total int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Twice()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &total, "SELECT COUNT(*) FROM users WHERE name = ?", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destTotal := dest.(*int64)
				*destTotal = 2
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE name = \"John\"", int64(-1), nil).Return().Once()

		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT name, age FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").
			Run(func(ctx context.Context, dest any, query string, args ...any) {
				destUsers := dest.(*[]TestUser)
				*destUsers = []TestUser{{ID: 1, Name: "John", Age: 25}, {ID: 2, Name: "Jane", Age: 30}}
			}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT name, age FROM users WHERE name = ? LIMIT 10 OFFSET 0", "John").Return("SELECT name, age FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name, age FROM users WHERE name = \"John\" LIMIT 10 OFFSET 0", int64(2), nil).Return().Once()

		err := s.query.Select("name", "age").Where("name", "John").Paginate(1, 10, &users, &total)
		s.Nil(err)
		s.Equal(int64(2), total)
		s.Equal(2, len(users))
	})
}

func (s *QueryTestSuite) TestPluck() {
	var names []string

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &names, "SELECT name FROM users WHERE name = ?", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
		destNames := dest.(*[]string)
		*destNames = []string{"John"}
	}).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT name FROM users WHERE name = ?", "John").Return("SELECT name FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Pluck("name", &names)
	s.NoError(err)
	s.Equal([]string{"John"}, names)
}

func (s *QueryTestSuite) TestRightJoin() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users RIGHT JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users RIGHT JOIN posts as p ON users.id = p.user_id AND p.id = ? WHERE age = ?", 1, 25).Return("SELECT * FROM users RIGHT JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users RIGHT JOIN posts as p ON users.id = p.user_id AND p.id = 1 WHERE age = 25", int64(0), nil).Return().Once()

	err := s.query.RightJoin("posts as p ON users.id = p.user_id AND p.id = ?", 1).Where("age", 25).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestSelect() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT id, name FROM users WHERE name = ?", "John").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT id, name FROM users WHERE name = ?", "John").Return("SELECT id, name FROM users WHERE name = \"John\"").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT id, name FROM users WHERE name = \"John\"", int64(0), nil).Return().Once()

	err := s.query.Select("id", "name").Where("name", "John").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestSharedLock() {
	s.Run("FOR SHARE", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		s.mockGrammar.EXPECT().CompileSharedLock(mock.Anything, mock.Anything).RunAndReturn(func(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
			return builder.Suffix("FOR SHARE")
		}).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? FOR SHARE", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? FOR SHARE", 25).Return("SELECT * FROM users WHERE age = 25 FOR SHARE").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age = 25 FOR SHARE", int64(0), nil).Return().Once()

		err := s.query.Where("age", 25).SharedLock().Get(&users)
		s.Nil(err)
	})

	s.Run("WITH (ROWLOCK, HOLDLOCK)", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		s.mockGrammar.EXPECT().CompileSharedLock(mock.Anything, mock.Anything).RunAndReturn(func(builder sq.SelectBuilder, conditions *driver.Conditions) sq.SelectBuilder {
			return builder.From(conditions.Table + " WITH (ROWLOCK, HOLDLOCK)")
		}).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WITH (ROWLOCK, HOLDLOCK) WHERE age = ?", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WITH (ROWLOCK, HOLDLOCK) WHERE age = ?", 25).Return("SELECT * FROM users WITH (ROWLOCK, HOLDLOCK) WHERE age = 25").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WITH (ROWLOCK, HOLDLOCK) WHERE age = 25", int64(0), nil).Return().Once()

		err := s.query.Where("age", 25).SharedLock().Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestSum() {
	var sum int64

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().GetContext(s.ctx, &sum, "SELECT SUM(age) FROM users WHERE age = ?", 25).Run(func(ctx context.Context, dest any, query string, args ...any) {
		destSum := dest.(*int64)
		*destSum = 25
	}).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT SUM(age) FROM users WHERE age = ?", 25).Return("SELECT SUM(age) FROM users WHERE age = 25").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT SUM(age) FROM users WHERE age = 25", int64(1), nil).Return().Once()

	sum, err := s.query.Where("age", 25).Sum("age")
	s.Nil(err)
	s.Equal(int64(25), sum)
}

func (s *QueryTestSuite) TestToSql() {
	s.Run("Count", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Times(7)

		sql := s.query.Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = ?", sql)

		s.mockLogger.EXPECT().Errorf(s.ctx, "failed to get sql: cannot use Count with Distinct without specifying columns").Once()
		sql = s.query.Distinct().Where("name", "John").ToSql().Count()
		s.Empty(sql)

		sql = s.query.Distinct("name").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(DISTINCT name) FROM users WHERE name = ?", sql)

		sql = s.query.Distinct("name", "avatar").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = ?", sql)

		sql = s.query.Select("name", "avatar").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = ?", sql)

		sql = s.query.Select("name as n").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = ?", sql)

		sql = s.query.Select("name n").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = ?", sql)

		sql = s.query.Select("name").Where("name", "John").ToSql().Count()
		s.Equal("SELECT COUNT(name) FROM users WHERE name = ?", sql)
	})

	s.Run("Delete", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		sql := s.query.Where("name", "John").ToSql().Delete()
		s.Equal("DELETE FROM users WHERE name = ?", sql)
	})

	s.Run("First", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		sql := s.query.Where("name", "John").ToSql().First()
		s.Equal("SELECT * FROM users WHERE name = ?", sql)
	})

	s.Run("Get", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		sql := s.query.Where("name", "John").ToSql().Get()
		s.Equal("SELECT * FROM users WHERE name = ?", sql)
	})

	s.Run("Insert", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Times(4)

		sql := s.query.Where("name", "John").ToSql().Insert(map[string]any{"name": "John"})
		s.Equal("INSERT INTO users (name) VALUES (?)", sql)

		sql = s.query.Where("name", "John").ToSql().Insert([]map[string]any{{"name": "John"}, {"name": "Jane"}})
		s.Equal("INSERT INTO users (name) VALUES (?),(?)", sql)

		sql = s.query.Where("name", "John").ToSql().Insert(TestUser{Phone: "1234567890"})
		s.Equal("INSERT INTO users (phone) VALUES (?)", sql)

		sql = s.query.Where("name", "John").ToSql().Insert([]TestUser{{Phone: "1234567890"}, {Phone: "1234567891"}})
		s.Equal("INSERT INTO users (phone) VALUES (?),(?)", sql)
	})

	s.Run("Pluck", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()

		sql := s.query.Where("name", "John").ToSql().Pluck("name", &[]string{})
		s.Equal("SELECT name FROM users WHERE name = ?", sql)
	})

	s.Run("Update", func() {
		update := map[string]any{"name": "Jane"}

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Times(3)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Twice()

		sql := s.query.Where("name", "John").ToSql().Update(update)
		s.Equal("UPDATE users SET name = ? WHERE name = ?", sql)

		sql = s.query.Where("name", "John").ToSql().Update("name", "Jane")
		s.Equal("UPDATE users SET name = ? WHERE name = ?", sql)

		update = map[string]any{"phone": "1234567890"}
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()
		sql = s.query.Where("name", "John").ToSql().Update(TestUser{Phone: "1234567890"})
		s.Equal("UPDATE users SET phone = ? WHERE name = ?", sql)
	})
}

func (s *QueryTestSuite) TestToRawSql() {
	s.Run("Count", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE name = ?", "John").Return("SELECT COUNT(*) FROM users WHERE name = \"John\"").Once()

		sql := s.query.Where("name", "John").ToRawSql().Count()
		s.Equal("SELECT COUNT(*) FROM users WHERE name = \"John\"", sql)
	})

	s.Run("Delete", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("DELETE FROM users WHERE name = ?", "John").Return("DELETE FROM users WHERE name = \"John\"").Once()

		sql := s.query.Where("name", "John").ToRawSql().Delete()
		s.Equal("DELETE FROM users WHERE name = \"John\"", sql)
	})

	s.Run("First", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()

		sql := s.query.Where("name", "John").ToRawSql().First()
		s.Equal("SELECT * FROM users WHERE name = \"John\"", sql)
	})

	s.Run("Get", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()

		sql := s.query.Where("name", "John").ToRawSql().Get()
		s.Equal("SELECT * FROM users WHERE name = \"John\"", sql)
	})

	s.Run("Insert", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Times(4)

		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (name) VALUES (?)", "John").Return("INSERT INTO users (name) VALUES (\"John\")").Once()
		sql := s.query.Where("name", "John").ToRawSql().Insert(map[string]any{"name": "John"})
		s.Equal("INSERT INTO users (name) VALUES (\"John\")", sql)

		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (name) VALUES (?),(?)", "John", "Jane").Return("INSERT INTO users (name) VALUES (\"John\"),(\"Jane\")").Once()
		sql = s.query.Where("name", "John").ToRawSql().Insert([]map[string]any{{"name": "John"}, {"name": "Jane"}})
		s.Equal("INSERT INTO users (name) VALUES (\"John\"),(\"Jane\")", sql)

		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (phone) VALUES (?)", "1234567890").Return("INSERT INTO users (phone) VALUES (\"1234567890\")").Once()
		sql = s.query.Where("name", "John").ToRawSql().Insert(TestUser{Phone: "1234567890"})
		s.Equal("INSERT INTO users (phone) VALUES (\"1234567890\")", sql)

		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (phone) VALUES (?),(?)", "1234567890", "1234567891").Return("INSERT INTO users (phone) VALUES (\"1234567890\"),(\"1234567891\")").Once()
		sql = s.query.Where("name", "John").ToRawSql().Insert([]TestUser{{Phone: "1234567890"}, {Phone: "1234567891"}})
		s.Equal("INSERT INTO users (phone) VALUES (\"1234567890\"),(\"1234567891\")", sql)
	})

	s.Run("Pluck", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT name FROM users WHERE name = ?", "John").Return("SELECT name FROM users WHERE name = \"John\"").Once()

		sql := s.query.Where("name", "John").ToRawSql().Pluck("name", &[]string{})
		s.Equal("SELECT name FROM users WHERE name = \"John\"", sql)
	})

	s.Run("Update", func() {
		update := map[string]any{"name": "Jane"}

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Times(3)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Twice()

		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET name = ? WHERE name = ?", "Jane", "John").Return("UPDATE users SET name = \"Jane\" WHERE name = \"John\"").Once()
		sql := s.query.Where("name", "John").ToRawSql().Update(update)
		s.Equal("UPDATE users SET name = \"Jane\" WHERE name = \"John\"", sql)

		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET name = ? WHERE name = ?", "Jane", "John").Return("UPDATE users SET name = \"Jane\" WHERE name = \"John\"").Once()
		sql = s.query.Where("name", "John").ToRawSql().Update("name", "Jane")
		s.Equal("UPDATE users SET name = \"Jane\" WHERE name = \"John\"", sql)

		update = map[string]any{"phone": "1234567890"}
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE name = ?", "1234567890", "John").Return("UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"").Once()
		sql = s.query.Where("name", "John").ToRawSql().Update(TestUser{Phone: "1234567890"})
		s.Equal("UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"", sql)
	})
}

func (s *QueryTestSuite) TestUpdate() {
	s.Run("single struct", func() {
		user := TestUser{
			Phone: "1234567890",
			Name:  "John",
			Age:   25,
		}

		update, err := convertToMap(user)
		s.Require().NoError(err)

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

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

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET age = ?, name = ?, phone = ? WHERE (name = ? AND id = ?)", 25, "John", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET age = ?, name = ?, phone = ? WHERE (name = ? AND id = ?)", 25, "John", "1234567890", "John", 1).Return("UPDATE users SET age = 25, name = \"John\", phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET age = 25, name = \"John\", phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(1), nil).Return().Once()
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(user).Return(user, nil).Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("single column", func() {
		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE name = ?", "1234567890", "John").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE name = ?", "1234567890", "John").Return("UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE name = \"John\"", int64(1), nil).Return().Once()
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(map[string]any{"phone": "1234567890"}).Return(map[string]any{"phone": "1234567890"}, nil).Once()

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

		update, err := convertToMap(user)
		s.Require().NoError(err)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(nil, assert.AnError).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

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

		update, err := convertToMap(user)
		s.Require().NoError(err)

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(0), assert.AnError).Once()

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (name = ? AND id = ?)", "1234567890", "John", 1).Return("UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (name = \"John\" AND id = 1)", int64(-1), assert.AnError).Return().Once()
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

		result, err := s.query.Where("name", "John").Where("id", 1).Update(user)
		s.Nil(result)
		s.Equal(assert.AnError, err)
	})
}

func (s *QueryTestSuite) TestUpdateOrInsert() {
	s.Run("update record with struct", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return("SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")", int64(-1), nil).Return().Once()

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return("UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")", int64(1), nil).Return().Once()

		whereUser := TestUser{
			Email: "john@example.com",
		}
		user := TestUser{
			Phone: "1234567890",
		}

		update, err := convertToMap(user)
		s.Require().NoError(err)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

		result, err := s.query.Where("id", 1).UpdateOrInsert(whereUser, user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("update record with struct and map", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return("SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")", int64(-1), nil).Return().Once()

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return("UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")", int64(1), nil).Return().Once()

		whereUser := TestUser{
			Email: "john@example.com",
		}
		user := map[string]any{
			"phone": "1234567890",
		}

		update, err := convertToMap(user)
		s.Require().NoError(err)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

		result, err := s.query.Where("id", 1).UpdateOrInsert(whereUser, user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("update record with map and struct", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return("SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")", int64(-1), nil).Return().Once()

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return("UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")", int64(1), nil).Return().Once()

		whereUser := map[string]any{
			"email": "john@example.com",
		}
		user := TestUser{
			Phone: "1234567890",
		}

		update, err := convertToMap(user)
		s.Require().NoError(err)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

		result, err := s.query.Where("id", 1).UpdateOrInsert(whereUser, user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("update record with map", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Run(func(ctx context.Context, dest any, query string, args ...any) {
			destCount := dest.(*int64)
			*destCount = 1
		}).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return("SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")", int64(-1), nil).Return().Once()

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("UPDATE users SET phone = ? WHERE (id = ? AND email = ?)", "1234567890", 1, "john@example.com").Return("UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "UPDATE users SET phone = \"1234567890\" WHERE (id = 1 AND email = \"john@example.com\")", int64(1), nil).Return().Once()

		whereUser := map[string]any{
			"email": "john@example.com",
		}
		user := map[string]any{
			"phone": "1234567890",
		}

		update, err := convertToMap(user)
		s.Require().NoError(err)
		s.mockGrammar.EXPECT().CompileJsonColumnsUpdate(update).Return(update, nil).Once()

		result, err := s.query.Where("id", 1).UpdateOrInsert(whereUser, user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})

	s.Run("insert record with struct", func() {
		var count int64

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &count, "SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT COUNT(*) FROM users WHERE (id = ? AND email = ?)", 1, "john@example.com").Return("SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT COUNT(*) FROM users WHERE (id = 1 AND email = \"john@example.com\")", int64(-1), nil).Return().Once()

		mockResult := &MockResult{}
		mockResult.On("RowsAffected").Return(int64(1), nil)

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockWriteBuilder.EXPECT().ExecContext(s.ctx, "INSERT INTO users (email,phone) VALUES (?,?)", "john@example.com", "1234567890").Return(mockResult, nil).Once()
		s.mockWriteBuilder.EXPECT().Explain("INSERT INTO users (email,phone) VALUES (?,?)", "john@example.com", "1234567890").Return("INSERT INTO users (email,phone) VALUES (\"john@example.com\", \"1234567890\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "INSERT INTO users (email,phone) VALUES (\"john@example.com\", \"1234567890\")", int64(1), nil).Return().Once()

		whereUser := TestUser{
			Email: "john@example.com",
		}
		user := TestUser{
			Phone: "1234567890",
		}
		result, err := s.query.Where("id", 1).UpdateOrInsert(whereUser, user)
		s.Nil(err)
		s.Equal(int64(1), result.RowsAffected)

		mockResult.AssertExpectations(s.T())
	})
}

func (s *QueryTestSuite) TestValue() {
	var name string

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().GetContext(s.ctx, &name, "SELECT name FROM users WHERE name = ? LIMIT 1", "John").Run(func(ctx context.Context, dest any, query string, args ...any) {
		destName := dest.(*string)
		*destName = "John"
	}).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT name FROM users WHERE name = ? LIMIT 1", "John").Return("SELECT name FROM users WHERE name = \"John\" LIMIT 1").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT name FROM users WHERE name = \"John\" LIMIT 1", int64(1), nil).Return().Once()

	err := s.query.Where("name", "John").Value("name", &name)
	s.NoError(err)
	s.Equal("John", name)
}

func (s *QueryTestSuite) TestWhen() {
	s.Run("when condition is true", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 25).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 25).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 25)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND age = 25)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").When(true, func(query db.Query) db.Query {
			return query.Where("age", 25)
		}).First(&user)
		s.Nil(err)
	})

	s.Run("when condition is false", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE name = ?", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name = ?", "John").Return("SELECT * FROM users WHERE name = \"John\"").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name = \"John\"", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").When(false, func(query db.Query) db.Query {
			return query.Where("age", 25)
		}).First(&user)
		s.Nil(err)
	})

	s.Run("when condition is false with false callback", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 30).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ?)", "John", 30).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 30)").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND age = 30)", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").When(false, func(query db.Query) db.Query {
			return query.Where("age", 25)
		}, func(query db.Query) db.Query {
			return query.Where("age", 30)
		}).First(&user)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhere() {
	s.Run("simple condition", func() {
		var user TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().GetContext(s.ctx, &user, "SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND age = ? AND age IN (?,?))", "John", 25, 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND age = 25 AND age IN (25,30))", int64(1), nil).Return().Once()

		err := s.query.Where("name", "John").Where("age", 25).Where("age", []int{25, 30}).First(&user)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age > ?", 18).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age > ?", 18).Return("SELECT * FROM users WHERE age > 18").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age > 18", int64(0), nil).Return().Once()

		err := s.query.Where("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND (age IN (?,?) AND name = ?))", "John", 25, 30, "Tom").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND (age IN (?,?) AND name = ?))", "John", 25, 30, "Tom").Return("SELECT * FROM users WHERE (name = \"John\" AND (age IN (25,30) AND name = \"Tom\"))").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND (age IN (25,30) AND name = \"Tom\"))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").Where(func(query db.Query) db.Query {
			return query.Where("age", []int{25, 30}).Where("name", "Tom")
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhereBetween() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age BETWEEN ? AND ?", 18, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age BETWEEN ? AND ?", 18, 30).Return("SELECT * FROM users WHERE age BETWEEN 18 AND 30").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age BETWEEN 18 AND 30", int64(0), nil).Return().Once()

	err := s.query.WhereBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereColumn() {
	var users []TestUser

	s.Run("simple condition", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (age = height AND name = ?)", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (age = height AND name = ?)", "John").Return("SELECT * FROM users WHERE (age = height AND name = \"John\")").Once()
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (age = height AND name = \"John\")", int64(0), nil).Return().Once()

		err := s.query.WhereColumn("age", "height").Where("name", "John").Get(&users)
		s.Nil(err)
	})

	s.Run("with operator", func() {
		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (age > height AND name = ?)", "John").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (age > height AND name = ?)", "John").Return("SELECT * FROM users WHERE (age > height AND name = \"John\")").Once()
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

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Twice()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM agents WHERE age = ?", 25).Return("SELECT * FROM agents WHERE age = 25").Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND EXISTS (SELECT * FROM agents WHERE age = 25))", "John").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND EXISTS (SELECT * FROM agents WHERE age = 25))", "John").Return("SELECT * FROM users WHERE (name = \"John\" AND EXISTS (SELECT * FROM agents WHERE age = 25))").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND EXISTS (SELECT * FROM agents WHERE age = 25))", int64(0), nil).Return().Once()

	err := s.query.Where("name", "John").WhereExists(func() db.Query {
		return NewQuery(s.ctx, s.mockReadBuilder, s.mockReadBuilder, s.mockGrammar, s.mockLogger, "agents", nil).Where("age", 25)
	}).Get(&users)
	s.Nil(err)

}

func (s *QueryTestSuite) TestWhereIn() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age IN (?,?)", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age IN (?,?)", 25, 30).Return("SELECT * FROM users WHERE age IN (25,30)").Once()
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IN (25,30)", int64(0), nil).Return().Once()

	err := s.query.WhereIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereLike() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE name LIKE ?", "%John%").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name LIKE ?", "%John%").Return("SELECT * FROM users WHERE name LIKE \"%John%\"")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name LIKE \"%John%\"", int64(0), nil).Return().Once()

	err := s.query.WhereLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNot() {
	s.Run("simple condition", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND NOT (name = ?))", "John", "Jane").Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT (name = ?))", "John", "Jane").Return("SELECT * FROM users WHERE (name = \"John\" AND NOT (name = \"Jane\"))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT (name = \"Jane\"))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot("name", "Jane").Get(&users)
		s.Nil(err)
	})

	s.Run("raw query", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND NOT (age > ?))", "John", 18).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT (age > ?))", "John", 18).Return("SELECT * FROM users WHERE (name = \"John\" AND NOT (age > 18))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT (age > 18))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot("age > ?", 18).Get(&users)
		s.Nil(err)
	})

	s.Run("nested condition", func() {
		var users []TestUser

		s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
		s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE (name = ? AND NOT ((name = ? AND age IN (?,?))))", "John", "Jane", 25, 30).Return(nil).Once()
		s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE (name = ? AND NOT ((name = ? AND age IN (?,?))))", "John", "Jane", 25, 30).Return("SELECT * FROM users WHERE (name = \"John\" AND NOT ((name = \"Jane\" AND age IN (25,30))))")
		s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE (name = \"John\" AND NOT ((name = \"Jane\" AND age IN (25,30))))", int64(0), nil).Return().Once()

		err := s.query.Where("name", "John").WhereNot(func(query db.Query) db.Query {
			return query.Where("name", "Jane").Where("age", []int{25, 30})
		}).Get(&users)
		s.Nil(err)
	})
}

func (s *QueryTestSuite) TestWhereNotBetween() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age NOT BETWEEN ? AND ?", 18, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age NOT BETWEEN ? AND ?", 18, 30).Return("SELECT * FROM users WHERE age NOT BETWEEN 18 AND 30")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age NOT BETWEEN 18 AND 30", int64(0), nil).Return().Once()

	err := s.query.WhereNotBetween("age", 18, 30).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotIn() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age NOT IN (?,?)", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age NOT IN (?,?)", 25, 30).Return("SELECT * FROM users WHERE age NOT IN (25,30)")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age NOT IN (25,30)", int64(0), nil).Return().Once()

	err := s.query.WhereNotIn("age", []any{25, 30}).Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotLike() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE name NOT LIKE ?", "%John%").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE name NOT LIKE ?", "%John%").Return("SELECT * FROM users WHERE name NOT LIKE \"%John%\"")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE name NOT LIKE \"%John%\"", int64(0), nil).Return().Once()

	err := s.query.WhereNotLike("name", "%John%").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNotNull() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age IS NOT NULL").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age IS NOT NULL").Return("SELECT * FROM users WHERE age IS NOT NULL")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IS NOT NULL", int64(0), nil).Return().Once()

	err := s.query.WhereNotNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereNull() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age IS NULL").Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age IS NULL").Return("SELECT * FROM users WHERE age IS NULL")
	s.mockLogger.EXPECT().Trace(s.ctx, s.now, "SELECT * FROM users WHERE age IS NULL", int64(0), nil).Return().Once()

	err := s.query.WhereNull("age").Get(&users)
	s.Nil(err)
}

func (s *QueryTestSuite) TestWhereRaw() {
	var users []TestUser

	s.mockGrammar.EXPECT().CompilePlaceholderFormat().Return(nil).Once()
	s.mockReadBuilder.EXPECT().SelectContext(s.ctx, &users, "SELECT * FROM users WHERE age = ? or age = ?", 25, 30).Return(nil).Once()
	s.mockReadBuilder.EXPECT().Explain("SELECT * FROM users WHERE age = ? or age = ?", 25, 30).Return("SELECT * FROM users WHERE age = 25 or age = 30")
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
