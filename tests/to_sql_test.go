package tests

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	mockslog "github.com/goravel/framework/mocks/log"
)

type ToSqlTestSuite struct {
	suite.Suite
	mockLog *mockslog.Log
	query   ormcontract.Query
}

func TestToSqlTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ToSqlTestSuite{})
}

func (s *ToSqlTestSuite) SetupSuite() {
	postgresTestQuery := NewTestQueryBuilder().Postgres("", false)
	postgresTestQuery.CreateTable(TestTableUsers)
	s.query = postgresTestQuery.Query()
}

func (s *ToSqlTestSuite) SetupTest() {
	s.mockLog = mockslog.NewLog(s.T())
}

func (s *ToSqlTestSuite) TestCount() {
	toSql := gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Distinct().Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT COUNT(DISTINCT(\"*\")) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Distinct("name").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT COUNT(DISTINCT(\"name\")) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Distinct("name", "avatar").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Select("name", "avatar").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Select("name as n").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Select("name n").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&User{}).Select("name").Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT COUNT(\"name\") FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Count())

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT count(*) FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Count())

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT count(*) FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Count())
}

func (s *ToSqlTestSuite) TestCreate() {
	user := User{Name: "to_sql_create"}

	toSql := gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES ($1,$2,$3,$4,$5,$6) RETURNING \"id\"", toSql.Create(&user))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Contains(toSql.Create(&user), "INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES (")
	s.Contains(toSql.Create(&user), ",NULL,'to_sql_create',NULL,'')")

	// global scopes
	globalScope := GlobalScope{Name: "to_sql_create"}

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	sql := toSql.Create(&globalScope)
	s.Equal(sql, "INSERT INTO \"global_scopes\" (\"created_at\",\"updated_at\",\"name\",\"deleted_at\") VALUES ($1,$2,$3,$4) RETURNING \"id\"")

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, true)
	sql = toSql.Create(&globalScope)
	s.Contains(sql, "INSERT INTO \"global_scopes\" (\"created_at\",\"updated_at\",\"name\",\"deleted_at\") VALUES (")
	s.Contains(sql, "'to_sql_create',NULL)")
	s.NotContains(sql, "WHERE")
}

func (s *ToSqlTestSuite) TestDelete() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"deleted_at\"=$1 WHERE \"id\" = $2 AND \"users\".\"deleted_at\" IS NULL", toSql.Delete(&User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = $1", toSql.Delete(&Role{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql := toSql.Delete(&User{})
	s.Contains(sql, "UPDATE \"users\" SET \"deleted_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = 1", toSql.Delete(&Role{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	sql = toSql.Delete(&GlobalScope{})
	s.Equal("UPDATE \"global_scopes\" SET \"deleted_at\"=$1 WHERE \"id\" = $2 AND \"name\" = $3 AND \"global_scopes\".\"deleted_at\" IS NULL", sql)

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql = toSql.Delete(&GlobalScope{})
	s.Contains(sql, "UPDATE \"global_scopes\" SET \"deleted_at\"='")
	s.Contains(sql, "' WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL")
}

func (s *ToSqlTestSuite) TestFind() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}))

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"users\".\"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}, 1))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}))

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"users\".\"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Find(&User{}, 1))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Find(&GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Find(&GlobalScope{}))
}

