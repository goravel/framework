package gorm

import (
	"testing"

	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

type ToSqlTestSuite struct {
	suite.Suite
	mockLog *mockslog.Log
	query   ormcontract.Query
}

func TestToSqlTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &ToSqlTestSuite{})
}

func (s *ToSqlTestSuite) SetupSuite() {
	postgresDocker := docker.Postgres()
	s.Require().NoError(postgresDocker.Ready())

	postgresQuery := NewTestQuery(postgresDocker)
	postgresQuery.CreateTable(TestTableUsers)

	s.query = postgresQuery.Query()
}

func (s *ToSqlTestSuite) SetupTest() {
	s.mockLog = mockslog.NewLog(s.T())
}

func (s *ToSqlTestSuite) TestCount() {
	toSql := NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())
}

func (s *ToSqlTestSuite) TestCreate() {
	user := User{Name: "to_sql_create"}
	toSql := NewToSql(s.query.(*Query), s.mockLog, false)
	s.Equal("INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES ($1,$2,$3,$4,$5,$6) RETURNING \"id\"", toSql.Create(&user))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	s.Contains(toSql.Create(&user), "INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES (")
	s.Contains(toSql.Create(&user), ",NULL,'to_sql_create',NULL,'')")

	var users []User
	s.NoError(s.query.Where("name", "to_sql_create").Get(&users))
	s.Len(users, 0)
}

func (s *ToSqlTestSuite) TestDelete() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"deleted_at\"=$1 WHERE \"id\" = $2 AND \"users\".\"deleted_at\" IS NULL", toSql.Delete(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = $1", toSql.Delete(&Role{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	sql := toSql.Delete(&User{})
	s.Contains(sql, "UPDATE \"users\" SET \"deleted_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = 1", toSql.Delete(&Role{}))
}

func (s *ToSqlTestSuite) TestFind() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}))

	toSql = NewToSql(s.query.(*Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"users\".\"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}, 1))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}))

	toSql = NewToSql(s.query.(*Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"users\".\"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}, 1))
}

func (s *ToSqlTestSuite) TestFirst() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL ORDER BY \"users\".\"id\" LIMIT $2", toSql.First(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL ORDER BY \"users\".\"id\" LIMIT 1", toSql.First(&User{}))
}

func (s *ToSqlTestSuite) TestForceDelete() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("DELETE FROM \"users\" WHERE \"id\" = $1", toSql.ForceDelete(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = $1", toSql.ForceDelete(&Role{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("DELETE FROM \"users\" WHERE \"id\" = 1", toSql.ForceDelete(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = 1", toSql.ForceDelete(&Role{}))
}

func (s *ToSqlTestSuite) TestGet() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Get([]User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Get([]User{}))
}

func (s *ToSqlTestSuite) TestInvalidModel() {
	s.mockLog.EXPECT().Errorf("failed to get sql: %v", errors.OrmQueryInvalidModel.Args("")).Once()
	toSql := NewToSql(s.query.Model("invalid").Where("id", 1).(*Query), s.mockLog, false)
	s.Empty(toSql.Get([]User{}))
}

func (s *ToSqlTestSuite) TestPluck() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT \"id\" FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Pluck("id", User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT \"id\" FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Pluck("id", User{}))
}

func (s *ToSqlTestSuite) TestSave() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES ($1,$2,$3,$4,$5,$6) RETURNING \"id\"", toSql.Save(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	sql := toSql.Save(&User{})
	s.Contains(sql, "INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES (")
	s.Contains(sql, ",NULL,'',NULL,'')")
}

func (s *ToSqlTestSuite) TestSum() {
	toSql := NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("SELECT SUM(id) FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Sum("id", User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*Query), s.mockLog, true)
	s.Equal("SELECT SUM(id) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Sum("id", User{}))
}

func (s *ToSqlTestSuite) TestUpdate() {
	toSql := NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"name\"=$1,\"updated_at\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update("name", "goravel"))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	sql := toSql.Update("name", "goravel")
	s.Contains(sql, "UPDATE \"users\" SET \"name\"='goravel',\"updated_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, false)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"name\"=$1,\"updated_at\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update(map[string]any{
		"name": "goravel",
	}))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	sql = toSql.Update(map[string]any{
		"name": "goravel",
	})
	s.Contains(sql, "UPDATE \"users\" SET \"name\"='goravel',\"updated_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"updated_at\"=$1,\"name\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update(User{
		Name: "goravel",
	}))

	toSql = NewToSql(s.query.Model(&User{}).Where("id", 1).(*Query), s.mockLog, true)
	sql = toSql.Update(User{
		Name: "goravel",
	})
	s.Contains(sql, "UPDATE \"users\" SET \"updated_at\"=")
	s.Contains(sql, ",\"name\"='goravel' WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")
}
