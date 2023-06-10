package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestMakeCommand(t *testing.T) {
	makeCommand := &MakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("CleanCache").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/console/commands/clean_cache.go"))

	mockContext.On("Argument", 0).Return("Goravel/CleanCache").Once()
	assert.Nil(t, makeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/console/commands/goravel/clean_cache.go"))

	assert.True(t, file.Remove("app"))
}
