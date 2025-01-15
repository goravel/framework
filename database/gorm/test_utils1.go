package gorm

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/testing/utils"
)

type TestQuery1 struct {
	config config.Config
	driver orm.Driver
	query  orm.Query
}

func NewTestQuery1(ctx context.Context, driver orm.Driver, config config.Config) (*TestQuery1, error) {
	db, err := driver.Gorm()
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery1{
		config: config,
		driver: driver,
		query:  NewQuery(ctx, config, driver.Config(), db, utils.NewTestLog(), nil, nil),
	}

	return testQuery, nil
}

func (r *TestQuery1) CreateTable(testTables ...TestTable) {
	driverName := database.Driver(r.driver.Config().Driver)

	for table, sql := range newTestTables(driverName).All() {
		if (len(testTables) == 0 && table != TestTableSchema) || slices.Contains(testTables, table) {
			if _, err := r.query.Exec(sql()); err != nil {
				panic(fmt.Sprintf("create table %v failed: %v", table, err))
			}
		}
	}
}

func (r *TestQuery1) Config() config.Config {
	return r.config
}

func (r *TestQuery1) Driver() orm.Driver {
	return r.driver
}

func (r *TestQuery1) Query() orm.Query {
	return r.query
}

func (r *TestQuery1) WithSchema(schema string) {
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
	query, err := BuildQuery(context.Background(), r.config, r.driver.Config().Driver, utils.NewTestLog(), nil)
	if err != nil {
		panic(fmt.Sprintf("connect to %s failed: %v", r.driver.Config().Driver, err))
	}

	r.query = query
}
