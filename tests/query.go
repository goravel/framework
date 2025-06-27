package tests

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/orm"
	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	databasedb "github.com/goravel/framework/database/db"
	databasedriver "github.com/goravel/framework/database/driver"
	databasegorm "github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/str"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/mysql"
	mysqlcontracts "github.com/goravel/mysql/contracts"
	"github.com/goravel/postgres"
	postgrescontracts "github.com/goravel/postgres/contracts"
	"github.com/goravel/sqlite"
	sqlitecontracts "github.com/goravel/sqlite/contracts"
	"github.com/goravel/sqlserver"
	sqlservercontracts "github.com/goravel/sqlserver/contracts"
)

type TestQuery struct {
	config config.Config
	db     contractsdb.DB
	driver contractsdriver.Driver
	query  orm.Query
}

func NewTestQuery(ctx context.Context, driver contractsdriver.Driver, config config.Config, connection string) (*TestQuery, error) {
	pool := driver.Pool()
	logger := databasedb.NewLogger(config, utils.NewTestLog())
	gorm, err := databasedriver.BuildGorm(config, logger.ToGorm(), pool, connection)
	if err != nil {
		return nil, err
	}

	db, err := databasedb.NewDB(ctx, config, driver, logger, gorm)
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery{
		config: config,
		db:     db,
		driver: driver,
		query:  databasegorm.NewQuery(ctx, config, pool.Writers[0], gorm, driver.Grammar(), utils.NewTestLog(), nil, nil),
	}

	return testQuery, nil
}

func (r *TestQuery) CreateTable(testTables ...TestTable) {
	driverName := r.driver.Pool().Writers[0].Driver

	for table, callback := range newTestTables(driverName, r.Driver().Grammar()).All() {
		if (len(testTables) == 0 && table != TestTableSchema) || slices.Contains(testTables, table) {
			sqls, err := callback()
			if err != nil {
				panic(fmt.Sprintf("create table %v failed: %v", table, err))
			}

			for _, sql := range sqls {
				if _, err = r.query.Exec(sql); err != nil {
					panic(fmt.Sprintf("create table %v failed: %v", table, err))
				}
			}
		}
	}
}

func (r *TestQuery) Config() config.Config {
	return r.config
}

func (r *TestQuery) DB() contractsdb.DB {
	return r.db
}

func (r *TestQuery) Driver() contractsdriver.Driver {
	return r.driver
}

func (r *TestQuery) MockConfig() *mocksconfig.Config {
	return r.config.(*mocksconfig.Config)
}

func (r *TestQuery) Query() orm.Query {
	return r.query
}

func (r *TestQuery) WithSchema(schema string) {
	dbConfig := r.driver.Pool().Writers[0]
	if dbConfig.Driver != postgres.Name && dbConfig.Driver != sqlserver.Name {
		panic(fmt.Sprintf("%s does not support schema", dbConfig.Driver))
	}

	if _, err := r.query.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		panic(fmt.Sprintf("create schema %s failed: %v", schema, err))
	}

	if dbConfig.Driver == sqlserver.Name {
		return
	}

	r.MockConfig().EXPECT().Add(fmt.Sprintf("database.connections.%s.schema", dbConfig.Connection), schema)
	r.config.Add(fmt.Sprintf("database.connections.%s.schema", dbConfig.Connection), schema)

	query, _, err := databasegorm.BuildQuery(context.Background(), r.config, dbConfig.Connection, utils.NewTestLog(), nil)
	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", dbConfig.Connection, err))
	}

	r.query = query
}

type TestQueryBuilder struct {
}

func NewTestQueryBuilder() *TestQueryBuilder {
	return &TestQueryBuilder{}
}

func (r *TestQueryBuilder) All(prefix string, singular bool) map[string]*TestQuery {
	postgresTestQuery := r.Postgres(prefix, singular)
	mysqlTestQuery := r.Mysql(prefix, singular)
	sqlserverTestQuery := r.Sqlserver(prefix, singular)
	sqliteTestQuery := r.Sqlite(prefix, singular)

	return map[string]*TestQuery{
		postgresTestQuery.Driver().Pool().Writers[0].Driver:  postgresTestQuery,
		mysqlTestQuery.Driver().Pool().Writers[0].Driver:     mysqlTestQuery,
		sqlserverTestQuery.Driver().Pool().Writers[0].Driver: sqlserverTestQuery,
		sqliteTestQuery.Driver().Pool().Writers[0].Driver:    sqliteTestQuery,
	}
}

