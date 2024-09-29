package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateFreshCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		createMigrations(driver)

		mockContext := mocksconsole.NewContext(t)
		mockArtisan := mocksconsole.NewArtisan(t)
		migrateCommand := NewMigrateCommand(mockConfig)
		assert.Nil(t, migrateCommand.Handle(mockContext))
		mockContext.On("OptionBool", "seed").Return(false).Once()
		migrateFreshCommand := NewMigrateFreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateFreshCommand.Handle(mockContext))

		var agent Agent
		err := query.Where("name", "goravel").First(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		// Test MigrateFreshCommand with --seed flag and seeders specified
		mockContext = &mocksconsole.Context{}
		mockArtisan = &mocksconsole.Artisan{}
		mockContext.On("OptionBool", "seed").Return(true).Once()
		mockContext.On("OptionSlice", "seeder").Return([]string{"MockSeeder"}).Once()
		mockArtisan.On("Call", "db:seed --seeder MockSeeder").Return(nil).Once()
		migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateFreshCommand.Handle(mockContext))

		var agent1 Agent
		err = query.Where("name", "goravel").First(&agent1)
		assert.Nil(t, err)
		assert.True(t, agent1.ID > 0)

		// Test MigrateFreshCommand with --seed flag and no seeders specified
		mockContext = &mocksconsole.Context{}
		mockArtisan = &mocksconsole.Artisan{}
		mockContext.On("OptionBool", "seed").Return(true).Once()
		mockContext.On("OptionSlice", "seeder").Return([]string{}).Once()
		mockArtisan.On("Call", "db:seed").Return(nil).Once()
		migrateFreshCommand = NewMigrateFreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateFreshCommand.Handle(mockContext))

		var agent2 Agent
		err = query.Where("name", "goravel").First(&agent2)
		assert.Nil(t, err)
		assert.True(t, agent2.ID > 0)
	}
}
