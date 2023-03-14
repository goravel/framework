package database

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
)

var connections = []contractsorm.Driver{
	contractsorm.DriverMysql,
	contractsorm.DriverPostgresql,
	contractsorm.DriverSqlite,
	contractsorm.DriverSqlserver,
}

type User struct {
	orm.Model
	orm.SoftDeletes
	Name   string
	Avatar string
}

type OrmSuite struct {
	suite.Suite
}

var (
	testMysqlDB      contractsorm.Query
	testPostgresqlDB contractsorm.Query
	testSqliteDB     contractsorm.Query
	testSqlserverDB  contractsorm.Query
)

func TestOrmSuite(t *testing.T) {
	mysqlPool, mysqlDocker, mysqlDB, err := gorm.MysqlDocker()
	if err != nil {
		log.Fatalf("Get mysql error: %s", err)
	}
	testMysqlDB = mysqlDB

	postgresqlPool, postgresqlDocker, postgresqlDB, err := gorm.PostgresqlDocker()
	if err != nil {
		log.Fatalf("Get postgresql error: %s", err)
	}
	testPostgresqlDB = postgresqlDB

	_, _, sqliteDB, err := gorm.SqliteDocker("goravel")
	if err != nil {
		log.Fatalf("Get sqlite error: %s", err)
	}
	testSqliteDB = sqliteDB

	sqlserverPool, sqlserverDocker, sqlserverDB, err := gorm.SqlserverDocker()
	if err != nil {
		log.Fatalf("Get sqlserver error: %s", err)
	}
	testSqlserverDB = sqlserverDB

	suite.Run(t, new(OrmSuite))

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

func (s *OrmSuite) TestConnection() {
	testOrm := newTestOrm()
	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()))
	}
}

func (s *OrmSuite) TestDB() {
	testOrm := newTestOrm()
	db, err := testOrm.DB()
	s.NotNil(db)
	s.Nil(err)

	for _, connection := range connections {
		db, err := testOrm.Connection(connection.String()).DB()
		s.NotNil(db)
		s.Nil(err)
	}
}

func (s *OrmSuite) TestQuery() {
	testOrm := newTestOrm()
	s.NotNil(testOrm.Query())

	s.NotPanics(func() {
		for i := 0; i < 5; i++ {
			go func() {
				var user User
				_ = testOrm.Query().Find(&user, 1)
			}()
		}
	})

	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()).Query())
	}
}

func (s *OrmSuite) TestTransactionSuccess() {
	testOrm := newTestOrm()
	for _, connection := range connections {
		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
		s.Nil(testOrm.Connection(connection.String()).Transaction(func(tx contractsorm.Transaction) error {
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))

			return nil
		}))

		var user2, user3 User
		s.Nil(testOrm.Connection(connection.String()).Query().Find(&user2, user.ID))
		s.Nil(testOrm.Connection(connection.String()).Query().Find(&user3, user1.ID))
	}
}

func (s *OrmSuite) TestTransactionError() {
	testOrm := newTestOrm()
	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()).Transaction(func(tx contractsorm.Transaction) error {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&user))

			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&user1))

			return errors.New("error")
		}))

		var users []User
		s.Nil(testOrm.Connection(connection.String()).Query().Find(&users))
		s.Equal(0, len(users))
	}
}

func newTestOrm() *Orm {
	return &Orm{
		ctx:      context.Background(),
		instance: testMysqlDB,
		instances: map[string]contractsorm.Query{
			contractsorm.DriverMysql.String():      testMysqlDB,
			contractsorm.DriverPostgresql.String(): testPostgresqlDB,
			contractsorm.DriverSqlite.String():     testSqliteDB,
			contractsorm.DriverSqlserver.String():  testSqlserverDB,
		},
	}
}
