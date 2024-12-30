package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestObserverMakeCommand(t *testing.T) {
	observerMakeCommand := &ObserverMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the observer name", mock.Anything).Return("", errors.New("the observer name cannot be empty")).Once()
	mockContext.EXPECT().Error("the observer name cannot be empty").Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/observers/user_observer.go"))

	mockContext.EXPECT().Argument(0).Return("UserObserver").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Observer created successfully").Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/observers/user_observer.go"))

	mockContext.EXPECT().Argument(0).Return("UserObserver").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the observer already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/PhoneObserver").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Observer created successfully").Once()
	assert.Nil(t, observerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/observers/User/phone_observer.go"))
	assert.True(t, file.Contain("app/observers/User/phone_observer.go", "package User"))
	assert.True(t, file.Contain("app/observers/User/phone_observer.go", "type PhoneObserver struct"))

	assert.Nil(t, file.Remove("app"))

	mockContext.AssertExpectations(t)
}
