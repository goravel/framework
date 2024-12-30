package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestEventMakeCommand(t *testing.T) {
	eventMakeCommand := &EventMakeCommand{}
	mockContext := consolemocks.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the event name", mock.Anything).Return("", errors.New("the event name cannot be empty")).Once()
	mockContext.EXPECT().Error("the event name cannot be empty").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("GoravelEvent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/goravel_event.go"))

	mockContext.EXPECT().Argument(0).Return("GoravelEvent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the event already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Goravel/Event").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/Goravel/event.go"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "package Goravel"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "type Event struct {"))
	assert.Nil(t, file.Remove("app"))
}
