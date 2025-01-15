package tests

import (
	"context"

	"github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	databaseorm "github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/database/schema"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/postgres"
)

const (
	testDatabase = "goravel"
	testUsername = "goravel"
	testPassword = "Framework!123"
	testSchema   = "goravel"
)

func postgresTestQuery(prefix string, singular bool) *TestQuery {
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
	mockConfig.EXPECT().Get("database.connections.postgres.via").Return(func() (driver.Driver, error) {
		return nil, nil
	})

	mockConfig.EXPECT().Add("database.connections.postgres.schema", testSchema)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	postgresDriver := postgres.NewPostgres(mockConfig, utils.NewTestLog(), nil, "postgres")
	testQuery, err := NewTestQuery(ctx, postgresDriver, mockConfig)
	if err != nil {
		panic(err)
	}

	testQuery.CreateTable()

	return testQuery
}

func newSchema(testQuery *TestQuery, connectionToTestQuery map[string]*TestQuery) *schema.Schema {
	queries := make(map[string]contractsorm.Query)
	for connection, testQuery := range connectionToTestQuery {
		queries[connection] = testQuery.Query()
	}

	log := utils.NewTestLog()
	orm := databaseorm.NewOrm(context.Background(), testQuery.Config(), testQuery.Driver().Config().Driver, testQuery.Query(), queries, log, nil, nil)
	// TODO Use a common method instead
	postgresDriver := postgres.NewPostgres(testQuery.Config(), log, orm, "postgres")

	return schema.NewSchema(testQuery.Config(), log, orm, postgresDriver, nil)
}
