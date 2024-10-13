package migration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/gorm"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

func TestMigrateRefreshCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()

		mockConfig := testQuery.MockConfig()
		mockConfig.EXPECT().GetString("database.migrations.table").Return("migrations")
		mockConfig.EXPECT().GetString("database.migrations.driver").Return(migration.DriverSql)
		mockConfig.EXPECT().GetString(fmt.Sprintf("database.connections.%s.charset", testQuery.Docker().Driver().String())).Return("utf8bm4")

		mockSchema := mocksmigration.NewSchema(t)
		createMigrations(driver)

		mockArtisan := mocksconsole.NewArtisan(t)
		mockContext := mocksconsole.NewContext(t)
		mockContext.On("Option", "step").Return("").Once()

		migrateCommand := NewMigrateCommand(mockConfig, mockSchema)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand without --seed flag
		mockContext.On("OptionBool", "seed").Return(false).Once()
		migrateRefreshCommand := NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		var agent Agent
		err := query.Where("name", "goravel").First(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		mockArtisan = mocksconsole.NewArtisan(t)
		mockContext = mocksconsole.NewContext(t)
		mockContext.On("Option", "step").Return("5").Once()
		mockSchema = mocksmigration.NewSchema(t)

		migrateCommand = NewMigrateCommand(mockConfig, mockSchema)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand with --seed flag and --seeder specified
		mockContext.On("OptionBool", "seed").Return(true).Once()
		mockContext.On("OptionSlice", "seeder").Return([]string{"UserSeeder"}).Once()
		mockArtisan.On("Call", "db:seed --seeder UserSeeder").Return(nil).Once()
		migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		mockArtisan = mocksconsole.NewArtisan(t)
		mockContext = mocksconsole.NewContext(t)

		// Test MigrateRefreshCommand with --seed flag and no --seeder specified
		mockContext.On("Option", "step").Return("").Once()
		mockContext.On("OptionBool", "seed").Return(true).Once()
		mockContext.On("OptionSlice", "seeder").Return([]string{}).Once()
		mockArtisan.On("Call", "db:seed").Return(nil).Once()
		migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		var agent1 Agent
		err = query.Where("name", "goravel").First(&agent1)
		assert.Nil(t, err)
		assert.True(t, agent1.ID > 0)
	}

	defer assert.Nil(t, file.Remove("database"))
}
