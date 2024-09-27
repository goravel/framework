package gorm

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
	mocksconfig "github.com/goravel/framework/mocks/config"
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

var testContext context.Context

type TestReadWriteConfig struct {
	ReadPort  int
	WritePort int

	// Used by Sqlite
	ReadDatabase string
}

type testMockDriver interface {
	Common()
	ReadWrite(config TestReadWriteConfig)
	WithPrefixAndSingular()
}

type TestQuery struct {
	docker     testing.DatabaseDriver
	mockConfig *mocksconfig.Config
	mockDriver testMockDriver
	query      orm.Query
}

func NewTestQuery(docker testing.DatabaseDriver, withPrefixAndSingular ...bool) (*TestQuery, error) {
	config := docker.Config()
	mockConfig := &mocksconfig.Config{}

	var mockDriver testMockDriver
	switch docker.Driver() {
	case orm.DriverMysql:
		mockDriver = NewMockMysql(mockConfig, config.Database, config.Username, config.Password, config.Port)
	case orm.DriverPostgres:
		mockDriver = NewMockPostgres(mockConfig, config.Database, config.Username, config.Password, config.Port)
	case orm.DriverSqlite:
		mockDriver = NewMockSqlite(mockConfig, config.Database)
	case orm.DriverSqlserver:
		mockDriver = NewMockSqlserver(mockConfig, config.Database, config.Username, config.Password, config.Port)
	default:
		return nil, fmt.Errorf("unsupported driver %s", docker.Driver())
	}

	testQuery := &TestQuery{
		docker:     docker,
		mockConfig: mockConfig,
		mockDriver: mockDriver,
	}

	var (
		query *QueryImpl
		err   error
	)
	if len(withPrefixAndSingular) > 0 && withPrefixAndSingular[0] {
		mockDriver.WithPrefixAndSingular()
		query, err = InitializeQuery(testContext, mockConfig, docker.Driver().String())
	} else {
		mockDriver.Common()
		query, err = InitializeQuery(testContext, mockConfig, docker.Driver().String())
	}

	if err != nil {
		return nil, fmt.Errorf("connect to %s failed", docker.Driver().String())
	}

	testQuery.query = query

	return testQuery, nil
}

func (r *TestQuery) CreateTable(testTables ...TestTable) error {
	for table, sql := range newTestTables(r.docker.Driver()).All() {
		if len(testTables) == 0 || slices.Contains(testTables, table) {
			if _, err := r.query.Exec(sql()); err != nil {
				return err
			}
		}
	}

	return nil
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

func (r *TestQuery) QueryOfReadWrite(config TestReadWriteConfig) (orm.Query, error) {
	r.mockDriver.ReadWrite(config)

	return InitializeQuery(testContext, r.mockConfig, r.docker.Driver().String())
}

type MockMysql struct {
	driver     orm.Driver
	mockConfig *mocksconfig.Config

	database string
	password string
	user     string
	port     int
}

func NewMockMysql(mockConfig *mocksconfig.Config, database, username, password string, port int) *MockMysql {
	return &MockMysql{
		driver:     orm.DriverMysql,
		mockConfig: mockConfig,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockMysql) Common() {
	r.mockConfig.On("GetString", "database.default").Return("mysql")
	r.mockConfig.On("GetString", "database.migrations").Return("migrations")
	r.mockConfig.On("GetString", "database.connections.mysql.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.mysql.singular").Return(false)
	r.single()
	r.basic()
}

func (r *MockMysql) ReadWrite(config TestReadWriteConfig) {
	r.mockConfig = &mocksconfig.Config{}
	r.mockConfig.On("Get", "database.connections.mysql.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.ReadPort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", "database.connections.mysql.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.WritePort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", "database.connections.mysql.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.mysql.singular").Return(false)
	r.basic()
}

func (r *MockMysql) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", "database.connections.mysql.prefix").Return("goravel_")
	r.mockConfig.On("GetBool", "database.connections.mysql.singular").Return(true)
	r.single()
	r.basic()
}

func (r *MockMysql) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", "database.connections.mysql.driver").Return(r.driver.String())
	r.mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4")
	r.mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local")
	r.mockConfig.On("GetString", "database.connections.mysql.database").Return(r.database)

	mockPool(r.mockConfig)
}

