package tests

import (
	"context"
	"fmt"

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
	connection := "postgres"
	mockConfig := &mocksconfig.Config{}
	postgresImage := postgres.NewDocker(postgres.NewConfig(mockConfig, connection), testDatabase, testUsername, testPassword)
	container := docker.NewContainer(postgresImage)
	postgresContainer, err := container.Build()
	if err != nil {
		panic(err)
	}

	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.write", connection)).Return(nil)
	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.port", connection), postgresContainer.Config().Port)

	if err := postgresContainer.Ready(); err != nil {
		panic(err)
	}

	mockConfig.EXPECT().GetBool("app.debug").Return(true)
	mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(200)
	mockConfig.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10)
	mockConfig.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600)
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.read", connection)).Return(nil)

	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.host", connection)).Return(postgresContainer.Config().Host)
	mockConfig.EXPECT().GetInt(fmt.Sprintf("database.connections.%s.port", connection)).Return(postgresContainer.Config().Port)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.username", connection)).Return(postgresContainer.Config().Username)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.password", connection)).Return(postgresContainer.Config().Password)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.database", connection)).Return(postgresContainer.Config().Database)
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.sslmode", connection)).Return("disable")
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.timezone", connection)).Return("UTC")
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.prefix", connection)).Return(prefix)
	mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.singular", connection)).Return(singular)
	mockConfig.EXPECT().GetBool(fmt.Sprintf("database.connections.%s.no_lower_case", connection)).Return(false)
	mockConfig.EXPECT().GetString("database.connections.postgres.dsn").Return("")
	mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.schema", connection), "public").Return("public")
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.name_replacer", connection)).Return(nil)
	mockConfig.EXPECT().Get(fmt.Sprintf("database.connections.%s.via", connection)).Return(func() (driver.Driver, error) {
		return postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection), nil
	})

	mockConfig.EXPECT().Add(fmt.Sprintf("database.connections.%s.schema", connection), testSchema)

	ctx := context.WithValue(context.Background(), testContextKey, "goravel")
	postgresDriver := postgres.NewPostgres(mockConfig, utils.NewTestLog(), connection)
	testQuery, err := NewTestQuery(ctx, postgresDriver, mockConfig)
	if err != nil {
		panic(err)
	}

	return testQuery
}

func newSchema(testQuery *TestQuery, connectionToTestQuery map[string]*TestQuery) *schema.Schema {
	queries := make(map[string]contractsorm.Query)
	for connection, testQuery := range connectionToTestQuery {
		queries[connection] = testQuery.Query()
	}

	log := utils.NewTestLog()
	orm := databaseorm.NewOrm(context.Background(), testQuery.Config(), testQuery.Driver().Config().Connection, testQuery.Driver().Config(), testQuery.Query(), queries, log, nil, nil)
	// TODO Use a common method instead
	postgresDriver := postgres.NewPostgres(testQuery.Config(), log, "postgres")

	return schema.NewSchema(testQuery.Config(), log, orm, postgresDriver, nil)
}
