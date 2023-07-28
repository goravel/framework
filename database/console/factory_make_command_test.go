package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestFactoryMakeCommand(t *testing.T) {
	factoryMakeCommand := &FactoryMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))

	mockContext.On("Argument", 0).Return("UserFactory").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/factories/user_factory.go"))
	assert.True(t, file.Contain("database/factories/user_factory.go", "package factories"))
	assert.True(t, file.Contain("database/factories/user_factory.go", "type UserFactory struct"))
	assert.Nil(t, file.Remove("database"))

	mockContext.On("Argument", 0).Return("subdir/DemoFactory").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/factories/subdir/demo_factory.go"))
	assert.True(t, file.Contain("database/factories/subdir/demo_factory.go", "package subdir"))
	assert.True(t, file.Contain("database/factories/subdir/demo_factory.go", "type DemoFactory struct"))
	assert.Nil(t, file.Remove("database"))
}
