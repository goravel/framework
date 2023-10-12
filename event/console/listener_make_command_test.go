package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestListenerMakeCommand(t *testing.T) {
	listenerMakeCommand := &ListenerMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := listenerMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("GoravelListen").Once()
	err = listenerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/listeners/goravel_listen.go"))

	mockContext.On("Argument", 0).Return("Goravel/Listen").Once()
	err = listenerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/listeners/Goravel/listen.go"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "package Goravel"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "type Listen struct {"))
	assert.Nil(t, file.Remove("app"))
}
