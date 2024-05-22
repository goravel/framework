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

func TestObserverMakeCommand(t *testing.T) {
	observerMakeCommand := &ObserverMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the observer name", mock.Anything).Return("", errors.New("the observer name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, observerMakeCommand.Handle(mockContext))
	}), "the observer name cannot be empty")
	assert.False(t, file.Exists("app/observers/user_observer.go"))

	mockContext.On("Argument", 0).Return("UserObserver").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/observers/user_observer.go"))

	mockContext.On("Argument", 0).Return("UserObserver").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, observerMakeCommand.Handle(mockContext))
	}), "the observer already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/PhoneObserver").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/observers/User/phone_observer.go"))
	assert.True(t, file.Contain("app/observers/User/phone_observer.go", "package User"))
	assert.True(t, file.Contain("app/observers/User/phone_observer.go", "type PhoneObserver struct"))

	assert.Nil(t, file.Remove("app"))

	mockContext.AssertExpectations(t)
}
