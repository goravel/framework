package gorm

import (
	"context"
	"errors"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	supportdocker "github.com/goravel/framework/support/docker"
)

const (
	dbDatabase  = "goravel"
	dbDatabase1 = "goravel1"
)

var testContext context.Context

type MysqlDocker struct {
	MockConfig *mocksconfig.Config
	Port       int
	user       string
	password   string
	database   string
}

func NewMysqlDocker(database *supportdocker.Database) *MysqlDocker {
	config := database.Mysql.Config()

	return &MysqlDocker{MockConfig: &mocksconfig.Config{}, Port: config.Port, user: config.Username, password: config.Password, database: config.Database}
}

func NewMysql1Docker(database *supportdocker.Database) *MysqlDocker {
	config := database.Mysql1.Config()

	return &MysqlDocker{MockConfig: &mocksconfig.Config{}, Port: config.Port, user: config.Username, password: config.Password, database: config.Database}
}

func (r *MysqlDocker) New() (orm.Query, error) {
	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *MysqlDocker) NewWithPrefixAndSingular() (orm.Query, error) {
	r.mockWithPrefixAndSingular()

	db, err := r.QueryWithPrefixAndSingular()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *MysqlDocker) Query(createTable bool) (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverMysql.String())
	if err != nil {
		return nil, errors.New("connect to mysql failed")
	}

	if createTable {
		err := Tables{}.Create(orm.DriverMysql, query)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (r *MysqlDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverMysql.String())
	if err != nil {
		return nil, errors.New("connect to mysql failed")
	}

	err = Tables{}.CreateWithPrefixAndSingular(orm.DriverMysql, query)
	if err != nil {
		return nil, err
	}

	return query, nil
}

func (r *MysqlDocker) MockReadWrite(readPort, writePort int) {
	r.MockConfig = &mocksconfig.Config{}
	r.MockConfig.On("Get", "database.connections.mysql.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: r.user, Password: r.password},
	})
	r.MockConfig.On("Get", "database.connections.mysql.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: r.user, Password: r.password},
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
	r.MockConfig.On("GetString", "database.connections.mysql.username").Return(r.user)
	r.MockConfig.On("GetString", "database.connections.mysql.password").Return(r.password)
	r.MockConfig.On("GetInt", "database.connections.mysql.port").Return(r.Port)
}

func (r *MysqlDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.mysql.driver").Return(orm.DriverMysql.String())
	r.MockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4")
	r.MockConfig.On("GetString", "database.connections.mysql.loc").Return("Local")
	r.MockConfig.On("GetString", "database.connections.mysql.database").Return(r.database)

	mockPool(r.MockConfig)
}

type PostgresqlDocker struct {
	MockConfig *mocksconfig.Config
	Port       int
	user       string
	database   string
	password   string
}

func NewPostgresqlDocker(database *supportdocker.Database) *PostgresqlDocker {
	config := database.Postgresql.Config()

	return &PostgresqlDocker{MockConfig: &mocksconfig.Config{}, Port: config.Port, user: config.Username, password: config.Password, database: config.Database}
}

func (r *PostgresqlDocker) New() (orm.Query, error) {
	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *PostgresqlDocker) NewWithPrefixAndSingular() (orm.Query, error) {
	r.mockWithPrefixAndSingular()

	db, err := r.QueryWithPrefixAndSingular()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *PostgresqlDocker) Query(createTable bool) (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverPostgresql.String())
	if err != nil {
		return nil, errors.New("connect to postgresql failed")
	}

	if createTable {
		err := Tables{}.Create(orm.DriverPostgresql, query)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (r *PostgresqlDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverPostgresql.String())
	if err != nil {
		return nil, errors.New("connect to postgresql failed")
	}

	err = Tables{}.CreateWithPrefixAndSingular(orm.DriverPostgresql, query)
	if err != nil {
		return nil, err
	}

	return query, nil
}

func (r *PostgresqlDocker) MockReadWrite(readPort, writePort int) {
	r.MockConfig = &mocksconfig.Config{}
	r.MockConfig.On("Get", "database.connections.postgresql.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: r.user, Password: r.password},
	})
	r.MockConfig.On("Get", "database.connections.postgresql.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: r.user, Password: r.password},
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
	r.MockConfig.On("GetString", "database.connections.postgresql.username").Return(r.user)
	r.MockConfig.On("GetString", "database.connections.postgresql.password").Return(r.password)
	r.MockConfig.On("GetInt", "database.connections.postgresql.port").Return(r.Port)
}

func (r *PostgresqlDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.postgresql.driver").Return(orm.DriverPostgresql.String())
	r.MockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	r.MockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	r.MockConfig.On("GetString", "database.connections.postgresql.database").Return(r.database)

	mockPool(r.MockConfig)
}

type SqliteDocker struct {
	name       string
	MockConfig *mocksconfig.Config
}

func NewSqliteDocker(dbName string) *SqliteDocker {
	return &SqliteDocker{MockConfig: &mocksconfig.Config{}, name: dbName}
}

