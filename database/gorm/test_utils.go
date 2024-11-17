package gorm

import (
	"context"
	"fmt"
	"slices"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
	mocksconfig "github.com/goravel/framework/mocks/config"
	supportdocker "github.com/goravel/framework/support/docker"
)

type TestTable int

const (
	TestTableAddresses TestTable = iota
	TestTableAuthors
	TestTableBooks
	TestTableHouses
	TestTablePeoples
	TestTablePhones
	TestTableProducts
	TestTableReviews
	TestTableRoles
	TestTableRoleUser
	TestTableUsers
	TestTableGoravelUser
)

var testContext = context.Background()

type testMockDriver interface {
	Common()
	ReadWrite(readDatabaseConfig testing.DatabaseConfig)
	WithPrefixAndSingular()
}

type TestQueries struct {
}

func NewTestQueries() *TestQueries {
	return &TestQueries{}
}

func (r *TestQueries) Queries() map[contractsdatabase.Driver]*TestQuery {
	return r.queries(false)
}

func (r *TestQueries) QueriesOfReadWrite() map[contractsdatabase.Driver]map[string]*TestQuery {
	postgresDockers := supportdocker.Postgreses(2)
	sqliteDockers := supportdocker.Sqlites(2)
	if err := supportdocker.Ready(postgresDockers...); err != nil {
		panic(err)
	}

	readPostgresQuery := NewTestQuery(postgresDockers[0])
	writePostgresQuery := NewTestQuery(postgresDockers[1])

	readSqliteQuery := NewTestQuery(sqliteDockers[0])
	writeSqliteQuery := NewTestQuery(sqliteDockers[1])

	queries := map[contractsdatabase.Driver]map[string]*TestQuery{
		contractsdatabase.DriverPostgres: {
			"read":  readPostgresQuery,
			"write": writePostgresQuery,
		},
		contractsdatabase.DriverSqlite: {
			"read":  readSqliteQuery,
			"write": writeSqliteQuery,
		},
	}

	if supportdocker.TestModel == supportdocker.TestModelMinimum {
		return queries
	}

	// Create all containers first, containers will be returned directly, then check containers status, the speed will be faster.
	mysqlDockers := supportdocker.Mysqls(2)
	sqlserverDockers := supportdocker.Sqlservers(2)
	if err := supportdocker.Ready(mysqlDockers...); err != nil {
		panic(err)
	}
	if err := supportdocker.Ready(sqlserverDockers...); err != nil {
		panic(err)
	}

	readMysqlQuery := NewTestQuery(mysqlDockers[0])
	writeMysqlQuery := NewTestQuery(mysqlDockers[1])

	readSqlserverQuery := NewTestQuery(sqlserverDockers[0])
	writeSqlserverQuery := NewTestQuery(sqlserverDockers[1])

	queries[contractsdatabase.DriverMysql] = map[string]*TestQuery{
		"read":  readMysqlQuery,
		"write": writeMysqlQuery,
	}
	queries[contractsdatabase.DriverSqlserver] = map[string]*TestQuery{
		"read":  readSqlserverQuery,
		"write": writeSqlserverQuery,
	}

	return queries
}

func (r *TestQueries) QueriesWithPrefixAndSingular() map[contractsdatabase.Driver]*TestQuery {
	return r.queries(true)
}

func (r *TestQueries) QueryOfAdditional() *TestQuery {
	postgresDocker := supportdocker.Postgres()
	if err := supportdocker.Ready(postgresDocker); err != nil {
		panic(err)
	}
	postgresQuery := NewTestQuery(postgresDocker)

	return postgresQuery
}

func (r *TestQueries) queries(withPrefixAndSingular bool) map[contractsdatabase.Driver]*TestQuery {
	driverToTestQuery := make(map[contractsdatabase.Driver]*TestQuery)
	postgresDocker := supportdocker.Postgres()
	if err := supportdocker.Ready(postgresDocker); err != nil {
		panic(err)
	}

	driverToDocker := map[contractsdatabase.Driver]testing.DatabaseDriver{
		contractsdatabase.DriverPostgres: postgresDocker,
		contractsdatabase.DriverSqlite:   supportdocker.Sqlite(),
	}

	if supportdocker.TestModel != supportdocker.TestModelMinimum {
		mysqlDocker := supportdocker.Mysql()
		sqlserverDocker := supportdocker.Sqlserver()
		if err := supportdocker.Ready(mysqlDocker); err != nil {
			panic(err)
		}
		if err := supportdocker.Ready(sqlserverDocker); err != nil {
			panic(err)
		}

		driverToDocker[contractsdatabase.DriverMysql] = mysqlDocker
		driverToDocker[contractsdatabase.DriverSqlserver] = sqlserverDocker
	}

	for driver, docker := range driverToDocker {
		query := NewTestQuery(docker, withPrefixAndSingular)
		driverToTestQuery[driver] = query
	}

	return driverToTestQuery
}

