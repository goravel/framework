package tests

import (
	"context"
	"runtime"
	"strings"

	contractsdatabase "github.com/goravel/framework/contracts/database"
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
	dbConfigs := make(map[string]contractsdatabase.Config)
	for connection, testQuery := range connectionToTestQuery {
		queries[connection] = testQuery.Query()
		dbConfigs[connection] = testQuery.Driver().Pool().Writers[0]
	}

	log := utils.NewTestLog()
	dbConfig := testQuery.Driver().Pool().Writers[0]
	orm := databaseorm.NewOrm(context.Background(), testQuery.Config(), dbConfig.Connection, dbConfig, testQuery.Query(), databaseorm.NewQueriesCache(queries, dbConfigs), log, nil, nil, nil)

	schema, err := schema.NewSchema(testQuery.Config(), log, orm, testQuery.Driver(), nil)
	if err != nil {
		log.Panic(err.Error())
	}

	return schema
}

// countGoroutines returns the number of goroutines whose stack trace contains the given function name.
func countGoroutines(function string) int {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, 2*len(buf))
	}

	var count int
	for _, goroutine := range strings.Split(string(buf), "\n\n") {
		if strings.Contains(goroutine, function) {
			count++
		}
	}

	return count
}
