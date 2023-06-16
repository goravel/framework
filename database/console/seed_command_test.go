package console

import (
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"

	// appmocks "github.com/goravel/framework/contracts/foundation/mocks"

	"github.com/stretchr/testify/assert"
)

func TestSeedCommand_Handle(t *testing.T) {
	// // Create mock config
	// mockConfig := &configmocks.Config{}
	// mockConfig.On("Env", "APP_ENV").Return("development")

	// // Create mock seeder facade

	// mockApp := &appmocks.Application{}
	// mockSeeder := mockApp.MakeSeeder()
	// mockSeeder.On("GetSeeders").Return([]seeder.Seeder{
	// 	&mocks.Seeder{},
	// }, nil)

	// // Create command instance
	// command := NewSeedCommand(mockConfig, mockSeeder)

	// // Create mock context
	// mockContext := &consolemocks.Context{}
	// mockContext.On("OptionSlice", "seeder").Return([]string{})

	// // Test Handle method with no seeders specified
	// err := command.Handle(mockContext)
	// assert.NoError(t, err)
	// mockSeeder.AssertCalled(t, "GetSeeders")

	// // Create mock context with seeders specified
	// mockContext.On("OptionSlice", "seeder").Return([]string{"Seeder1", "Seeder2"})
	// mockSeeder.On("GetSeeder", "seeders.Seeder1").Return(&mocks.Seeder{}, nil)
	// mockSeeder.On("GetSeeder", "seeders.Seeder2").Return(&mocks.Seeder{}, nil)

	// // Test Handle method with seeders specified
	// err = command.Handle(mockContext)
	// assert.NoError(t, err)
	// mockSeeder.AssertCalled(t, "GetSeeder", "seeders.Seeder1")
	// mockSeeder.AssertCalled(t, "GetSeeder", "seeders.Seeder2")
}

func TestSeedCommand_ConfirmToProceed(t *testing.T) {
	// Create mock config
	mockConfig := &configmocks.Config{}
	mockConfig.On("Env", "APP_ENV").Return("production")

	// Create command instance
	command := NewSeedCommand(mockConfig, nil)

	// Create mock context with force option
	mockContext := &consolemocks.Context{}
	mockContext.On("OptionBool", "force").Return(true).Once()

	// Test ConfirmToProceed method with force option
	err := command.ConfirmToProceed(mockContext)
	assert.NoError(t, err)

	// Create mock context without force option
	mockContext.On("OptionBool", "force").Return(false).Once()

	// Test ConfirmToProceed method without force option
	err = command.ConfirmToProceed(mockContext)
	assert.EqualError(t, err, "application in production use --force to run this command")
}