func (r *TestQueryBuilder) AllWithTimezone(timezone string) map[string]*TestQuery {
	postgresTestQuery := r.PostgresWithTimezone(timezone)
	mysqlTestQuery := r.MysqlWithTimezone(timezone)
	sqlserverTestQuery := r.SqlserverWithTimezone(timezone)
	sqliteTestQuery := r.SqliteWithTimezone(timezone)

	return map[string]*TestQuery{
		postgresTestQuery.Driver().Pool().Writers[0].Driver:  postgresTestQuery,
		mysqlTestQuery.Driver().Pool().Writers[0].Driver:     mysqlTestQuery,
		sqlserverTestQuery.Driver().Pool().Writers[0].Driver: sqlserverTestQuery,
		sqliteTestQuery.Driver().Pool().Writers[0].Driver:    sqliteTestQuery,
	}
}

func (r *TestQueryBuilder) AllWithReadWrite() map[string]map[string]*TestQuery {
	return map[string]map[string]*TestQuery{
		postgres.Name:  r.PostgresWithReadWrite(),
		mysql.Name:     r.MysqlWithReadWrite(),
		sqlserver.Name: r.SqlserverWithReadWrite(),
		sqlite.Name:    r.SqliteWithReadWrite(),
	}
}

func (r *TestQueryBuilder) Mysql(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(mysql.Name, prefix, "UTC", singular)
	return testQuery
}

func (r *TestQueryBuilder) MysqlWithTimezone(timezone string) *TestQuery {
	testQuery, _ := r.single(mysql.Name, "", timezone, false)
	return testQuery
}

func (r *TestQueryBuilder) MysqlWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(mysql.Name)
}

func (r *TestQueryBuilder) Postgres(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(postgres.Name, prefix, "UTC", singular)
	return testQuery
}

func (r *TestQueryBuilder) PostgresWithTimezone(timezone string) *TestQuery {
	testQuery, _ := r.single(postgres.Name, "", timezone, false)
	return testQuery
}

func (r *TestQueryBuilder) PostgresWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(postgres.Name)
}

func (r *TestQueryBuilder) Sqlite(prefix string, singular bool) *TestQuery {
	connection := sqlite.Name + "_" + str.Random(6)
	mockConfig := &mocksconfig.Config{}
	docker := sqlite.NewDocker(fmt.Sprintf("%s_%s", testDatabase, str.Random(6)))
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:     sqlite.Name,
		Database:   docker.Config().Database,
		Connection: connection,
		Prefix:     prefix,
		Singular:   singular,
	})

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := sqlite.NewSqlite(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig, connection)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func (r *TestQueryBuilder) SqliteWithTimezone(timezone string) *TestQuery {
	connection := sqlite.Name
	mockConfig := &mocksconfig.Config{}
	docker := sqlite.NewDocker(fmt.Sprintf("%s_%s", testDatabase, str.Random(6)))
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:     sqlite.Name,
		Database:   docker.Config().Database,
		Connection: connection,
		Timezone:   timezone,
	})

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := sqlite.NewSqlite(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig, connection)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func (r *TestQueryBuilder) SqliteWithReadWrite() map[string]*TestQuery {
	writeTestQuery := r.Sqlite("", false)
	readTestQuery := r.Sqlite("", false)

	return map[string]*TestQuery{
		"write": writeTestQuery,
		"read":  readTestQuery,
		"mix": r.mix(sqlite.Name, contractsdocker.DatabaseConfig{
			Database: writeTestQuery.Driver().Pool().Writers[0].Database,
		}, contractsdocker.DatabaseConfig{
			Database: readTestQuery.Driver().Pool().Writers[0].Database,
		}),
	}
}

func (r *TestQueryBuilder) Sqlserver(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(sqlserver.Name, prefix, "UTC", singular)
	return testQuery
}

func (r *TestQueryBuilder) SqlserverWithTimezone(timezone string) *TestQuery {
	testQuery, _ := r.single(sqlserver.Name, "", timezone, false)
	return testQuery
}

func (r *TestQueryBuilder) SqlserverWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(sqlserver.Name)
}

