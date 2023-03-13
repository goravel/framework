package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/file"
)


func TestMigrateResetCommand(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "mysql",
			setup: func() {
				mysqlPool, mysqlResource, mysqlQuery, err := gorm.MysqlDocker()
				assert.Nil(t, err)

				createMysqlMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = mysqlQuery.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = mysqlPool.Purge(mysqlResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "postgresql",
			setup: func() {
				postgresqlPool, postgresqlResource, postgresqlDB, err := gorm.PostgresqlDocker()
				assert.Nil(t, err)

				createPostgresqlMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = postgresqlDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = postgresqlPool.Purge(postgresqlResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "sqlserver",
			setup: func() {
				sqlserverPool, sqlserverResource, sqlserverDB, err := gorm.SqlserverDocker()
				assert.Nil(t, err)

				createSqlserverMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = sqlserverDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				err = sqlserverPool.Purge(sqlserverResource)
				assert.Nil(t, err)
			},
		},
		{
			name: "sqlite",
			setup: func() {
				_, _, sqliteDB, err := gorm.SqliteDocker("goravel")
				assert.Nil(t, err)

				createSqliteMigrations()
				migrateCommand := &MigrateCommand{}
				mockContext := &consolemocks.Context{}
				err = migrateCommand.Handle(mockContext)
				assert.Nil(t, err)

				var agent Agent
				err = sqliteDB.Where("name", "goravel").First(&agent)
				assert.Nil(t, err)
				assert.True(t, agent.ID > 0)

				removeMigrations()
				file.Remove("goravel")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(_ *testing.T) {
			test.setup()
		})
	}
}

