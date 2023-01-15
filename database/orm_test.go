package database

import (
	"errors"
	"log"
	"testing"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/suite"
)

var connections = []ormcontract.Driver{
	ormcontract.DriverMysql,
	ormcontract.DriverPostgresql,
	ormcontract.DriverSqlite,
	ormcontract.DriverSqlserver,
}

type User struct {
	orm.Model
	orm.SoftDeletes
	Name   string
	Avatar string
}

type OrmSuite struct {
	suite.Suite
	orm ormcontract.Orm
}

func TestOrmSuite(t *testing.T) {
	mysqlPool, mysqlDocker, mysqlDB, err := gorm.MysqlDocker()
	if err != nil {
		log.Fatalf("Get gorm mysql error: %s", err)
	}

	postgresqlPool, postgresqlDocker, postgresqlDB, err := gorm.PostgresqlDocker()
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

	_, _, sqliteDB, err := gorm.SqliteDocker()
	if err != nil {
		log.Fatalf("Get gorm sqlite error: %s", err)
	}

	sqlserverPool, sqlserverDocker, sqlserverDB, err := gorm.SqlserverDocker()
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

	suite.Run(t, &OrmSuite{
		orm: &Orm{
			instances: map[string]ormcontract.DB{
				ormcontract.DriverMysql.String():      mysqlDB,
				ormcontract.DriverPostgresql.String(): postgresqlDB,
				ormcontract.DriverSqlite.String():     sqliteDB,
				ormcontract.DriverSqlserver.String():  sqlserverDB,
			},
		},
	})

	file.Remove("goravel")

	if err := mysqlPool.Purge(mysqlDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := postgresqlPool.Purge(postgresqlDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	if err := sqlserverPool.Purge(sqlserverDocker); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func (s *OrmSuite) SetupTest() {

}

func (s *OrmSuite) TestTransactionSuccess() {
	for _, connection := range connections {
		mockConfig := mock.Config()
		mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(3)
		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
		s.Nil(s.orm.Connection(connection.String()).Transaction(func(tx ormcontract.Transaction) error {
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))

			return nil
		}))

		var user2, user3 User
		s.Nil(s.orm.Connection(connection.String()).Query().Find(&user2, user.ID))
		s.Nil(s.orm.Connection(connection.String()).Query().Find(&user3, user1.ID))
		mockConfig.AssertExpectations(s.T())
	}
}

func (s *OrmSuite) TestTransactionError() {
	for _, connection := range connections {
		mockConfig := mock.Config()
		mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Twice()
		s.NotNil(s.orm.Connection(connection.String()).Transaction(func(tx ormcontract.Transaction) error {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&user))

			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&user1))

			return errors.New("error")
		}))

		var users []User
		s.Nil(s.orm.Connection(connection.String()).Query().Find(&users))
		s.Equal(0, len(users))
		mockConfig.AssertExpectations(s.T())
	}
}
