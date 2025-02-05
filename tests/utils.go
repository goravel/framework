package tests

import (
	"context"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	databaseorm "github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/testing/utils"
)

const (
	testDatabase = "goravel"
	testUsername = "goravel"
	testPassword = "Framework!123"
	testSchema   = "goravel"
)

func newSchema(testQuery *TestQuery, connectionToTestQuery map[string]*TestQuery) *schema.Schema {
	queries := make(map[string]contractsorm.Query)
	for connection, testQuery := range connectionToTestQuery {
		queries[connection] = testQuery.Query()
	}

	log := utils.NewTestLog()
	orm := databaseorm.NewOrm(context.Background(), testQuery.Config(), testQuery.Driver().Config().Connection, testQuery.Driver().Config(), testQuery.Query(), queries, log, nil, nil)

	return schema.NewSchema(testQuery.Config(), log, orm, testQuery.Driver(), nil)
}
