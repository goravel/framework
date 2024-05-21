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

func TestListenerMakeCommand(t *testing.T) {
	listenerMakeCommand := &ListenerMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the listener name", mock.Anything).Return("", errors.New("the listener name cannot be empty")).Once()
	err := listenerMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "the listener name cannot be empty")

	mockContext.On("Argument", 0).Return("GoravelListen").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	err = listenerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/listeners/goravel_listen.go"))

	mockContext.On("Argument", 0).Return("GoravelListen").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, listenerMakeCommand.Handle(mockContext))
	}), "The listener already exists. Use the --force flag to overwrite")

	mockContext.On("Argument", 0).Return("Goravel/Listen").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	err = listenerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/listeners/Goravel/listen.go"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "package Goravel"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "type Listen struct {"))
	assert.Nil(t, file.Remove("app"))
}
