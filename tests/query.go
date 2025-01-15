package tests

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/testing/utils"
)

type TestQuery struct {
	config config.Config
	driver driver.Driver
	query  orm.Query
}

func NewTestQuery(ctx context.Context, driver driver.Driver, config config.Config) (*TestQuery, error) {
	db, err := driver.Gorm()
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery{
		config: config,
		driver: driver,
		query:  gorm.NewQuery(ctx, config, driver.Config(), db, utils.NewTestLog(), nil, nil),
	}

	return testQuery, nil
}

func (r *TestQuery) CreateTable(testTables ...TestTable) {
	driverName := database.Driver(r.driver.Config().Driver)

	for table, sql := range newTestTables(driverName).All() {
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

func (r *TestQuery) Query() orm.Query {
	return r.query
}

func (r *TestQuery) WithSchema(schema string) {
	if r.driver.Config().Driver != contractsdatabase.DriverPostgres.String() && r.driver.Config().Driver != contractsdatabase.DriverSqlserver.String() {
		panic(fmt.Sprintf("%s does not support schema", r.driver.Config().Driver))
	}

	if _, err := r.query.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		panic(fmt.Sprintf("create schema %s failed: %v", schema, err))
	}

	if r.driver.Config().Driver == contractsdatabase.DriverSqlserver.String() {
		return
	}

	r.config.Add("database.connections.postgres.schema", schema)
	query, err := gorm.BuildQuery(context.Background(), r.config, r.driver.Config().Driver, utils.NewTestLog(), nil)
	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", r.driver.Config().Driver, err))
	}

	r.query = query
}