func (r *MockMysql) single() {
	r.mockConfig.On("Get", "database.connections.mysql.read").Return(nil)
	r.mockConfig.On("Get", "database.connections.mysql.write").Return(nil)
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", "database.connections.mysql.host").Return("127.0.0.1")
	r.mockConfig.On("GetString", "database.connections.mysql.username").Return(r.user)
	r.mockConfig.On("GetString", "database.connections.mysql.password").Return(r.password)
	r.mockConfig.On("GetInt", "database.connections.mysql.port").Return(r.port)
}

type MockPostgres struct {
	driver     orm.Driver
	mockConfig *mocksconfig.Config

	database string
	password string
	user     string
	port     int
}

func NewMockPostgres(mockConfig *mocksconfig.Config, database, username, password string, port int) *MockPostgres {
	return &MockPostgres{
		driver:     orm.DriverPostgres,
		mockConfig: mockConfig,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockPostgres) Common() {
	r.mockConfig.On("GetString", "database.default").Return("postgres")
	r.mockConfig.On("GetString", "database.migrations").Return("migrations")
	r.mockConfig.On("GetString", "database.connections.postgres.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.postgres.singular").Return(false)
	r.single()
	r.basic()
}

func (r *MockPostgres) ReadWrite(config TestReadWriteConfig) {
	r.mockConfig = &mocksconfig.Config{}
	r.mockConfig.On("Get", "database.connections.postgres.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.ReadPort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", "database.connections.postgres.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.WritePort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", "database.connections.postgres.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.postgres.singular").Return(false)
	r.basic()
}

func (r *MockPostgres) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", "database.connections.postgres.prefix").Return("goravel_")
	r.mockConfig.On("GetBool", "database.connections.postgres.singular").Return(true)
	r.single()
	r.basic()
}

func (r *MockPostgres) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", "database.connections.postgres.driver").Return(orm.DriverPostgres.String())
	r.mockConfig.On("GetString", "database.connections.postgres.sslmode").Return("disable")
	r.mockConfig.On("GetString", "database.connections.postgres.timezone").Return("UTC")
	r.mockConfig.On("GetString", "database.connections.postgres.database").Return(r.database)

	mockPool(r.mockConfig)
}

func (r *MockPostgres) single() {
	r.mockConfig.On("Get", "database.connections.postgres.read").Return(nil)
	r.mockConfig.On("Get", "database.connections.postgres.write").Return(nil)
	r.mockConfig.On("GetString", "database.connections.postgres.host").Return("127.0.0.1")
	r.mockConfig.On("GetString", "database.connections.postgres.username").Return(r.user)
	r.mockConfig.On("GetString", "database.connections.postgres.password").Return(r.password)
	r.mockConfig.On("GetInt", "database.connections.postgres.port").Return(r.port)
}

type MockSqlite struct {
	driver     orm.Driver
	mockConfig *mocksconfig.Config

	database string
}

func NewMockSqlite(mockConfig *mocksconfig.Config, database string) *MockSqlite {
	return &MockSqlite{
		driver:     orm.DriverSqlite,
		mockConfig: mockConfig,
		database:   database,
	}
}

func (r *MockSqlite) Common() {
	r.mockConfig.On("GetString", "database.default").Return("sqlite")
	r.mockConfig.On("GetString", "database.migrations").Return("migrations")
	r.mockConfig.On("GetString", "database.connections.sqlite.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.sqlite.singular").Return(false)
	r.single()
	r.basic()
}

func (r *MockSqlite) ReadWrite(config TestReadWriteConfig) {
	r.mockConfig = &mocksconfig.Config{}
	r.mockConfig.On("Get", "database.connections.sqlite.read").Return([]database.Config{
		{Database: config.ReadDatabase},
	})
	r.mockConfig.On("Get", "database.connections.sqlite.write").Return([]database.Config{
		{Database: r.database},
	})
	r.mockConfig.On("GetString", "database.connections.sqlite.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.sqlite.singular").Return(false)
	r.basic()
}

func (r *MockSqlite) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", "database.connections.sqlite.prefix").Return("goravel_")
	r.mockConfig.On("GetBool", "database.connections.sqlite.singular").Return(true)
	r.single()
	r.basic()
}

