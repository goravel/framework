package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestListenerMakeCommand(t *testing.T) {
	listenerMakeCommand := &ListenerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{""}).Once()
	mockContext.EXPECT().Ask("Enter the listener name", mock.Anything).Return("", errors.New("the listener name cannot be empty")).Once()
	mockContext.EXPECT().Error("the listener name cannot be empty").Once()
	assert.Nil(t, listenerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"GoravelListen"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Listener created successfully").Once()
	assert.Nil(t, listenerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/listeners/goravel_listen.go"))

	mockContext.EXPECT().Arguments().Return([]string{"GoravelListen"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the listener already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, listenerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"Goravel/Listen"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Listener created successfully").Once()
	assert.Nil(t, listenerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/listeners/Goravel/listen.go"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "package Goravel"))
	assert.True(t, file.Contain("app/listeners/Goravel/listen.go", "type Listen struct {"))
	assert.Nil(t, file.Remove("app"))
}

func TestListenerMakeCommand_MultipleNames(t *testing.T) {
	listenerMakeCommand := &ListenerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"SendOrderEmail", "SendReceiptEmail"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Listener created successfully").Times(2)
	assert.Nil(t, listenerMakeCommand.Handle(mockContext))

	assert.True(t, file.Exists("app/listeners/send_order_email.go"))
	assert.True(t, file.Exists("app/listeners/send_receipt_email.go"))
	assert.Nil(t, file.Remove("app"))
}
