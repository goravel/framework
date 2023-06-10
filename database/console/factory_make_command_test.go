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
	assert.False(t, file.Exists("database/factories/userFactory.go"))

	mockContext.On("Argument", 0).Return("UserFactory").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/factories/userFactory.go"))
	assert.True(t, file.Remove("database"))
}