func (r *MockSqlite) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", "database.connections.sqlite.driver").Return(orm.DriverSqlite.String())
	mockPool(r.mockConfig)
}

func (r *MockSqlite) single() {
	r.mockConfig.On("Get", "database.connections.sqlite.read").Return(nil)
	r.mockConfig.On("Get", "database.connections.sqlite.write").Return(nil)
	r.mockConfig.On("GetString", "database.connections.sqlite.database").Return(r.database)
}

type MockSqlserver struct {
	driver     orm.Driver
	mockConfig *mocksconfig.Config

	database string
	password string
	user     string
	port     int
}

func NewMockSqlserver(mockConfig *mocksconfig.Config, database, username, password string, port int) *MockSqlserver {
	return &MockSqlserver{
		driver:     orm.DriverSqlserver,
		mockConfig: mockConfig,
		database:   database,
		user:       username,
		password:   password,
		port:       port,
	}
}

func (r *MockSqlserver) Common() {
	r.mockConfig.On("GetString", "database.default").Return("sqlserver")
	r.mockConfig.On("GetString", "database.migrations").Return("migrations")
	r.mockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(false)
	r.single()
	r.basic()
}

func (r *MockSqlserver) ReadWrite(config TestReadWriteConfig) {
	r.mockConfig = &mocksconfig.Config{}
	r.mockConfig.On("Get", "database.connections.sqlserver.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.ReadPort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("Get", "database.connections.sqlserver.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: config.WritePort, Username: r.user, Password: r.password},
	})
	r.mockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("")
	r.mockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(false)
	r.basic()
}

func (r *MockSqlserver) WithPrefixAndSingular() {
	r.mockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("goravel_")
	r.mockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(true)
	r.single()
	r.basic()
}

func (r *MockSqlserver) basic() {
	r.mockConfig.On("GetBool", "app.debug").Return(true)
	r.mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(orm.DriverSqlserver.String())
	r.mockConfig.On("GetString", "database.connections.sqlserver.database").Return(r.database)
	r.mockConfig.On("GetString", "database.connections.sqlserver.charset").Return("utf8mb4")
	mockPool(r.mockConfig)
}

func (r *MockSqlserver) single() {
	r.mockConfig.On("Get", "database.connections.sqlserver.read").Return(nil)
	r.mockConfig.On("Get", "database.connections.sqlserver.write").Return(nil)
	r.mockConfig.On("GetString", "database.connections.sqlserver.host").Return("127.0.0.1")
	r.mockConfig.On("GetString", "database.connections.sqlserver.username").Return(r.user)
	r.mockConfig.On("GetString", "database.connections.sqlserver.password").Return(r.password)
	r.mockConfig.On("GetInt", "database.connections.sqlserver.port").Return(r.port)
}

type testTables struct {
	driver orm.Driver
}

func newTestTables(driver orm.Driver) *testTables {
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE peoples (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE peoples (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE reviews (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE reviews (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE products (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE products (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
		return `
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgres:
		return `
CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
		return `
CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
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
	case orm.DriverMysql:
		return `
CREATE TABLE goravel_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgres:
		return `
CREATE TABLE goravel_user (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE goravel_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
		return `
CREATE TABLE goravel_user (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) NOT NULL,
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
	case orm.DriverMysql:
		return `
CREATE TABLE addresses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_addresses_created_at (created_at),
  KEY idx_addresses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgres:
		return `
CREATE TABLE addresses (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE addresses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case orm.DriverSqlserver:
		return `
CREATE TABLE addresses (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE books (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE books (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE authors (
  id SERIAL PRIMARY KEY NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE authors (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
		return `
CREATE TABLE roles (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE roles (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
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
	case orm.DriverSqlite:
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
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
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
	case orm.DriverPostgres:
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
	case orm.DriverSqlite:
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
	case orm.DriverSqlserver:
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
	case orm.DriverMysql:
		return `
CREATE TABLE role_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint(20) unsigned NOT NULL,
  user_id bigint(20) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgres:
		return `
CREATE TABLE role_user (
  id SERIAL PRIMARY KEY NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE role_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case orm.DriverSqlserver:
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
