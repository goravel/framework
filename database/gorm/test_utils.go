package gorm

import (
	"context"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	testingdocker "github.com/goravel/framework/testing/docker"
	"github.com/goravel/framework/testing/mock"
)

const (
	dbDatabase     = "goravel"
	dbDatabase1    = "goravel1"
	dbPassword     = "Goravel(!)"
	dbUser         = "root"
	dbUser1        = "sa"
	resourceExpire = 600
)

func MysqlDocker() (*dockertest.Pool, *dockertest.Resource, contractsorm.DB, error) {
	pool, resource, err := initMysqlDocker()
	if err != nil {
		return nil, nil, nil, err
	}

	mockSingleMysql(cast.ToInt(resource.GetPort("3306/tcp")))

	db, err := mysqlDockerDB(pool, true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func PostgresqlDocker() (*dockertest.Pool, *dockertest.Resource, contractsorm.DB, error) {
	pool, resource, err := initPostgresqlDocker()
	if err != nil {
		return nil, nil, nil, err
	}

	mockSinglePostgresql(cast.ToInt(resource.GetPort("5432/tcp")))

	db, err := postgresqlDockerDB(pool, true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func SqliteDocker(dbName string) (*dockertest.Pool, *dockertest.Resource, contractsorm.DB, error) {
	pool, resource, err := initSqliteDocker()
	if err != nil {
		return nil, nil, nil, err
	}

	mockSingleSqlite(dbName)

	db, err := sqliteDockerDB(pool, true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func SqlserverDocker() (*dockertest.Pool, *dockertest.Resource, contractsorm.DB, error) {
	pool, resource, err := initSqlserverDocker()
	if err != nil {
		return nil, nil, nil, err
	}

	mockSingleSqlserver(cast.ToInt(resource.GetPort("1433/tcp")))

	db, err := sqlserverDockerDB(pool, true)
	if err != nil {
		return nil, nil, nil, err
	}

	return pool, resource, db, nil
}

func mockSingleMysql(port int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.mysql.read").Return(nil)
	mockConfig.On("Get", "database.connections.mysql.write").Return(nil)
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return(contractsorm.DriverMysql.String())
	mockConfig.On("GetString", "database.connections.mysql.host").Return("localhost")
	mockConfig.On("GetString", "database.connections.mysql.username").Return(dbUser)
	mockConfig.On("GetString", "database.connections.mysql.password").Return(dbPassword)
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4")
	mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local")
	mockConfig.On("GetString", "database.connections.mysql.database").Return("mysql")
	mockConfig.On("GetInt", "database.connections.mysql.port").Return(port)
}

func mockReadWriteMysql(readPort, writePort int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.mysql.read").Return([]database.Config{
		{Host: "localhost", Port: readPort, Username: dbUser, Password: dbPassword},
	})
	mockConfig.On("Get", "database.connections.mysql.write").Return([]database.Config{
		{Host: "localhost", Port: writePort, Username: dbUser, Password: dbPassword},
	})
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.mysql.driver").Return(contractsorm.DriverMysql.String())
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4")
	mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local")
	mockConfig.On("GetString", "database.connections.mysql.database").Return("mysql")
	mockConfig.On("GetString", "database.connections.mysql.database").Return(dbDatabase)
}

func mysqlDockerDB(pool *dockertest.Pool, createTable bool) (contractsorm.DB, error) {
	db, err := initMysql(pool)
	if err != nil {
		return nil, err
	}

	if createTable {
		if err := initTables(contractsorm.DriverMysql, db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func initMysqlDocker() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "5.7",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=" + dbPassword,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)

	return pool, resource, nil
}

func initMysql(pool *dockertest.Pool) (contractsorm.DB, error) {
	var db contractsorm.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = NewDB(context.Background(), contractsorm.DriverMysql.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

func mockSinglePostgresql(port int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.postgresql.read").Return(nil)
	mockConfig.On("Get", "database.connections.postgresql.write").Return(nil)
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.postgresql.driver").Return(contractsorm.DriverPostgresql.String())
	mockConfig.On("GetString", "database.connections.postgresql.host").Return("localhost")
	mockConfig.On("GetString", "database.connections.postgresql.username").Return(dbUser)
	mockConfig.On("GetString", "database.connections.postgresql.password").Return(dbPassword)
	mockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	mockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	mockConfig.On("GetString", "database.connections.postgresql.database").Return("postgres")
	mockConfig.On("GetInt", "database.connections.postgresql.port").Return(port)
}

func mockReadWritePostgresql(readPort, writePort int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.postgresql.read").Return([]database.Config{
		{Host: "localhost", Port: readPort, Username: dbUser, Password: dbPassword},
	})
	mockConfig.On("Get", "database.connections.postgresql.write").Return([]database.Config{
		{Host: "localhost", Port: writePort, Username: dbUser, Password: dbPassword},
	})
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.postgresql.driver").Return(contractsorm.DriverPostgresql.String())
	mockConfig.On("GetString", "database.connections.postgresql.sslmode").Return("disable")
	mockConfig.On("GetString", "database.connections.postgresql.timezone").Return("UTC")
	mockConfig.On("GetString", "database.connections.postgresql.database").Return("postgres")
}

func postgresqlDockerDB(pool *dockertest.Pool, createTable bool) (contractsorm.DB, error) {
	db, err := initPostgresql(pool)
	if err != nil {
		return nil, err
	}

	if createTable {
		if err := initTables(contractsorm.DriverPostgresql, db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func initPostgresqlDocker() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)

	return pool, resource, nil
}

func initPostgresql(pool *dockertest.Pool) (contractsorm.DB, error) {
	var db contractsorm.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = NewDB(context.Background(), contractsorm.DriverPostgresql.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

func mockSingleSqlite(dbName string) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.sqlite.read").Return(nil)
	mockConfig.On("Get", "database.connections.sqlite.write").Return(nil)
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.sqlite.driver").Return(contractsorm.DriverSqlite.String())
	mockConfig.On("GetString", "database.connections.sqlite.database").Return(dbName)
}

func mockReadWriteSqlite() {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.sqlite.read").Return([]database.Config{
		{Database: dbDatabase},
	})
	mockConfig.On("Get", "database.connections.sqlite.write").Return([]database.Config{
		{Database: dbDatabase1},
	})
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.sqlite.driver").Return(contractsorm.DriverSqlite.String())
}

func sqliteDockerDB(pool *dockertest.Pool, createTable bool) (contractsorm.DB, error) {
	db, err := initSqlite(pool)
	if err != nil {
		return nil, err
	}

	if createTable {
		if err := initTables(contractsorm.DriverSqlite, db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func initSqliteDocker() (*dockertest.Pool, *dockertest.Resource, error) {
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

	return pool, resource, nil
}

func initSqlite(pool *dockertest.Pool) (contractsorm.DB, error) {
	var db contractsorm.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = NewDB(context.Background(), contractsorm.DriverSqlite.String())

		return err
	}); err != nil {
		return nil, err
	}

	return db, nil
}

func mockSingleSqlserver(port int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.sqlserver.read").Return(nil)
	mockConfig.On("Get", "database.connections.sqlserver.write").Return(nil)
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(contractsorm.DriverSqlserver.String())
	mockConfig.On("GetString", "database.connections.sqlserver.host").Return("localhost")
	mockConfig.On("GetString", "database.connections.sqlserver.username").Return(dbUser1)
	mockConfig.On("GetString", "database.connections.sqlserver.password").Return(dbPassword)
	mockConfig.On("GetString", "database.connections.sqlserver.database").Return("msdb")
	mockConfig.On("GetString", "database.connections.sqlserver.charset").Return("utf8mb4")
	mockConfig.On("GetInt", "database.connections.sqlserver.port").Return(port)
}

func mockReadWriteSqlserver(readPort, writePort int) {
	mockConfig := mock.Config()
	mockConfig.On("Get", "database.connections.sqlserver.read").Return([]database.Config{
		{Host: "localhost", Port: readPort, Username: dbUser1, Password: dbPassword},
	})
	mockConfig.On("Get", "database.connections.sqlserver.write").Return([]database.Config{
		{Host: "localhost", Port: writePort, Username: dbUser1, Password: dbPassword},
	})
	mockConfig.On("GetBool", "app.debug").Return(true)
	mockConfig.On("GetString", "database.connections.sqlserver.driver").Return(contractsorm.DriverSqlserver.String())
	mockConfig.On("GetString", "database.connections.sqlserver.database").Return("msdb")
	mockConfig.On("GetString", "database.connections.sqlserver.charset").Return("utf8mb4")
}

func sqlserverDockerDB(pool *dockertest.Pool, createTable bool) (contractsorm.DB, error) {
	db, err := initSqlserver(pool)
	if err != nil {
		return nil, err
	}

	if createTable {
		if err := initTables(contractsorm.DriverSqlserver, db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func initSqlserverDocker() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	_ = resource.Expire(resourceExpire)

	return pool, resource, nil
}

func initSqlserver(pool *dockertest.Pool) (contractsorm.DB, error) {
	var db contractsorm.DB
	if err := pool.Retry(func() error {
		var err error
		db, err = NewDB(context.Background(), contractsorm.DriverSqlserver.String())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}

func initTables(driver contractsorm.Driver, db contractsorm.DB) error {
	if err := db.Exec(createUserTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createAddressTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createBookTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createRoleTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createHouseTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createPhoneTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createRoleUserTable(driver)); err != nil {
		return err
	}
	if err := db.Exec(createAuthorTable(driver)); err != nil {
		return err
	}

	return nil
}

func createUserTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
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
	case contractsorm.DriverSqlite:
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
	case contractsorm.DriverSqlserver:
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

func createAddressTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
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
	case contractsorm.DriverSqlite:
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
	case contractsorm.DriverSqlserver:
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

func createBookTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
		return `
CREATE TABLE books (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsorm.DriverSqlite:
		return `
CREATE TABLE books (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsorm.DriverSqlserver:
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

func createAuthorTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
		return `
CREATE TABLE authors (
  id SERIAL PRIMARY KEY NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsorm.DriverSqlite:
		return `
CREATE TABLE authors (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsorm.DriverSqlserver:
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

func createRoleTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
		return `
CREATE TABLE roles (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsorm.DriverSqlite:
		return `
CREATE TABLE roles (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsorm.DriverSqlserver:
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

func createHouseTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
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
	case contractsorm.DriverSqlite:
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
	case contractsorm.DriverSqlserver:
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

func createPhoneTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
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
	case contractsorm.DriverPostgresql:
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
	case contractsorm.DriverSqlite:
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
	case contractsorm.DriverSqlserver:
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

func createRoleUserTable(driver contractsorm.Driver) string {
	switch driver {
	case contractsorm.DriverMysql:
		return `
CREATE TABLE role_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint(20) unsigned NOT NULL,
  user_id bigint(20) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsorm.DriverPostgresql:
		return `
CREATE TABLE role_user (
  id SERIAL PRIMARY KEY NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsorm.DriverSqlite:
		return `
CREATE TABLE role_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsorm.DriverSqlserver:
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
