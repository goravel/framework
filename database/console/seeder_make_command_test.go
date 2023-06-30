package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestSeederMakeCommand(t *testing.T) {
	seederMakeCommand := &SeederMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))

	mockContext.On("Argument", 0).Return("UserSeeder").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/user_seeder.go"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "package seeders"))
	assert.True(t, file.Contain("database/seeders/user_seeder.go", "type UserSeeder struct"))
	assert.Nil(t, file.Remove("database"))

	mockContext.On("Argument", 0).Return("subdir/DemoSeeder").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/subdir/demo_seeder.go"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "package subdir"))
	assert.True(t, file.Contain("database/seeders/subdir/demo_seeder.go", "type DemoSeeder struct"))
	assert.Nil(t, file.Remove("database"))
}
