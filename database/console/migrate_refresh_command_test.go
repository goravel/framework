package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/gorm"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/env"
)

func TestMigrateRefreshCommand(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	testQueries := gorm.NewTestQueries().Queries()
	for driver, testQuery := range testQueries {
		query := testQuery.Query()
		mockConfig := testQuery.MockConfig()
		createMigrations(driver)

		mockArtisan := mocksconsole.NewArtisan(t)
		mockContext := mocksconsole.NewContext(t)
		mockContext.On("Option", "step").Return("").Once()

		migrateCommand := NewMigrateCommand(mockConfig)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand without --seed flag
		mockContext.On("OptionBool", "seed").Return(false).Once()
		migrateRefreshCommand := NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		var agent Agent
		err := query.Where("name", "goravel").First(&agent)
		assert.Nil(t, err)
		assert.True(t, agent.ID > 0)

		mockArtisan = &mocksconsole.Artisan{}
		mockContext = &mocksconsole.Context{}
		mockContext.On("Option", "step").Return("5").Once()

		migrateCommand = NewMigrateCommand(mockConfig)
		assert.Nil(t, migrateCommand.Handle(mockContext))

		// Test MigrateRefreshCommand with --seed flag and --seeder specified
		mockContext.On("OptionBool", "seed").Return(true).Once()
		mockContext.On("OptionSlice", "seeder").Return([]string{"UserSeeder"}).Once()
		mockArtisan.On("Call", "db:seed --seeder UserSeeder").Return(nil).Once()
		migrateRefreshCommand = NewMigrateRefreshCommand(mockConfig, mockArtisan)
		assert.Nil(t, migrateRefreshCommand.Handle(mockContext))

		mockArtisan = &mocksconsole.Artisan{}
		mockContext = &mocksconsole.Context{}

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
}
