package console

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestEventMakeCommand(t *testing.T) {
	eventMakeCommand := &EventMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the event name", mock.Anything).Return("", errors.New("the event name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, eventMakeCommand.Handle(mockContext))
	}), "the event name cannot be empty")

	mockContext.On("Argument", 0).Return("GoravelEvent").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/goravel_event.go"))

	mockContext.On("Argument", 0).Return("GoravelEvent").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, eventMakeCommand.Handle(mockContext))
	}), "the event already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("Goravel/Event").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/Goravel/event.go"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "package Goravel"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "type Event struct {"))
	assert.Nil(t, file.Remove("app"))
}