type TestQuery struct {
	docker     testing.DatabaseDriver
	mockConfig *mocksconfig.Config
	mockDriver testMockDriver
	query      orm.Query
}

func NewTestQuery(docker testing.DatabaseDriver, withPrefixAndSingular ...bool) *TestQuery {
	mockConfig := &mocksconfig.Config{}
	mockDriver := getMockDriver(docker, mockConfig, docker.Driver().String())

	testQuery := &TestQuery{
		docker:     docker,
		mockConfig: mockConfig,
		mockDriver: mockDriver,
	}

	var (
		query *Query
		err   error
	)
	if len(withPrefixAndSingular) > 0 && withPrefixAndSingular[0] {
		mockDriver.WithPrefixAndSingular()
		query, err = BuildQuery(testContext, mockConfig, docker.Driver().String(), nil, nil)
	} else {
		mockDriver.Common()
		query, err = BuildQuery(testContext, mockConfig, docker.Driver().String(), nil, nil)
	}

	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", docker.Driver().String(), err))
	}

	testQuery.query = query

	return testQuery
}

func (r *TestQuery) CreateTable(testTables ...TestTable) {
	for table, sql := range newTestTables(r.docker.Driver()).All() {
		if len(testTables) == 0 || slices.Contains(testTables, table) {
			if _, err := r.query.Exec(sql()); err != nil {
				panic(fmt.Sprintf("create table %v failed: %v", table, err))
			}
		}
	}
}

func (r *TestQuery) Docker() testing.DatabaseDriver {
	return r.docker
}

func (r *TestQuery) MockConfig() *mocksconfig.Config {
	return r.mockConfig
}

func (r *TestQuery) Query() orm.Query {
	return r.query
}

func (r *TestQuery) QueryOfReadWrite(readDatabaseConfig testing.DatabaseConfig) (orm.Query, error) {
	mockConfig := &mocksconfig.Config{}
	mockDriver := getMockDriver(r.Docker(), mockConfig, r.Docker().Driver().String())
	mockDriver.ReadWrite(readDatabaseConfig)

	return BuildQuery(testContext, mockConfig, r.docker.Driver().String(), nil, nil)
}

func getMockDriver(docker testing.DatabaseDriver, mockConfig *mocksconfig.Config, connection string) testMockDriver {
	config := docker.Config()

	switch docker.Driver() {
	case contractsdatabase.DriverMysql:
		return NewMockMysql(mockConfig, connection, config.Database, config.Username, config.Password, config.Port)
	case contractsdatabase.DriverPostgres:
		return NewMockPostgres(mockConfig, connection, config.Database, config.Username, config.Password, config.Port)
	case contractsdatabase.DriverSqlite:
		return NewMockSqlite(mockConfig, connection, config.Database)
	case contractsdatabase.DriverSqlserver:
		return NewMockSqlserver(mockConfig, connection, config.Database, config.Username, config.Password, config.Port)
	default:
		panic("unsupported driver")
	}
}

type MockMysql struct {
	driver     contractsdatabase.Driver
	mockConfig *mocksconfig.Config

	connection string
	database   string
	password   string
	user       string
	port       int
}

func NewMockMysql(mockConfig *mocksconfig.Config, connection, database, username, password string, port int) *MockMysql {
	return &MockMysql{
		driver:     contractsdatabase.DriverMysql,
		mockConfig: mockConfig,
		connection: connection,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockMysql) Common() {
	r.mockConfig.On("GetString", "database.default").Return("mysql")
	r.mockConfig.On("GetString", "database.migrations.table").Return("migrations")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.single()
	r.basic()
}

func (r *MockMysql) ReadWrite(readDatabaseConfig testing.DatabaseConfig) {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: readDatabaseConfig.Database, Port: readDatabaseConfig.Port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: r.database, Port: r.port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.basic()
}

func (r *MockMysql) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("goravel_")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(true)
	r.single()
	r.basic()
}

