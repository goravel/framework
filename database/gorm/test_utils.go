package gorm

import (
	"context"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	testingdocker "github.com/goravel/framework/support/docker"
)

const (
	DbPassword     = "Goravel(!)"
	DbUser         = "root"
	dbDatabase     = "goravel"
	dbDatabase1    = "goravel1"
	dbUser1        = "sa"
	resourceExpire = 1200
)

var testContext context.Context

type MysqlDocker struct {
	MockConfig *configmock.Config
	Port       int
	pool       *dockertest.Pool
}

func NewMysqlDocker() *MysqlDocker {
	return &MysqlDocker{MockConfig: &configmock.Config{}}
}

func (r *MysqlDocker) New() (*dockertest.Pool, *dockertest.Resource, orm.Query, error) {
	pool, resource, err := r.Init()
	if err != nil {
		return nil, nil, nil, err
	}

	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func (r *MysqlDocker) Init() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=" + DbPassword,
			"MYSQL_DATABASE=" + dbDatabase,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)
	r.pool = pool
	r.Port = cast.ToInt(resource.GetPort("3306/tcp"))

	return pool, resource, nil
}

func (r *MysqlDocker) Query(createTable bool) (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	if createTable {
		err = Table{}.Create(orm.DriverMysql, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (r *MysqlDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	err = Table{}.CreateWithPrefixAndSingular(orm.DriverMysql, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *MysqlDocker) MockReadWrite(readPort, writePort int) {
	r.MockConfig = &configmock.Config{}
	r.MockConfig.On("Get", "database.connections.mysql.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: DbUser, Password: DbPassword},
	})
	r.MockConfig.On("Get", "database.connections.mysql.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: DbUser, Password: DbPassword},
	})
	r.MockConfig.On("GetString", "database.connections.mysql.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.mysql.singular").Return(false)
	r.mockOfCommon()
}

func (r *MysqlDocker) mock() {
	r.MockConfig.On("GetString", "database.default").Return("mysql")
	r.MockConfig.On("GetString", "database.migrations").Return("migrations")
	r.MockConfig.On("GetString", "database.connections.mysql.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.mysql.singular").Return(false)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *MysqlDocker) mockWithPrefixAndSingular() {
	r.MockConfig.On("GetString", "database.connections.mysql.prefix").Return("goravel_")
	r.MockConfig.On("GetBool", "database.connections.mysql.singular").Return(true)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *MysqlDocker) mockSingleOfCommon() {
	r.MockConfig.On("Get", "database.connections.mysql.read").Return(nil)
	r.MockConfig.On("Get", "database.connections.mysql.write").Return(nil)
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.mysql.host").Return("127.0.0.1")
	r.MockConfig.On("GetString", "database.connections.mysql.username").Return(DbUser)
	r.MockConfig.On("GetString", "database.connections.mysql.password").Return(DbPassword)
	r.MockConfig.On("GetInt", "database.connections.mysql.port").Return(r.Port)
}

func (r *MysqlDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.mysql.driver").Return(orm.DriverMysql.String())
	r.MockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4")
	r.MockConfig.On("GetString", "database.connections.mysql.loc").Return("Local")
	r.MockConfig.On("GetString", "database.connections.mysql.database").Return(dbDatabase)

	mockPool(r.MockConfig)
}

func (r *MysqlDocker) query() (orm.Query, error) {
	var db orm.Query
	if err := r.pool.Retry(func() error {
		var err error
		db, err = InitializeQuery(testContext, r.MockConfig, orm.DriverMysql.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

type PostgresqlDocker struct {
	pool       *dockertest.Pool
	MockConfig *configmock.Config
	Port       int
}

func NewPostgresqlDocker() *PostgresqlDocker {
	return &PostgresqlDocker{MockConfig: &configmock.Config{}}
}

func (r *PostgresqlDocker) New() (*dockertest.Pool, *dockertest.Resource, orm.Query, error) {
	pool, resource, err := r.Init()
	if err != nil {
		return nil, nil, nil, err
	}

	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func (r *PostgresqlDocker) Init() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + DbUser,
			"POSTGRES_PASSWORD=" + DbPassword,
			"listen_addresses = '*'",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)
	r.pool = pool
	r.Port = cast.ToInt(resource.GetPort("5432/tcp"))

	return pool, resource, nil
}

func (r *PostgresqlDocker) Query(createTable bool) (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	if createTable {
		err = Table{}.Create(orm.DriverPostgresql, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (r *PostgresqlDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	err = Table{}.CreateWithPrefixAndSingular(orm.DriverPostgresql, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *PostgresqlDocker) MockReadWrite(readPort, writePort int) {
	r.MockConfig = &configmock.Config{}
	r.MockConfig.On("Get", "database.connections.postgresql.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: DbUser, Password: DbPassword},
	})
	r.MockConfig.On("Get", "database.connections.postgresql.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: DbUser, Password: DbPassword},
	})
	r.MockConfig.On("GetString", "database.connections.postgresql.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.postgresql.singular").Return(false)
	r.mockOfCommon()
}

func (r *PostgresqlDocker) mock() {
	r.MockConfig.On("GetString", "database.default").Return("postgresql")
	r.MockConfig.On("GetString", "database.migrations").Return("migrations")
	r.MockConfig.On("GetString", "database.connections.postgresql.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.postgresql.singular").Return(false)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *PostgresqlDocker) mockWithPrefixAndSingular() {
	r.MockConfig.On("GetString", "database.connections.postgresql.prefix").Return("goravel_")
	r.MockConfig.On("GetBool", "database.connections.postgresql.singular").Return(true)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *PostgresqlDocker) mockSingleOfCommon() {
	r.MockConfig.On("Get", "database.connections.postgresql.read").Return(nil)
	r.MockConfig.On("Get", "database.connections.postgresql.write").Return(nil)
	r.MockConfig.On("GetString", "database.connections.postgresql.host").Return("127.0.0.1")
	r.MockConfig.On("GetString", "database.connections.postgresql.username").Return(DbUser)
	r.MockConfig.On("GetString", "database.connections.postgresql.password").Return(DbPassword)
	r.MockConfig.On("GetInt", "database.connections.postgresql.port").Return(r.Port)
}

func (r *PostgresqlDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.postgresql.driver").Return(orm.DriverPostgresql.String())
	r.MockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	r.MockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	r.MockConfig.On("GetString", "database.connections.postgresql.database").Return("postgres")

	mockPool(r.MockConfig)
}

func (r *PostgresqlDocker) query() (orm.Query, error) {
	var db orm.Query
	if err := r.pool.Retry(func() error {
		var err error
		db, err = InitializeQuery(testContext, r.MockConfig, orm.DriverPostgresql.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

type SqliteDocker struct {
	name       string
	MockConfig *configmock.Config
	pool       *dockertest.Pool
}

func NewSqliteDocker(dbName string) *SqliteDocker {
	return &SqliteDocker{MockConfig: &configmock.Config{}, name: dbName}
}

func (r *SqliteDocker) New() (*dockertest.Pool, *dockertest.Resource, orm.Query, error) {
	pool, resource, err := r.Init()
	if err != nil {
		return nil, nil, nil, err
	}

	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func (r *SqliteDocker) Init() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "nouchka/sqlite3",
		Tag:        "latest",
		Env:        []string{},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)
	r.pool = pool

	return pool, resource, nil
}

func (r *SqliteDocker) Query(createTable bool) (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	if createTable {
		err = Table{}.Create(orm.DriverSqlite, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (r *SqliteDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	err = Table{}.CreateWithPrefixAndSingular(orm.DriverSqlite, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqliteDocker) MockReadWrite() {
	r.MockConfig = &configmock.Config{}
	r.MockConfig.On("Get", "database.connections.sqlite.read").Return([]database.Config{
		{Database: dbDatabase},
	})
	r.MockConfig.On("Get", "database.connections.sqlite.write").Return([]database.Config{
		{Database: dbDatabase1},
	})
	r.MockConfig.On("GetString", "database.connections.sqlite.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.sqlite.singular").Return(false)
	r.mockOfCommon()
}

func (r *SqliteDocker) mock() {
	r.MockConfig.On("GetString", "database.default").Return("sqlite")
	r.MockConfig.On("GetString", "database.migrations").Return("migrations")
	r.MockConfig.On("GetString", "database.connections.sqlite.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.sqlite.singular").Return(false)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *SqliteDocker) mockWithPrefixAndSingular() {
	r.MockConfig.On("GetString", "database.connections.sqlite.prefix").Return("goravel_")
	r.MockConfig.On("GetBool", "database.connections.sqlite.singular").Return(true)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *SqliteDocker) mockSingleOfCommon() {
	r.MockConfig.On("Get", "database.connections.sqlite.read").Return(nil)
	r.MockConfig.On("Get", "database.connections.sqlite.write").Return(nil)
	r.MockConfig.On("GetString", "database.connections.sqlite.database").Return(r.name)
}

func (r *SqliteDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.sqlite.driver").Return(orm.DriverSqlite.String())
	mockPool(r.MockConfig)
}

func (r *SqliteDocker) query() (orm.Query, error) {
	var db orm.Query
	if err := r.pool.Retry(func() error {
		var err error
		db, err = InitializeQuery(testContext, r.MockConfig, orm.DriverSqlite.String())

		return err
	}); err != nil {
		return nil, err
	}

	return db, nil
}

type SqlserverDocker struct {
	pool       *dockertest.Pool
	MockConfig *configmock.Config
	Port       int
}

func NewSqlserverDocker() *SqlserverDocker {
	return &SqlserverDocker{MockConfig: &configmock.Config{}}
}

func (r *SqlserverDocker) New() (*dockertest.Pool, *dockertest.Resource, orm.Query, error) {
	pool, resource, err := r.Init()
	if err != nil {
		return nil, nil, nil, err
	}

	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func (r *SqlserverDocker) Init() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "mcr.microsoft.com/mssql/server",
		Tag:        "latest",
		Env: []string{
			"MSSQL_SA_PASSWORD=" + DbPassword,
			"ACCEPT_EULA=Y",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)
	r.pool = pool
	r.Port = cast.ToInt(resource.GetPort("1433/tcp"))

	return pool, resource, nil
}

func (r *SqlserverDocker) Query(createTable bool) (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	if createTable {
		err = Table{}.Create(orm.DriverSqlserver, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (r *SqlserverDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	db, err := r.query()
	if err != nil {
		return nil, err
	}

	err = Table{}.CreateWithPrefixAndSingular(orm.DriverSqlserver, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqlserverDocker) mock() {
	r.MockConfig.On("GetString", "database.default").Return("sqlserver")
	r.MockConfig.On("GetString", "database.migrations").Return("migrations")
	r.MockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(false)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *SqlserverDocker) MockReadWrite(readPort, writePort int) {
	r.MockConfig = &configmock.Config{}
	r.MockConfig.On("Get", "database.connections.sqlserver.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: dbUser1, Password: DbPassword},
	})
	r.MockConfig.On("Get", "database.connections.sqlserver.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: dbUser1, Password: DbPassword},
	})
	r.MockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("")
	r.MockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(false)
	r.mockOfCommon()
}

func (r *SqlserverDocker) mockWithPrefixAndSingular() {
	r.MockConfig.On("GetString", "database.connections.sqlserver.prefix").Return("goravel_")
	r.MockConfig.On("GetBool", "database.connections.sqlserver.singular").Return(true)
	r.mockSingleOfCommon()
	r.mockOfCommon()
}

func (r *SqlserverDocker) mockSingleOfCommon() {
	r.MockConfig.On("Get", "database.connections.sqlserver.read").Return(nil)
	r.MockConfig.On("Get", "database.connections.sqlserver.write").Return(nil)
	r.MockConfig.On("GetString", "database.connections.sqlserver.host").Return("127.0.0.1")
	r.MockConfig.On("GetString", "database.connections.sqlserver.username").Return(dbUser1)
	r.MockConfig.On("GetString", "database.connections.sqlserver.password").Return(DbPassword)
	r.MockConfig.On("GetInt", "database.connections.sqlserver.port").Return(r.Port)
}

func (r *SqlserverDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.sqlserver.driver").Return(orm.DriverSqlserver.String())
	r.MockConfig.On("GetString", "database.connections.sqlserver.database").Return("msdb")
	r.MockConfig.On("GetString", "database.connections.sqlserver.charset").Return("utf8mb4")
	mockPool(r.MockConfig)
}

func (r *SqlserverDocker) query() (orm.Query, error) {
	var db orm.Query
	if err := r.pool.Retry(func() error {
		var err error
		db, err = InitializeQuery(testContext, r.MockConfig, orm.DriverSqlserver.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

type Table struct {
}

func (r Table) Create(driver orm.Driver, db orm.Query) error {
	_, err := db.Exec(r.createPersonTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createReviewTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createUserTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createProductTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createAddressTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createBookTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createRoleTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createHouseTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createPhoneTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createRoleUserTable(driver))
	if err != nil {
		return err
	}
	_, err = db.Exec(r.createAuthorTable(driver))
	if err != nil {
		return err
	}

	return nil
}

func (r Table) CreateWithPrefixAndSingular(driver orm.Driver, db orm.Query) error {
	_, err := db.Exec(r.createUserTableWithPrefixAndSingular(driver))
	if err != nil {
		return err
	}

	return nil
}

func (r Table) createPersonTable(driver orm.Driver) string {
	switch driver {
	case orm.DriverMysql:
		return `
CREATE TABLE people (
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
	case orm.DriverPostgresql:
		return `
CREATE TABLE people (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case orm.DriverSqlite:
		return `
CREATE TABLE people (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case orm.DriverSqlserver:
		return `
CREATE TABLE people (
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

func (r Table) createReviewTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createProductTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createUserTable(driver orm.Driver) string {
	switch driver {
	case orm.DriverMysql:
		return `
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgresql:
		return `
CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
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

func (r Table) createUserTableWithPrefixAndSingular(driver orm.Driver) string {
	switch driver {
	case orm.DriverMysql:
		return `
CREATE TABLE goravel_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  avatar varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgresql:
		return `
CREATE TABLE goravel_user (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
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

func (r Table) createAddressTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createBookTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createAuthorTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createRoleTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createHouseTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createPhoneTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Table) createRoleUserTable(driver orm.Driver) string {
	switch driver {
	case orm.DriverMysql:
		return `
CREATE TABLE role_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint(20) unsigned NOT NULL,
  user_id bigint(20) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case orm.DriverPostgresql:
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

func mockPool(mockConfig *configmock.Config) {
	mockConfig.On("GetInt", "database.pool.max_idle_conns", 10).Return(10)
	mockConfig.On("GetInt", "database.pool.max_open_conns", 100).Return(100)
	mockConfig.On("GetInt", "database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.On("GetInt", "database.pool.conn_max_lifetime", 3600).Return(3600)
}
