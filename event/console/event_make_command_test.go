package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestEventMakeCommand(t *testing.T) {
	eventMakeCommand := &EventMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := eventMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("GoravelEvent").Once()
	err = eventMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/events/goravel_event.go"))

	mockContext.On("Argument", 0).Return("Goravel/Event").Once()
	err = eventMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/events/Goravel/event.go"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "package Goravel"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "type Event struct {"))
	assert.Nil(t, file.Remove("app"))
}