func (r *MockMysql) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", r.connection)).Return(r.driver.String())
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.charset", r.connection)).Return("utf8mb4")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.loc", r.connection)).Return("Local")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", r.connection)).Return(r.database)

	mockPool(r.mockConfig)
}

func (r *MockMysql) single() {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return(nil)
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return(nil)
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.host", r.connection)).Return("127.0.0.1")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.username", r.connection)).Return(r.user)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.password", r.connection)).Return(r.password)
	r.mockConfig.On("GetInt", fmt.Sprintf("database.connections.%s.port", r.connection)).Return(r.port)
}

type MockPostgres struct {
	driver     contractsdatabase.Driver
	mockConfig *mocksconfig.Config

	connection string
	database   string
	password   string
	user       string
	port       int
}

func NewMockPostgres(mockConfig *mocksconfig.Config, connection, database, username, password string, port int) *MockPostgres {
	return &MockPostgres{
		driver:     contractsdatabase.DriverPostgres,
		mockConfig: mockConfig,
		connection: connection,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockPostgres) Common() {
	r.mockConfig.On("GetString", "database.default").Return("postgres")
	r.mockConfig.On("GetString", "database.migrations.table").Return("migrations")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.single()
	r.basic()
}

func (r *MockPostgres) ReadWrite(readDatabaseConfig testing.DatabaseConfig) {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: readDatabaseConfig.Database, Port: readDatabaseConfig.Port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: r.database, Port: r.port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.basic()
}

func (r *MockPostgres) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("goravel_")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(true)
	r.single()
	r.basic()
}

func (r *MockPostgres) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", r.connection)).Return(r.driver.String())
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.sslmode", r.connection)).Return("disable")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.timezone", r.connection)).Return("UTC")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", r.connection)).Return(r.database)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.search_path", r.connection), "public").Return("public")

	mockPool(r.mockConfig)
}

func (r *MockPostgres) single() {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return(nil)
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return(nil)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.host", r.connection)).Return("127.0.0.1")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.username", r.connection)).Return(r.user)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.password", r.connection)).Return(r.password)
	r.mockConfig.On("GetInt", fmt.Sprintf("database.connections.%s.port", r.connection)).Return(r.port)
}

type MockSqlite struct {
	driver     contractsdatabase.Driver
	mockConfig *mocksconfig.Config

	connection string
	database   string
}

func NewMockSqlite(mockConfig *mocksconfig.Config, connection, database string) *MockSqlite {
	return &MockSqlite{
		driver:     contractsdatabase.DriverSqlite,
		mockConfig: mockConfig,
		connection: connection,
		database:   database,
	}
}

func (r *MockSqlite) Common() {
	r.mockConfig.On("GetString", "database.default").Return("sqlite")
	r.mockConfig.On("GetString", "database.migrations.table").Return("migrations")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.single()
	r.basic()
}

func (r *MockSqlite) ReadWrite(readDatabaseConfig testing.DatabaseConfig) {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return([]contractsdatabase.Config{
		{Database: readDatabaseConfig.Database},
	})
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return([]contractsdatabase.Config{
		{Database: r.database},
	})
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.basic()
}

func (r *MockSqlite) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("goravel_")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(true)
	r.single()
	r.basic()
}

func (r *MockSqlite) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", r.connection)).Return(r.driver.String())
	mockPool(r.mockConfig)
}

func (r *MockSqlite) single() {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return(nil)
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return(nil)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", r.connection)).Return(r.database)
}

type MockSqlserver struct {
	driver     contractsdatabase.Driver
	mockConfig *mocksconfig.Config

	connection string
	database   string
	password   string
	user       string
	port       int
}

