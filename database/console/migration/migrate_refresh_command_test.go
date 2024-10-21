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

func TestMigrateRefreshCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests that use Docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()

		mockConfig := testQuery.MockConfig()
		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations")
		mockConfig.EXPECT().GetString("database.migrations.driver").Return(contractsmigration.MigratorSql)
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", testQuery.Docker().Driver().String())).Return("utf8bm4")

		migration.CreateTestMigrations(driver)

		migrator, err := migration.NewSqlMigrator(mockConfig)
		require.NoError(t, err)

		mockArtisan := mocksconsole.NewArtisan(t)
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("step").Return("").Once()
		mockContext.EXPECT().Info("Migration success").Once()

		migrateCommand := NewMigrateCommand(migrator)
		require.NotNil(t, migrateCommand)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand without --seed flag
		mockContext.EXPECT().OptionBool("seed").Return(false).Once()
		mockContext.EXPECT().Info("Migration refresh success").Once()

		migrateRefreshCommand := NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		var agent migration.Agent
		err = query.Where("name", "goravel").First(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		mockArtisan = mocksconsole.NewArtisan(t)
		mockContext = mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("step").Return("5").Once()
		mockContext.EXPECT().Info("Migration success").Once()

		migrateCommand = NewMigrateCommand(migrator)
		require.NotNil(t, migrateCommand)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand with --seed flag and --seeder specified
		mockContext.EXPECT().OptionBool("seed").Return(true).Once()
		mockContext.EXPECT().OptionSlice("seeder").Return([]string{"UserSeeder"}).Once()
		mockContext.EXPECT().Info("Migration refresh success").Once()
		mockArtisan.EXPECT().Call("db:seed --seeder UserSeeder").Once()

		migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		mockArtisan = mocksconsole.NewArtisan(t)
		mockContext = mocksconsole.NewContext(t)

		// Test MigrateRefreshCommand with --seed flag and no --seeder specified
		mockContext.EXPECT().Option("step").Return("").Once()
		mockContext.EXPECT().OptionBool("seed").Return(true).Once()
		mockContext.EXPECT().OptionSlice("seeder").Return([]string{}).Once()
		mockContext.EXPECT().Info("Migration refresh success").Once()
		mockArtisan.EXPECT().Call("db:seed").Once()

		migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		var agent1 migration.Agent
		err = query.Where("name", "goravel").First(&agent1)
		assert.Nil(t, err)
		assert.True(t, agent1.ID > 0)
	}

	defer assert.Nil(t, file.Remove("database"))
}
