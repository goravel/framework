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
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/logger"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/docker"
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
	"github.com/jmoiron/sqlx"
)

type TestQuery struct {
	config config.Config
	db     contractsdb.DB
	driver contractsdriver.Driver
	query  orm.Query
}

func NewTestQuery(ctx context.Context, driver contractsdriver.Driver, config config.Config) (*TestQuery, error) {
	query, err := driver.Gorm()
	if err != nil {
		return nil, err
	}

	db, err := driver.DB()
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery{
		config: config,
		db:     databasedb.NewDB(ctx, config, driver, logger.NewLogger(config, utils.NewTestLog()), sqlx.NewDb(db, driver.Config().Driver)),
		driver: driver,
		query:  gorm.NewQuery(ctx, config, driver.Config(), query, driver.Grammar(), utils.NewTestLog(), nil, nil),
	}

	return testQuery, nil
}

func (r *TestQuery) CreateTable(testTables ...TestTable) {
	driverName := r.driver.Config().Driver

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
	if r.driver.Config().Driver != postgres.Name && r.driver.Config().Driver != sqlserver.Name {
		panic(fmt.Sprintf("%s does not support schema", r.driver.Config().Driver))
	}

	if _, err := r.query.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		panic(fmt.Sprintf("create schema %s failed: %v", schema, err))
	}

	if r.driver.Config().Driver == sqlserver.Name {
		return
	}

	r.MockConfig().EXPECT().Add(fmt.Sprintf("database.connections.%s.schema", r.driver.Config().Connection), schema)
	r.config.Add(fmt.Sprintf("database.connections.%s.schema", r.driver.Config().Driver), schema)

	query, _, err := gorm.BuildQuery(context.Background(), r.config, r.driver.Config().Driver, utils.NewTestLog(), nil)
	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", r.driver.Config().Driver, err))
	}

	r.query = query
}

type TestQueryBuilder struct {
}

func NewTestQueryBuilder() *TestQueryBuilder {
	return &TestQueryBuilder{}
}

func (r *TestQueryBuilder) All(prefix string, singular bool) map[string]*TestQuery {
	// postgresTestQuery := r.Postgres(prefix, singular)
	mysqlTestQuery := r.Mysql(prefix, singular)
	// sqlserverTestQuery := r.Sqlserver(prefix, singular)
	// sqliteTestQuery := r.Sqlite(prefix, singular)

	return map[string]*TestQuery{
		// postgresTestQuery.Driver().Config().Driver: postgresTestQuery,
		mysqlTestQuery.Driver().Config().Driver: mysqlTestQuery,
		// sqlserverTestQuery.Driver().Config().Driver: sqlserverTestQuery,
		// sqliteTestQuery.Driver().Config().Driver: sqliteTestQuery,
	}
}

func (r *TestQueryBuilder) AllOfReadWrite() map[string]map[string]*TestQuery {
	return map[string]map[string]*TestQuery{
		postgres.Name:  r.PostgresWithReadWrite(),
		mysql.Name:     r.MysqlWithReadWrite(),
		sqlserver.Name: r.SqlserverWithReadWrite(),
		sqlite.Name:    r.SqliteWithReadWrite(),
	}
}

func (r *TestQueryBuilder) Mysql(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(mysql.Name, prefix, singular)
	return testQuery
}

func (r *TestQueryBuilder) MysqlWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(mysql.Name)
}

func (r *TestQueryBuilder) Postgres(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(postgres.Name, prefix, singular)
	return testQuery
}

func (r *TestQueryBuilder) PostgresWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(postgres.Name)
}

func (r *TestQueryBuilder) Sqlite(prefix string, singular bool) *TestQuery {
	connection := sqlite.Name
	mockConfig := &mocksconfig.Config{}
	docker := sqlite.NewDocker(fmt.Sprintf("%s_%s", testDatabase, str.Random(6)))
	err := docker.Build()
	if err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:   sqlite.Name,
		Database: docker.Config().Database,
	}, connection, prefix, singular)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := sqlite.NewSqlite(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig)
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
			Database: writeTestQuery.Driver().Config().Database,
		}, contractsdocker.DatabaseConfig{
			Database: readTestQuery.Driver().Config().Database,
		}),
	}
}

func (r *TestQueryBuilder) Sqlserver(prefix string, singular bool) *TestQuery {
	testQuery, _ := r.single(sqlserver.Name, prefix, singular)
	return testQuery
}

func (r *TestQueryBuilder) SqlserverWithReadWrite() map[string]*TestQuery {
	return r.readWriteMix(sqlserver.Name)
}

func (r *TestQueryBuilder) single(driver string, prefix string, singular bool) (*TestQuery, contractsdocker.DatabaseDriver) {
	var (
		dockerDriver   contractsdocker.DatabaseDriver
		databaseDriver contractsdriver.Driver

		connection = driver
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

	container := docker.NewContainer(dockerDriver)
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
		Driver:   driver,
		Host:     containerInstance.Config().Host,
		Port:     containerInstance.Config().Port,
		Username: containerInstance.Config().Username,
		Password: containerInstance.Config().Password,
		Database: containerInstance.Config().Database,
	}, connection, prefix, singular)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	testQuery, err := NewTestQuery(ctx, databaseDriver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery, containerInstance
}

func (r *TestQueryBuilder) readWriteMix(driver string) map[string]*TestQuery {
	writeTestQuery, writeDatabaseDriver := r.single(driver, "", false)
	readTestQuery, readDatabaseDriver := r.single(driver, "", false)

	return map[string]*TestQuery{
		"write": writeTestQuery,
		"read":  readTestQuery,
		"mix":   r.mix(driver, writeDatabaseDriver.Config(), readDatabaseDriver.Config()),
	}
}

func (r *TestQueryBuilder) mix(driver string, writeDatabaseConfig, readDatabaseConfig contractsdocker.DatabaseConfig) *TestQuery {
	var (
		databaseDriver contractsdriver.Driver

		connection = postgres.Name
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
		Driver: driver,
	}, connection, "", false)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	testQuery, err := NewTestQuery(ctx, databaseDriver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery
}
