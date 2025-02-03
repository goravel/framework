package tests

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/postgres"
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

func (r *TestQuery) Query() orm.Query {
	return r.query
}

func (r *TestQuery) WithSchema(schema string) {
	// TODO: Add Sqlserver
	if r.driver.Config().Driver != postgres.Name {
		panic(fmt.Sprintf("%s does not support schema", r.driver.Config().Driver))
	}

	if _, err := r.query.Exec(fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		panic(fmt.Sprintf("create schema %s failed: %v", schema, err))
	}

	// TODO Replace with sqlserver.Name
	if r.driver.Config().Driver == "sqlserver" {
		return
	}

	r.config.Add("database.connections.postgres.schema", schema)
	query, _, err := gorm.BuildQuery(context.Background(), r.config, r.driver.Config().Driver, utils.NewTestLog(), nil)
	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", r.driver.Config().Driver, err))
	}

	r.query = query
}
