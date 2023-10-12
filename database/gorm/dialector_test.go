package gorm

import (
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	databasecontract "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
)

type DialectorTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	config     databasecontract.Config
}

func TestDialectorTestSuite(t *testing.T) {
	suite.Run(t, &DialectorTestSuite{
		config: databasecontract.Config{
			Host:     "localhost",
			Port:     3306,
			Database: "forge",
			Username: "root",
			Password: "123123",
		},
	})
}

func (s *DialectorTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
}

func (s *DialectorTestSuite) TestMysql() {
	dialector := NewDialectorImpl(s.mockConfig, orm.DriverMysql.String())
	s.mockConfig.On("GetString", "database.connections.mysql.driver").
		Return(orm.DriverMysql.String()).Once()
	s.mockConfig.On("GetString", "database.connections.mysql.charset").
		Return("utf8mb4").Once()
	s.mockConfig.On("GetString", "database.connections.mysql.loc").
		Return("Local").Once()
	dialectors, err := dialector.Make([]databasecontract.Config{s.config})
	s.Equal(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s&multiStatements=true",
			s.config.Username, s.config.Password, s.config.Host, s.config.Port, s.config.Database, "utf8mb4", true, "Local"),
	}), dialectors[0])
	s.Nil(err)
}

func (s *DialectorTestSuite) TestPostgresql() {
	dialector := NewDialectorImpl(s.mockConfig, orm.DriverPostgresql.String())
	s.mockConfig.On("GetString", "database.connections.postgresql.driver").
		Return(orm.DriverPostgresql.String()).Once()
	s.mockConfig.On("GetString", "database.connections.postgresql.sslmode").
		Return("disable").Once()
	s.mockConfig.On("GetString", "database.connections.postgresql.timezone").
		Return("UTC").Once()
	dialectors, err := dialector.Make([]databasecontract.Config{s.config})
	s.Equal(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
			s.config.Username, s.config.Password, s.config.Host, s.config.Port, s.config.Database, "disable", "UTC"),
	}), dialectors[0])
	s.Nil(err)
}

func (s *DialectorTestSuite) TestSqlite() {
	dialector := NewDialectorImpl(s.mockConfig, orm.DriverSqlite.String())
	s.mockConfig.On("GetString", "database.connections.sqlite.driver").
		Return(orm.DriverSqlite.String()).Once()
	dialectors, err := dialector.Make([]databasecontract.Config{s.config})
	s.Equal(sqlite.Open(fmt.Sprintf("%s?multi_stmts=true", s.config.Database)), dialectors[0])
	s.Nil(err)
}

func (s *DialectorTestSuite) TestSqlserver() {
	dialector := NewDialectorImpl(s.mockConfig, orm.DriverSqlserver.String())
	s.mockConfig.On("GetString", "database.connections.sqlserver.driver").
		Return(orm.DriverSqlserver.String()).Once()
	s.mockConfig.On("GetString", "database.connections.sqlserver.charset").
		Return("utf8mb4").Once()
	dialectors, err := dialector.Make([]databasecontract.Config{s.config})
	s.Equal(sqlserver.New(sqlserver.Config{
		DSN: fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&charset=%s&MultipleActiveResultSets=true",
			s.config.Username, s.config.Password, s.config.Host, s.config.Port, s.config.Database, "utf8mb4"),
	}), dialectors[0])
	s.Nil(err)
}