func (r *TestQueryBuilder) single(driver, prefix, timezone string, singular bool) (*TestQuery, contractsdocker.DatabaseDriver) {
	var (
		dockerDriver   contractsdocker.DatabaseDriver
		databaseDriver contractsdriver.Driver

		connection = driver + "_" + str.Random(6)
		mockConfig = &mocksconfig.Config{}
	)

	switch driver {
	case postgres.Name:
		dockerDriver = postgres.NewDocker(postgres.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
		databaseDriver = postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection)
	case mysql.Name:
		dockerDriver = mysql.NewDocker(mysql.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
		databaseDriver = mysql.NewMysql(mockConfig, utils.NewTestLog(), connection)
	case sqlserver.Name:
		dockerDriver = sqlserver.NewDocker(sqlserver.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
		databaseDriver = sqlserver.NewSqlserver(mockConfig, utils.NewTestLog(), connection)
	}

	container := NewContainer(dockerDriver)
	containerInstance, err := container.Build()
	if err != nil {
		panic(err)
	}

	// goravel/*:docker.go#resetConfigPort
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return(nil)
	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.port", connection), containerInstance.Config().Port)

	if err := containerInstance.Ready(); err != nil {
		panic(err)
	}

	mockDatabaseConfigWithoutWriteAndRead(mockConfig, database.Config{
		Driver:     driver,
		Host:       containerInstance.Config().Host,
		Port:       containerInstance.Config().Port,
		Username:   containerInstance.Config().Username,
		Password:   containerInstance.Config().Password,
		Database:   containerInstance.Config().Database,
		Connection: connection,
		Prefix:     prefix,
		Singular:   singular,
		Timezone:   timezone,
	})

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	testQuery, err := NewTestQuery(ctx, databaseDriver, mockConfig, connection)
	if err != nil {
		panic(err)
	}

	return testQuery, containerInstance
}

func (r *TestQueryBuilder) readWriteMix(driver string) map[string]*TestQuery {
	writeTestQuery, writeDatabaseDriver := r.single(driver, "", "UTC", false)
	readTestQuery, readDatabaseDriver := r.single(driver, "", "UTC", false)

	return map[string]*TestQuery{
		"write": writeTestQuery,
		"read":  readTestQuery,
		"mix":   r.mix(driver, writeDatabaseDriver.Config(), readDatabaseDriver.Config()),
	}
}

func (r *TestQueryBuilder) mix(driver string, writeDatabaseConfig, readDatabaseConfig contractsdocker.DatabaseConfig) *TestQuery {
	var (
		databaseDriver contractsdriver.Driver

		connection = driver + "_" + str.Random(6)
		mockConfig = &mocksconfig.Config{}
	)

	switch driver {
	case postgres.Name:
		databaseDriver = postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection)
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return([]postgrescontracts.Config{
			{
				Host:     writeDatabaseConfig.Host,
				Port:     writeDatabaseConfig.Port,
				Username: writeDatabaseConfig.Username,
				Password: writeDatabaseConfig.Password,
				Database: writeDatabaseConfig.Database,
			},
		})
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return([]postgrescontracts.Config{
			{
				Host:     readDatabaseConfig.Host,
				Port:     readDatabaseConfig.Port,
				Username: readDatabaseConfig.Username,
				Password: readDatabaseConfig.Password,
				Database: readDatabaseConfig.Database,
			},
		})

	case mysql.Name:
		databaseDriver = mysql.NewMysql(mockConfig, utils.NewTestLog(), connection)
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return([]mysqlcontracts.Config{
			{
				Host:     writeDatabaseConfig.Host,
				Port:     writeDatabaseConfig.Port,
				Username: writeDatabaseConfig.Username,
				Password: writeDatabaseConfig.Password,
				Database: writeDatabaseConfig.Database,
			},
		})
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return([]mysqlcontracts.Config{
			{
				Host:     readDatabaseConfig.Host,
				Port:     readDatabaseConfig.Port,
				Username: readDatabaseConfig.Username,
				Password: readDatabaseConfig.Password,
				Database: readDatabaseConfig.Database,
			},
		})
	case sqlserver.Name:
		databaseDriver = sqlserver.NewSqlserver(mockConfig, utils.NewTestLog(), connection)
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return([]sqlservercontracts.Config{
			{
				Host:     writeDatabaseConfig.Host,
				Port:     writeDatabaseConfig.Port,
				Username: writeDatabaseConfig.Username,
				Password: writeDatabaseConfig.Password,
				Database: writeDatabaseConfig.Database,
			},
		})
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return([]sqlservercontracts.Config{
			{
				Host:     readDatabaseConfig.Host,
				Port:     readDatabaseConfig.Port,
				Username: readDatabaseConfig.Username,
				Password: readDatabaseConfig.Password,
				Database: readDatabaseConfig.Database,
			},
		})
	case sqlite.Name:
		databaseDriver = sqlite.NewSqlite(mockConfig, utils.NewTestLog(), connection)
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return([]sqlitecontracts.Config{
			{
				Database: writeDatabaseConfig.Database,
			},
		})
		mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return([]sqlitecontracts.Config{
			{
				Database: readDatabaseConfig.Database,
			},
		})
	}

	mockDatabaseConfigWithoutWriteAndRead(mockConfig, database.Config{
		Driver:     driver,
		Connection: connection,
		Prefix:     "",
		Singular:   false,
		Timezone:   "UTC",
	})

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	testQuery, err := NewTestQuery(ctx, databaseDriver, mockConfig, connection)
	if err != nil {
		panic(err)
	}

	return testQuery
}
