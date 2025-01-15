package tests

import (
	"context"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	databaseorm "github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/database/schema"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/postgres"
	"github.com/stretchr/testify/mock"
)

const (
	testDatabase = "goravel"
	testUsername = "goravel"
	testPassword = "Framework!123"
	testSchema   = "goravel"
)

func postgresTestQuery(prefix string, singular bool) *gorm.TestQuery1 {
	postgresImage := postgres.NewDocker(testDatabase, testUsername, testPassword)
	builder := docker.NewBuilder(postgresImage)
	postgresContainer, err := builder.Build()
	if err != nil {
		panic(err)
	}
	if err := postgresContainer.Ready(); err != nil {
		panic(err)
	}

	mockConfig := &mocksconfig.Config{}

	mockConfig.EXPECT().GetBool("app.debug").Return(true)
	mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(200)
	mockConfig.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10)
	mockConfig.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600)

	mockConfig.EXPECT().Get("database.connections.postgres.read").Return(nil)
	mockConfig.EXPECT().Get("database.connections.postgres.write").Return(nil)
	mockConfig.EXPECT().GetString("database.connections.postgres.host").Return("localhost")
	mockConfig.EXPECT().GetInt("database.connections.postgres.port").Return(postgresContainer.Config().Port)
	mockConfig.EXPECT().GetString("database.connections.postgres.username").Return(testUsername)
	mockConfig.EXPECT().GetString("database.connections.postgres.password").Return(testPassword)
	mockConfig.EXPECT().GetString("database.connections.postgres.database").Return(postgresContainer.Config().Database)
	mockConfig.EXPECT().GetString("database.connections.postgres.sslmode").Return("disable")
	mockConfig.EXPECT().GetString("database.connections.postgres.timezone").Return("UTC")
	mockConfig.EXPECT().GetString("database.connections.postgres.prefix").Return(prefix)
	mockConfig.EXPECT().GetBool("database.connections.postgres.singular").Return(singular)
	mockConfig.EXPECT().GetBool("database.connections.postgres.no_lower_case").Return(false)
	mockConfig.EXPECT().GetString("database.connections.postgres.dsn").Return("")
	mockConfig.EXPECT().GetString("database.connections.postgres.schema", "public").Return("public")
	mockConfig.EXPECT().Get("database.connections.postgres.name_replacer").Return(nil)

	mockConfig.EXPECT().Add("database.connections.postgres.schema", testSchema)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	postgresDriver := postgres.NewPostgres(mockConfig, utils.NewTestLog(), "postgres")
	testQuery, err := gorm.NewTestQuery1(ctx, postgresDriver, mockConfig)
	if err != nil {
		panic(err)
	}

	testQuery.CreateTable()

	return testQuery
}

func newSchema(testQuery *gorm.TestQuery1, connectionToTestQuery map[string]*gorm.TestQuery1) *schema.Schema {
	queries := make(map[string]contractsorm.Query)
	for connection, testQuery := range connectionToTestQuery {
		queries[connection] = testQuery.Query()
	}

	mockLog := &mockslog.Log{}
	mockLog.EXPECT().Errorf(mock.Anything).Maybe()
	orm := databaseorm.NewOrm(context.Background(), testQuery.Config(), testQuery.Driver().Config().Driver, testQuery.Query(), queries, mockLog, nil, nil)
	schema := schema.NewSchema(testQuery.Config(), mockLog, orm, nil)

	return schema
}
