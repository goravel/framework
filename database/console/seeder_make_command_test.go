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
	assert.False(t, file.Exists("database/seeders/user_seeder.go"))

	mockContext.On("Argument", 0).Return("UserSeeder").Once()
	assert.Nil(t, seederMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/seeders/user_seeder.go"))
	assert.True(t, file.Remove("database"))
}
