package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	configmock "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

func TestMigrateResetCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	var (
		mockConfig *configmock.Config
		query      ormcontract.Query
	)

	beforeEach := func() {
		mockConfig = &configmock.Config{}
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "mysql",
			setup: func() {
				var err error
				mysqlQuery, err := gorm.NewTestQuery(docker.Mysql())
				require.Nil(t, err)

				query = mysqlQuery.Query()
				mockConfig = mysqlQuery.MockConfig()
				createMysqlMigrations()
			},
		},
		{
			name: "postgres",
			setup: func() {
				var err error
				postgresQuery, err := gorm.NewTestQuery(docker.Postgres())
				require.NoError(t, err)

				query = postgresQuery.Query()
				mockConfig = postgresQuery.MockConfig()
				createPostgresMigrations()
			},
		},
		{
			name: "sqlserver",
			setup: func() {
				sqlserverQuery, err := gorm.NewTestQuery(docker.Sqlserver())
				require.NoError(t, err)

				query = sqlserverQuery.Query()
				mockConfig = sqlserverQuery.MockConfig()
				createSqlserverMigrations()
			},
		},
		{
			name: "sqlite",
			setup: func() {
				var err error
				sqliteQuery, err := gorm.NewTestQuery(docker.Sqlite())
				require.NoError(t, err)

				query = sqliteQuery.Query()
				mockConfig = sqliteQuery.MockConfig()
				createSqliteMigrations()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()

			mockContext := &consolemocks.Context{}

			migrateCommand := NewMigrateCommand(mockConfig)
			assert.Nil(t, migrateCommand.Handle(mockContext))

			migrateResetCommand := NewMigrateResetCommand(mockConfig)
			assert.Nil(t, migrateResetCommand.Handle(mockContext))

			var agent Agent
			err := query.Where("name", "goravel").FirstOrFail(&agent)
			assert.Error(t, err)

			removeMigrations()
		})
	}
}
