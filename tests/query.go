package tests

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/str"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
)

type TestQuery struct {
	config config.Config
	driver driver.Driver
	query  orm.Query
}

func NewTestQuery(ctx context.Context, driver driver.Driver, config config.Config) (*TestQuery, error) {
	db, gormQuery, err := driver.Gorm()
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery{
		config: config,
		driver: driver,
		query:  gorm.NewQuery(ctx, config, driver.Config(), db, gormQuery, utils.NewTestLog(), nil, nil),
	}

	return testQuery, nil
}

func (r *TestQuery) CreateTable(testTables ...TestTable) {
	driverName := r.driver.Config().Driver

	for table, sql := range newTestTables(driverName, r.Driver().Grammar()).All() {
		if (len(testTables) == 0 && table != TestTableSchema) || slices.Contains(testTables, table) {
			if _, err := r.query.Exec(sql()); err != nil {
				panic(fmt.Sprintf("create table %v failed: %v", table, err))
			}
		}
	}
}

func (r *TestQuery) Config() config.Config {
	return r.config
}

func (r *TestQuery) Driver() driver.Driver {
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

func postgresTestQuery(prefix string, singular bool) *TestQuery {
	connection := postgres.Name
	mockConfig := &mocksconfig.Config{}
	image := postgres.NewDocker(postgres.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
	container := docker.NewContainer(image)
	containerInstance, err := container.Build()
	if err != nil {
		panic(err)
	}

	// goravel/postgres:docker.go#resetConfigPort
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.port", connection), containerInstance.Config().Port)

	if err := containerInstance.Ready(); err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:   postgres.Name,
		Host:     containerInstance.Config().Host,
		Port:     containerInstance.Config().Port,
		Username: containerInstance.Config().Username,
		Password: containerInstance.Config().Password,
		Database: containerInstance.Config().Database,
	}, connection, prefix, singular)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func mysqlTestQuery(prefix string, singular bool) *TestQuery {
	connection := mysql.Name
	mockConfig := &mocksconfig.Config{}
	image := mysql.NewDocker(mysql.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
	container := docker.NewContainer(image)
	containerInstance, err := container.Build()
	if err != nil {
		panic(err)
	}

	// goravel/mysql:docker.go#resetConfigPort
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.port", connection), containerInstance.Config().Port)

	if err := containerInstance.Ready(); err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:   mysql.Name,
		Host:     containerInstance.Config().Host,
		Port:     containerInstance.Config().Port,
		Username: containerInstance.Config().Username,
		Password: containerInstance.Config().Password,
		Database: containerInstance.Config().Database,
	}, connection, prefix, singular)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := mysql.NewMysql(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func sqlserverTestQuery(prefix string, singular bool) *TestQuery {
	connection := sqlserver.Name
	mockConfig := &mocksconfig.Config{}
	image := sqlserver.NewDocker(sqlserver.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
	container := docker.NewContainer(image)
	containerInstance, err := container.Build()
	if err != nil {
		panic(err)
	}

	// goravel/sqlserver:docker.go#resetConfigPort
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.port", connection), containerInstance.Config().Port)

	if err := containerInstance.Ready(); err != nil {
		panic(err)
	}

	mockDatabaseConfig(mockConfig, database.Config{
		Driver:   sqlserver.Name,
		Host:     containerInstance.Config().Host,
		Port:     containerInstance.Config().Port,
		Username: containerInstance.Config().Username,
		Password: containerInstance.Config().Password,
		Database: containerInstance.Config().Database,
	}, connection, prefix, singular)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	driver := sqlserver.NewSqlserver(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, driver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func sqliteTestQuery(prefix string, singular bool) *TestQuery {
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
