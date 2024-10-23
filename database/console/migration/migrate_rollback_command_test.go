package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/migration"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMigrateRollbackCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests that use Docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()

		mockConfig := testQuery.MockConfig()
		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations").Once()
		mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.MigratorSql).Once()
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", testQuery.Docker().Driver().String())).Return("utf8bm4").Once()

		migration.CreateTestMigrations(driver)

		migrator, err := migration.NewSqlMigrator(mockConfig)
		require.NoError(t, err)

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("step").Return("1").Once()
		mockContext.EXPECT().Info("Migration success").Once()

		migrateCommand := NewMigrateCommand(migrator)
		require.NotNil(t, migrateCommand)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		var agent migration.Agent
		err = query.Where("name", "goravel").FirstOrFail(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		mockContext.EXPECT().Info("Migration rollback success").Once()

		migrateRollbackCommand := NewMigrateRollbackCommand(mockConfig)
		assert.Nil(t, migrateRollbackCommand.Handle(mockContext))

		var agent1 migration.Agent
		err = query.Where("name", "goravel").FirstOrFail(&agent1)
		assert.Error(t, err)
	}

	defer assert.Nil(t, file.Remove("database"))
}
