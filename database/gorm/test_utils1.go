package gorm

import (
	"context"
	"fmt"
	"slices"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/testing/utils"
)

type TestQuery1 struct {
	driver orm.Driver
	query  orm.Query
}

func NewTestQuery1(driver orm.Driver, config config.Config) (*TestQuery1, error) {
	db, err := driver.Gorm()
	if err != nil {
		return nil, err
	}

	testQuery := &TestQuery1{
		driver: driver,
		query:  NewQuery(context.Background(), config, driver.Config(), db, utils.NewTestLog(), nil, nil),
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

func (r *TestQuery1) Query() orm.Query {
	return r.query
}
