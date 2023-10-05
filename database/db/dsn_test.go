package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	databasecontract "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
)

const (
	testHost     = "127.0.0.1"
	testPort     = 3306
	testDatabase = "forge"
	testUsername = "root"
	testPassword = "123123"
)

var testConfig = databasecontract.Config{
	Host:     testHost,
	Port:     testPort,
	Database: testDatabase,
	Username: testUsername,
	Password: testPassword,
}

type DsnTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
}

func TestDsnTestSuite(t *testing.T) {
	suite.Run(t, new(DsnTestSuite))
}

func (s *DsnTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
}

func (s *DsnTestSuite) TestMysql() {
	connection := orm.DriverMysql.String()
	dsn := NewDsnImpl(s.mockConfig, connection)
	charset := "utf8mb4"
	loc := "Local"
	s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.charset", connection)).Return(charset).Once()
	s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.loc", connection)).Return(loc).Once()

	s.Equal(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
		testUsername, testPassword, testHost, testPort, testDatabase, charset, true, loc), dsn.Mysql(testConfig))
}

func (s *DsnTestSuite) TestPostgresql() {
	connection := orm.DriverPostgresql.String()
	dsn := NewDsnImpl(s.mockConfig, connection)
	sslmode := "disable"
	timezone := "UTC"
	s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.sslmode", connection)).Return(sslmode).Once()
	s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.timezone", connection)).Return(timezone).Once()

	s.Equal(fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		testUsername, testPassword, testHost, testPort, testDatabase, sslmode, timezone), dsn.Postgresql(testConfig))
}

func (s *DsnTestSuite) TestSqlite() {
	dsn := NewDsnImpl(s.mockConfig, "")
	s.Equal(fmt.Sprintf("%s?multi_stmts=true", testDatabase), dsn.Sqlite(testConfig))
}

func (s *DsnTestSuite) TestSqlserver() {
	connection := orm.DriverSqlserver.String()
	dsn := NewDsnImpl(s.mockConfig, connection)
	charset := "utf8mb4"
	s.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.charset", connection)).Return(charset).Once()

	s.Equal(fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
		testUsername, testPassword, testHost, testPort, testDatabase, charset), dsn.Sqlserver(testConfig))
}
