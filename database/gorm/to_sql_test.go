package gorm

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/env"
)

type ToSqlTestSuite struct {
	suite.Suite
	query ormcontract.Query
}

func TestToSqlTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	if err := testDatabaseDocker.Fresh(); err != nil {
		t.Fatal(err)
	}

	mysqlDocker := NewMysqlDocker(testDatabaseDocker)
	query, err := mysqlDocker.New()
	if err != nil {
		log.Fatalf("Init mysql error: %s", err)
	}

	suite.Run(t, &ToSqlTestSuite{
		query: query,
	})
}

func (s *ToSqlTestSuite) SetupTest() {}

func (s *ToSqlTestSuite) TestCount() {
	toSql := NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT count(*) FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Count())

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT count(*) FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Count())
}

func (s *ToSqlTestSuite) TestCreate() {
	user := User{Name: "to_sql_create"}
	toSql := NewToSql(s.query.(*QueryImpl), false)
	s.Equal("INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`bio`,`avatar`) VALUES (?,?,?,?,?,?)", toSql.Create(&user))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	s.Contains(toSql.Create(&user), "INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`bio`,`avatar`) VALUES (")
	s.Contains(toSql.Create(&user), ",NULL,'to_sql_create',NULL,'')")

	var users []User
	s.NoError(s.query.Where("name", "to_sql_create").Get(&users))
	s.Len(users, 0)
}

func (s *ToSqlTestSuite) TestDelete() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("UPDATE `users` SET `deleted_at`=? WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Delete(User{}))

	toSql = NewToSql(s.query.(*QueryImpl), false)
	s.Equal("UPDATE `users` SET `deleted_at`=? WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL", toSql.Delete(User{}, 1))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("DELETE FROM `roles` WHERE `id` = ?", toSql.Delete(Role{}))

	toSql = NewToSql(s.query.(*QueryImpl), false)
	s.Equal("DELETE FROM `roles` WHERE `roles`.`id` = ?", toSql.Delete(Role{}, 1))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	sql := toSql.Delete(User{})
	s.Contains(sql, "UPDATE `users` SET `deleted_at`=")
	s.Contains(sql, "WHERE `id` = 1 AND `users`.`deleted_at` IS NULL")

	toSql = NewToSql(s.query.(*QueryImpl), true)
	sql = toSql.Delete(User{}, 1)
	s.Contains(sql, "UPDATE `users` SET `deleted_at`=")
	s.Contains(sql, "WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL")

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("DELETE FROM `roles` WHERE `id` = 1", toSql.Delete(Role{}))

	toSql = NewToSql(s.query.(*QueryImpl), true)
	s.Equal("DELETE FROM `roles` WHERE `roles`.`id` = 1", toSql.Delete(Role{}, 1))
}

func (s *ToSqlTestSuite) TestFind() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT * FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Find(User{}))

	toSql = NewToSql(s.query.(*QueryImpl), false)
	s.Equal("SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL", toSql.Find(User{}, 1))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT * FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Find(User{}))

	toSql = NewToSql(s.query.(*QueryImpl), true)
	s.Equal("SELECT * FROM `users` WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Find(User{}, 1))
}

func (s *ToSqlTestSuite) TestFirst() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT * FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT ?", toSql.First(User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT * FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1", toSql.First(User{}))
}

func (s *ToSqlTestSuite) TestGet() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT * FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Get([]User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT * FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Get([]User{}))
}

func (s *ToSqlTestSuite) TestPluck() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT `id` FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Pluck("id", User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT `id` FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Pluck("id", User{}))
}

func (s *ToSqlTestSuite) TestSave() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`bio`,`avatar`) VALUES (?,?,?,?,?,?)", toSql.Save(&User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	sql := toSql.Save(&User{})
	s.Contains(sql, "INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`bio`,`avatar`) VALUES (")
	s.Contains(sql, ",NULL,'',NULL,'')")
}

func (s *ToSqlTestSuite) TestSum() {
	toSql := NewToSql(s.query.Where("id", 1).(*QueryImpl), false)
	s.Equal("SELECT SUM(id) FROM `users` WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Sum("id", User{}))

	toSql = NewToSql(s.query.Where("id", 1).(*QueryImpl), true)
	s.Equal("SELECT SUM(id) FROM `users` WHERE `id` = 1 AND `users`.`deleted_at` IS NULL", toSql.Sum("id", User{}))
}

func (s *ToSqlTestSuite) TestUpdate() {
	toSql := NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), false)
	s.Equal("UPDATE `users` SET `name`=?,`updated_at`=? WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Update("name", "goravel"))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	sql := toSql.Update("name", "goravel")
	s.Contains(sql, "UPDATE `users` SET `name`='goravel',`updated_at`=")
	s.Contains(sql, "WHERE `id` = 1 AND `users`.`deleted_at` IS NULL")

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), false)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	s.Empty(toSql.Update(0, "goravel"))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), false)
	s.Equal("UPDATE `users` SET `name`=?,`updated_at`=? WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Update(map[string]any{
		"name": "goravel",
	}))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	sql = toSql.Update(map[string]any{
		"name": "goravel",
	})
	s.Contains(sql, "UPDATE `users` SET `name`='goravel',`updated_at`=")
	s.Contains(sql, "WHERE `id` = 1 AND `users`.`deleted_at` IS NULL")

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), false)
	s.Equal("UPDATE `users` SET `updated_at`=?,`name`=? WHERE `id` = ? AND `users`.`deleted_at` IS NULL", toSql.Update(User{
		Name: "goravel",
	}))

	toSql = NewToSql(s.query.Model(User{}).Where("id", 1).(*QueryImpl), true)
	sql = toSql.Update(User{
		Name: "goravel",
	})
	s.Contains(sql, "UPDATE `users` SET `updated_at`=")
	s.Contains(sql, ",`name`='goravel' WHERE `id` = 1 AND `users`.`deleted_at` IS NULL")
}
