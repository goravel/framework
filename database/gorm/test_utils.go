package gorm

import (
	"context"
	"strconv"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"

	"github.com/ory/dockertest/v3"
)

const (
	dbDatabase = "goravel"
	dbPassword = "Goravel(!)"
	dbUser     = "root"
)

func MysqlDocker() (*dockertest.Pool, *dockertest.Resource, ormcontract.DB, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "5.7",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=" + dbPassword,
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_ = resource.Expire(60)

	if err := pool.Retry(func() error {
		return initDatabase(ormcontract.DriverMysql, resource.GetPort("3306/tcp"))
	}); err != nil {
		return nil, nil, nil, err
	}

	db, err := getDB(ormcontract.DriverMysql, dbDatabase, resource.GetPort("3306/tcp"))
	if err != nil {
		return nil, nil, nil, err
	}

	if err := initTables(ormcontract.DriverMysql, db); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func PostgresqlDocker() (*dockertest.Pool, *dockertest.Resource, ormcontract.DB, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_USER=" + dbUser,
			"POSTGRES_PASSWORD=" + dbPassword,
			"listen_addresses = '*'",
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_ = resource.Expire(60)

	if err := pool.Retry(func() error {
		return initDatabase(ormcontract.DriverPostgresql, resource.GetPort("5432/tcp"))
	}); err != nil {
		return nil, nil, nil, err
	}

	db, err := getDB(ormcontract.DriverPostgresql, dbDatabase, resource.GetPort("5432/tcp"))
	if err != nil {
		return nil, nil, nil, err
	}

	if err := initTables(ormcontract.DriverPostgresql, db); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func SqliteDocker() (*dockertest.Pool, *dockertest.Resource, ormcontract.DB, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "nouchka/sqlite3",
		Tag:        "latest",
		Env:        []string{},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_ = resource.Expire(60)

	var db ormcontract.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = getDB(ormcontract.DriverSqlite, dbDatabase, "")

		return err
	}); err != nil {
		return nil, nil, nil, err
	}

	if err := initTables(ormcontract.DriverSqlite, db); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func SqlserverDocker() (*dockertest.Pool, *dockertest.Resource, ormcontract.DB, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "mcr.microsoft.com/mssql/server",
		Tag:        "2022-latest",
		Env: []string{
			"MSSQL_SA_PASSWORD=" + dbPassword,
			"ACCEPT_EULA=Y",
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	_ = resource.Expire(60)

	if err := pool.Retry(func() error {
		return initDatabase(ormcontract.DriverSqlserver, resource.GetPort("1433/tcp"))
	}); err != nil {
		return nil, nil, nil, err
	}

	db, err := getDB(ormcontract.DriverSqlserver, dbDatabase, resource.GetPort("1433/tcp"))
	if err != nil {
		return nil, nil, nil, err
	}

	if err := initTables(ormcontract.DriverSqlserver, db); err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func initDatabase(connection ormcontract.Driver, port string) error {
	var (
		database  = ""
		createSql = ""
	)

	switch connection {
	case ormcontract.DriverMysql:
		database = "mysql"
		createSql = "CREATE DATABASE `goravel` DEFAULT CHARACTER SET = `utf8mb4` DEFAULT COLLATE = `utf8mb4_general_ci`;"
	case ormcontract.DriverPostgresql:
		database = "postgres"
		createSql = "CREATE DATABASE goravel;"
	case ormcontract.DriverSqlserver:
		database = "msdb"
		createSql = "CREATE DATABASE goravel;"
	}

	db, err := getDB(connection, database, port)
	if err != nil {
		return err
	}

	if err := db.Exec(createSql); err != nil {
		return err
	}

	return nil
}

func getDB(driver ormcontract.Driver, database, port string) (ormcontract.DB, error) {
	mockConfig := mock.Config()
	switch driver {
	case ormcontract.DriverMysql:
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("GetString", "database.connections.mysql.driver").Return(ormcontract.DriverMysql.String()).Once()
		mockConfig.On("GetString", "database.connections.mysql.host").Return("localhost").Once()
		mockConfig.On("GetString", "database.connections.mysql.port").Return(port).Once()
		mockConfig.On("GetString", "database.connections.mysql.database").Return(database).Once()
		mockConfig.On("GetString", "database.connections.mysql.username").Return(dbUser).Once()
		mockConfig.On("GetString", "database.connections.mysql.password").Return(dbPassword).Once()
		mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Once()
		mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local").Once()
	case ormcontract.DriverPostgresql:
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("GetString", "database.connections.postgresql.driver").Return(ormcontract.DriverPostgresql.String()).Once()
		mockConfig.On("GetString", "database.connections.postgresql.host").Return("localhost").Once()
		mockConfig.On("GetString", "database.connections.postgresql.port").Return(port).Once()
		mockConfig.On("GetString", "database.connections.postgresql.database").Return(database).Once()
		mockConfig.On("GetString", "database.connections.postgresql.username").Return(dbUser).Once()
		mockConfig.On("GetString", "database.connections.postgresql.password").Return(dbPassword).Once()
		mockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable").Once()
		mockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC").Once()
	case ormcontract.DriverSqlite:
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("GetString", "database.connections.sqlite.driver").Return(ormcontract.DriverSqlite.String()).Once()
		mockConfig.On("GetString", "database.connections.sqlite.database").Return(database).Once()
	case ormcontract.DriverSqlserver:
		mockConfig.On("GetBool", "app.debug").Return(true).Once()
		mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(ormcontract.DriverSqlserver.String()).Once()
		mockConfig.On("GetString", "database.connections.sqlserver.host").Return("localhost").Once()
		mockConfig.On("GetString", "database.connections.sqlserver.port").Return(port).Once()
		mockConfig.On("GetString", "database.connections.sqlserver.database").Return(database).Once()
		mockConfig.On("GetString", "database.connections.sqlserver.username").Return("sa").Once()
		mockConfig.On("GetString", "database.connections.sqlserver.password").Return(dbPassword).Once()
	}

	return NewDB(context.Background(), driver.String())
}

func initTables(driver ormcontract.Driver, db ormcontract.DB) error {
	if err := db.Exec(createUserTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createUserAddressTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createUserBookTable(driver)); err != nil {
		return err
	}

	return nil
}

func createUserTable(driver ormcontract.Driver) string {
	switch driver {
	case ormcontract.DriverMysql:
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
	case ormcontract.DriverPostgresql:
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
	case ormcontract.DriverSqlite:
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
	case ormcontract.DriverSqlserver:
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

func createUserAddressTable(driver ormcontract.Driver) string {
	switch driver {
	case ormcontract.DriverMysql:
		return `
CREATE TABLE user_addresses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned NOT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at datetime(3) DEFAULT NULL,
  updated_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_user_addresses_created_at (created_at),
  KEY idx_user_addresses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case ormcontract.DriverPostgresql:
		return `
CREATE TABLE user_addresses (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int NOT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case ormcontract.DriverSqlite:
		return `
CREATE TABLE user_addresses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int NOT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case ormcontract.DriverSqlserver:
		return `
CREATE TABLE user_addresses (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint NOT NULL,
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

func createUserBookTable(driver ormcontract.Driver) string {
	switch driver {
	case ormcontract.DriverMysql:
		return `
CREATE TABLE user_books (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime(3) DEFAULT NULL,
  updated_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_user_addresses_created_at (created_at),
  KEY idx_user_addresses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case ormcontract.DriverPostgresql:
		return `
CREATE TABLE user_books (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case ormcontract.DriverSqlite:
		return `
CREATE TABLE user_books (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case ormcontract.DriverSqlserver:
		return `
CREATE TABLE user_books (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint NOT NULL,
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

func paginator(page string, limit string) func(methods ormcontract.Query) ormcontract.Query {
	return func(query ormcontract.Query) ormcontract.Query {
		page, _ := strconv.Atoi(page)
		limit, _ := strconv.Atoi(limit)
		offset := (page - 1) * limit

		return query.Offset(offset).Limit(limit)
	}
}
