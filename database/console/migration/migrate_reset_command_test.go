package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/database/migration"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMigrateResetCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
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

		mockContext := consolemocks.NewContext(t)

		migrateCommand := NewMigrateCommand(migrator)
		require.NotNil(t, migrateCommand)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		migrateResetCommand := NewMigrateResetCommand(mockConfig)
		assert.Nil(t, migrateResetCommand.Handle(mockContext))

		var agent migration.Agent
		err = query.Where("name", "goravel").FirstOrFail(&agent)
		assert.Error(t, err)
	}

	defer assert.Nil(t, file.Remove("database"))
}
