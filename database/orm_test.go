package database

import (
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/testing/mock"
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
}

var (
	testMysqlDB      ormcontract.DB
	testPostgresqlDB ormcontract.DB
	testSqliteDB     ormcontract.DB
	testSqlserverDB  ormcontract.DB
)

func TestOrmSuite(t *testing.T) {
	mysqlPool, mysqlDocker, mysqlDB, err := gorm.MysqlDocker()
	testMysqlDB = mysqlDB
	if err != nil {
		log.Fatalf("Get gorm mysql error: %s", err)
	}

	postgresqlPool, postgresqlDocker, postgresqlDB, err := gorm.PostgresqlDocker()
	testPostgresqlDB = postgresqlDB
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

	_, _, sqliteDB, err := gorm.SqliteDocker()
	testSqliteDB = sqliteDB
	if err != nil {
		log.Fatalf("Get gorm sqlite error: %s", err)
	}

	sqlserverPool, sqlserverDocker, sqlserverDB, err := gorm.SqlserverDocker()
	testSqlserverDB = sqlserverDB
	if err != nil {
		log.Fatalf("Get gorm postgresql error: %s", err)
	}

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
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(4)
	testOrm := newTestOrm()
	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()))
	}

	mockConfig.AssertExpectations(s.T())
}

func (s *OrmSuite) TestDB() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(5)

	testOrm := newTestOrm()
	db, err := testOrm.DB()
	s.NotNil(db)
	s.Nil(err)

	for _, connection := range connections {
		db, err := testOrm.Connection(connection.String()).DB()
		s.NotNil(db)
		s.Nil(err)
	}

	mockConfig.AssertExpectations(s.T())
}

func (s *OrmSuite) TestQuery() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(5)
	testOrm := newTestOrm()
	s.NotNil(testOrm.Query())

	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()).Query())
	}

	mockConfig.AssertExpectations(s.T())
}

func (s *OrmSuite) TestTransactionSuccess() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(12)

	testOrm := newTestOrm()
	for _, connection := range connections {
		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
		s.Nil(testOrm.Connection(connection.String()).Transaction(func(tx ormcontract.Transaction) error {
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))

			return nil
		}))

		var user2, user3 User
		s.Nil(testOrm.Connection(connection.String()).Query().Find(&user2, user.ID))
		s.Nil(testOrm.Connection(connection.String()).Query().Find(&user3, user1.ID))

	}

	mockConfig.AssertExpectations(s.T())
}

func (s *OrmSuite) TestTransactionError() {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "database.default").Return(ormcontract.DriverMysql.String()).Times(8)

	testOrm := newTestOrm()
	for _, connection := range connections {
		s.NotNil(testOrm.Connection(connection.String()).Transaction(func(tx ormcontract.Transaction) error {
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

	mockConfig.AssertExpectations(s.T())
}

func newTestOrm() *Orm {
	return &Orm{
		instances: map[string]ormcontract.DB{
			ormcontract.DriverMysql.String():      testMysqlDB,
			ormcontract.DriverPostgresql.String(): testPostgresqlDB,
			ormcontract.DriverSqlite.String():     testSqliteDB,
			ormcontract.DriverSqlserver.String():  testSqlserverDB,
		},
	}
}