func NewMockSqlserver(mockConfig *mocksconfig.Config, connection, database, username, password string, port int) *MockSqlserver {
	return &MockSqlserver{
		driver:     contractsdatabase.DriverSqlserver,
		mockConfig: mockConfig,
		connection: connection,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockSqlserver) Common() {
	r.mockConfig.On("GetString", "database.default").Return("sqlserver")
	r.mockConfig.On("GetString", "database.migrations.table").Return("migrations")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.single()
	r.basic()
}

func (r *MockSqlserver) ReadWrite(readDatabaseConfig testing.DatabaseConfig) {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: readDatabaseConfig.Database, Port: readDatabaseConfig.Port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return([]contractsdatabase.Config{
		{Host: "127.0.0.1", Database: r.database, Port: r.port, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(false)
	r.basic()
}

func (r *MockSqlserver) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.prefix", r.connection)).Return("goravel_")
	r.mockConfig.On("GetBool", fmt.Sprintf("database.connections.%s.singular", r.connection)).Return(true)
	r.single()
	r.basic()
}

func (r *MockSqlserver) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.driver", r.connection)).Return(r.driver.String())
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.database", r.connection)).Return(r.database)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.charset", r.connection)).Return("utf8mb4")
	mockPool(r.mockConfig)
}

func (r *MockSqlserver) single() {
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.read", r.connection)).Return(nil)
	r.mockConfig.On("Get", fmt.Sprintf("database.connections.%s.write", r.connection)).Return(nil)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.host", r.connection)).Return("127.0.0.1")
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.username", r.connection)).Return(r.user)
	r.mockConfig.On("GetString", fmt.Sprintf("database.connections.%s.password", r.connection)).Return(r.password)
	r.mockConfig.On("GetInt", fmt.Sprintf("database.connections.%s.port", r.connection)).Return(r.port)
}

type testTables struct {
	driver contractsdatabase.Driver
}

func newTestTables(driver contractsdatabase.Driver) *testTables {
	return &testTables{driver: driver}
}

func (r *testTables) All() map[TestTable]func() string {
	return map[TestTable]func() string{
		TestTableAddresses:   r.addresses,
		TestTableAuthors:     r.authors,
		TestTableBooks:       r.books,
		TestTableHouses:      r.houses,
		TestTablePeoples:     r.peoples,
		TestTablePhones:      r.phones,
		TestTableProducts:    r.products,
		TestTableReviews:     r.reviews,
		TestTableRoles:       r.roles,
		TestTableRoleUser:    r.roleUser,
		TestTableUsers:       r.users,
		TestTableGoravelUser: r.goravelUser,
	}
}

func (r *testTables) peoples() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE peoples (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  body varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE peoples (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE peoples (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE peoples (
  id bigint NOT NULL IDENTITY(1,1),
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) reviews() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE reviews (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  body varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE reviews (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE reviews (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE reviews (
  id bigint NOT NULL IDENTITY(1,1),
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) products() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE products (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE products (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE products (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE products (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) users() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) goravelUser() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE goravel_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE goravel_user (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE goravel_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE goravel_user (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) addresses() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE addresses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_addresses_created_at (created_at),
  KEY idx_addresses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE addresses (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE addresses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE addresses (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) books() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE books (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_books_created_at (created_at),
  KEY idx_books_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE books (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE books (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE books (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) authors() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE authors (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  book_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_books_created_at (created_at),
  KEY idx_books_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE authors (
  id SERIAL PRIMARY KEY NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE authors (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE authors (
  id bigint NOT NULL IDENTITY(1,1),
  book_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) roles() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE roles (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_roles_created_at (created_at),
  KEY idx_roles_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE roles (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE roles (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE roles (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) houses() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE houses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  houseable_id bigint(20) unsigned NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_houses_created_at (created_at),
  KEY idx_houses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE houses (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  houseable_id int NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE houses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  houseable_id int NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE houses (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  houseable_id bigint NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) phones() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE phones (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  phoneable_id bigint(20) unsigned NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_phones_created_at (created_at),
  KEY idx_phones_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE phones (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  phoneable_id int NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE phones (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  phoneable_id int NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE phones (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  phoneable_id bigint NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) roleUser() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE role_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint(20) unsigned NOT NULL,
  user_id bigint(20) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE role_user (
  id SERIAL PRIMARY KEY NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE role_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE role_user (
  id bigint NOT NULL IDENTITY(1,1),
  role_id bigint NOT NULL,
  user_id bigint NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func mockPool(mockConfig *mocksconfig.Config) {
	mockConfig.On("GetInt", "database.pool.max_idle_conns", 10).Return(10)
	mockConfig.On("GetInt", "database.pool.max_open_conns", 100).Return(100)
	mockConfig.On("GetInt", "database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.On("GetInt", "database.pool.conn_max_lifetime", 3600).Return(3600)
}