func (r *SqliteDocker) New() (orm.Query, error) {
	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqliteDocker) NewWithPrefixAndSingular() (orm.Query, error) {
	r.mockWithPrefixAndSingular()

	db, err := r.QueryWithPrefixAndSingular()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqliteDocker) Query(createTable bool) (orm.Query, error) {
	db, err := InitializeQuery(testContext, r.MockConfig, orm.DriverSqlite.String())
	if err != nil {
		return nil, err
	}

	if createTable {
		err = Tables{}.Create(orm.DriverSqlite, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (r *SqliteDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	db, err := InitializeQuery(testContext, r.MockConfig, orm.DriverSqlite.String())
	if err != nil {
		return nil, err
	}

	err = Tables{}.CreateWithPrefixAndSingular(orm.DriverSqlite, db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqliteDocker) MockReadWrite() {
	r.MockConfig = &mocksconfig.Config{}
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

type SqlserverDocker struct {
	MockConfig *mocksconfig.Config
	Port       int
	user       string
	database   string
	password   string
}

func NewSqlserverDocker(database *supportdocker.Database) *SqlserverDocker {
	config := database.Sqlserver.Config()

	return &SqlserverDocker{MockConfig: &mocksconfig.Config{}, Port: config.Port, user: config.Username, password: config.Password, database: config.Database}
}

func (r *SqlserverDocker) New() (orm.Query, error) {
	r.mock()

	db, err := r.Query(true)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqlserverDocker) NewWithPrefixAndSingular() (orm.Query, error) {
	r.mockWithPrefixAndSingular()

	db, err := r.QueryWithPrefixAndSingular()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *SqlserverDocker) Query(createTable bool) (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverSqlserver.String())
	if err != nil {
		return nil, errors.New("connect to sqlserver failed")
	}

	if createTable {
		err := Tables{}.Create(orm.DriverSqlserver, query)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (r *SqlserverDocker) QueryWithPrefixAndSingular() (orm.Query, error) {
	query, err := InitializeQuery(testContext, r.MockConfig, orm.DriverSqlserver.String())
	if err != nil {
		return nil, errors.New("connect to sqlserver failed")
	}

	err = Tables{}.CreateWithPrefixAndSingular(orm.DriverSqlserver, query)
	if err != nil {
		return nil, err
	}

	return query, nil
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
	r.MockConfig = &mocksconfig.Config{}
	r.MockConfig.On("Get", "database.connections.sqlserver.read").Return([]database.Config{
		{Host: "127.0.0.1", Port: readPort, Username: r.user, Password: r.password},
	})
	r.MockConfig.On("Get", "database.connections.sqlserver.write").Return([]database.Config{
		{Host: "127.0.0.1", Port: writePort, Username: r.user, Password: r.password},
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
	r.MockConfig.On("GetString", "database.connections.sqlserver.username").Return(r.user)
	r.MockConfig.On("GetString", "database.connections.sqlserver.password").Return(r.password)
	r.MockConfig.On("GetInt", "database.connections.sqlserver.port").Return(r.Port)
}

func (r *SqlserverDocker) mockOfCommon() {
	r.MockConfig.On("GetBool", "app.debug").Return(true)
	r.MockConfig.On("GetString", "database.connections.sqlserver.driver").Return(orm.DriverSqlserver.String())
	r.MockConfig.On("GetString", "database.connections.sqlserver.database").Return(r.database)
	r.MockConfig.On("GetString", "database.connections.sqlserver.charset").Return("utf8mb4")
	mockPool(r.MockConfig)
}

type Tables struct {
}

func (r Tables) Create(driver orm.Driver, db orm.Query) error {
	_, err := db.Exec(r.createPeopleTable(driver))
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

func (r Tables) CreateWithPrefixAndSingular(driver orm.Driver, db orm.Query) error {
	_, err := db.Exec(r.createUserTableWithPrefixAndSingular(driver))
	if err != nil {
		return err
	}

	return nil
}

func (r Tables) createPeopleTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Tables) createReviewTable(driver orm.Driver) string {
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

func (r Tables) createProductTable(driver orm.Driver) string {
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

func (r Tables) createUserTable(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Tables) createUserTableWithPrefixAndSingular(driver orm.Driver) string {
	switch driver {
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
	case orm.DriverPostgresql:
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

func (r Tables) createAddressTable(driver orm.Driver) string {
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

func (r Tables) createBookTable(driver orm.Driver) string {
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

func (r Tables) createAuthorTable(driver orm.Driver) string {
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

func (r Tables) createRoleTable(driver orm.Driver) string {
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

func (r Tables) createHouseTable(driver orm.Driver) string {
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

func (r Tables) createPhoneTable(driver orm.Driver) string {
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

func (r Tables) createRoleUserTable(driver orm.Driver) string {
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

func mockPool(mockConfig *mocksconfig.Config) {
	mockConfig.On("GetInt", "database.pool.max_idle_conns", 10).Return(10)
	mockConfig.On("GetInt", "database.pool.max_open_conns", 100).Return(100)
	mockConfig.On("GetInt", "database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.On("GetInt", "database.pool.conn_max_lifetime", 3600).Return(3600)
}