func (s *ToSqlTestSuite) TestFirst() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL ORDER BY \"users\".\"id\" LIMIT $2", toSql.First(&User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL ORDER BY \"users\".\"id\" LIMIT 1", toSql.First(&User{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL ORDER BY \"global_scopes\".\"id\" LIMIT $3", toSql.First(&GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL ORDER BY \"global_scopes\".\"id\" LIMIT 1", toSql.First(&GlobalScope{}))
}

func (s *ToSqlTestSuite) TestForceDelete() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("DELETE FROM \"users\" WHERE \"id\" = $1", toSql.ForceDelete(&User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = $1", toSql.ForceDelete(&Role{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("DELETE FROM \"users\" WHERE \"id\" = 1", toSql.ForceDelete(&User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("DELETE FROM \"roles\" WHERE \"id\" = 1", toSql.ForceDelete(&Role{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("DELETE FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2", toSql.ForceDelete(&GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("DELETE FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope'", toSql.ForceDelete(&GlobalScope{}))
}

func (s *ToSqlTestSuite) TestGet() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Get([]User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Get([]User{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Get([]GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT * FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Get([]GlobalScope{}))
}

func (s *ToSqlTestSuite) TestInvalidModel() {
	s.mockLog.EXPECT().Errorf("failed to get sql: %v", mock.MatchedBy(func(err error) bool {
		return s.EqualError(err, "unsupported data type: invalid: Table not set, please set it like: db.Model(&user) or db.Table(\"users\")")
	})).Once()

	toSql := gorm.NewToSql(s.query.Model("invalid").Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Empty(toSql.Get([]User{}))
}

func (s *ToSqlTestSuite) TestPluck() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT \"id\" FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Pluck("id", User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT \"id\" FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Pluck("id", User{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT \"id\" FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Pluck("id", GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT \"id\" FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Pluck("id", GlobalScope{}))
}

func (s *ToSqlTestSuite) TestSave() {
	toSql := gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES ($1,$2,$3,$4,$5,$6) RETURNING \"id\"", toSql.Save(&User{}))

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"created_at\"=$1,\"updated_at\"=$2,\"deleted_at\"=$3,\"name\"=$4,\"bio\"=$5,\"avatar\"=$6 WHERE \"users\".\"deleted_at\" IS NULL AND \"id\" = $7", toSql.Save(&User{Model: Model{ID: 2}}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql := toSql.Save(&User{Name: "to_sql_save"})
	s.Contains(sql, "INSERT INTO \"users\" (\"created_at\",\"updated_at\",\"deleted_at\",\"name\",\"bio\",\"avatar\") VALUES (")
	s.Contains(sql, ",NULL,'to_sql_save',NULL,'')")

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, true)
	sql = toSql.Save(&User{Model: Model{ID: 2}, Name: "to_sql_save"})
	s.Contains(sql, "UPDATE \"users\" SET \"created_at\"=")
	s.Contains(sql, ",\"updated_at\"=")
	s.Contains(sql, ",\"deleted_at\"=")
	s.Contains(sql, ",\"name\"='to_sql_save'")
	s.Contains(sql, ",\"bio\"=")
	s.Contains(sql, ",\"avatar\"=")
	s.Contains(sql, "WHERE \"users\".\"deleted_at\" IS NULL AND \"id\" = 2")

	// global scopes
	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("INSERT INTO \"global_scopes\" (\"created_at\",\"updated_at\",\"name\",\"deleted_at\") VALUES ($1,$2,$3,$4) RETURNING \"id\"", toSql.Save(&GlobalScope{}))

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"global_scopes\" SET \"created_at\"=$1,\"updated_at\"=$2,\"name\"=$3,\"deleted_at\"=$4 WHERE \"name\" = $5 AND \"global_scopes\".\"deleted_at\" IS NULL AND \"id\" = $6", toSql.Save(&GlobalScope{Model: Model{ID: 2}}))

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, true)
	sql = toSql.Save(&GlobalScope{Name: "to_sql_save"})
	s.Contains(sql, "INSERT INTO \"global_scopes\" (\"created_at\",\"updated_at\",\"name\",\"deleted_at\") VALUES (")
	s.Contains(sql, ",'to_sql_save',NULL) RETURNING \"id\"")

	toSql = gorm.NewToSql(s.query.(*gorm.Query), s.mockLog, true)
	sql = toSql.Save(&GlobalScope{Model: Model{ID: 2}, Name: "to_sql_save"})
	s.Contains(sql, "UPDATE \"global_scopes\" SET \"created_at\"=NULL,\"updated_at\"='")
	s.Contains(sql, "',\"name\"='to_sql_save',\"deleted_at\"=NULL WHERE \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL AND \"id\" = 2")
}

func (s *ToSqlTestSuite) TestSum() {
	toSql := gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT SUM(id) FROM \"users\" WHERE \"id\" = $1 AND \"users\".\"deleted_at\" IS NULL", toSql.Sum("id", User{}))

	toSql = gorm.NewToSql(s.query.Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT SUM(id) FROM \"users\" WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL", toSql.Sum("id", User{}))

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("SELECT SUM(id) FROM \"global_scopes\" WHERE \"id\" = $1 AND \"name\" = $2 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Sum("id", GlobalScope{}))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Equal("SELECT SUM(id) FROM \"global_scopes\" WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Sum("id", GlobalScope{}))
}

func (s *ToSqlTestSuite) TestUpdate() {
	toSql := gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"name\"=$1,\"updated_at\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update("name", "goravel"))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql := toSql.Update("name", "goravel")
	s.Contains(sql, "UPDATE \"users\" SET \"name\"='goravel',\"updated_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"name\"=$1,\"updated_at\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update(map[string]any{
		"name": "goravel",
	}))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql = toSql.Update(map[string]any{
		"name": "goravel",
	})
	s.Contains(sql, "UPDATE \"users\" SET \"name\"='goravel',\"updated_at\"=")
	s.Contains(sql, "WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"users\" SET \"updated_at\"=$1,\"name\"=$2 WHERE \"id\" = $3 AND \"users\".\"deleted_at\" IS NULL", toSql.Update(User{
		Name: "goravel",
	}))

	toSql = gorm.NewToSql(s.query.Model(&User{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql = toSql.Update(User{
		Name: "goravel",
	})
	s.Contains(sql, "UPDATE \"users\" SET \"updated_at\"=")
	s.Contains(sql, ",\"name\"='goravel' WHERE \"id\" = 1 AND \"users\".\"deleted_at\" IS NULL")

	// global scopes
	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, false)
	s.Equal("UPDATE \"global_scopes\" SET \"name\"=$1,\"updated_at\"=$2 WHERE \"id\" = $3 AND \"name\" = $4 AND \"global_scopes\".\"deleted_at\" IS NULL", toSql.Update("name", "goravel"))

	toSql = gorm.NewToSql(s.query.Model(&GlobalScope{}).Where("id", 1).(*gorm.Query), s.mockLog, true)
	sql = toSql.Update("name", "goravel")
	s.Contains(sql, "UPDATE \"global_scopes\" SET \"name\"='goravel',\"updated_at\"='")
	s.Contains(sql, "' WHERE \"id\" = 1 AND \"name\" = 'global_scope' AND \"global_scopes\".\"deleted_at\" IS NULL")
}
