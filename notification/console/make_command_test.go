package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestNotificationMakeCommand(t *testing.T) {
	notificationMakeCommand := &NotificationMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the notification name", mock.Anything).Return("", errors.New("the notification name cannot be empty")).Once()
	mockContext.EXPECT().Error("the notification name cannot be empty").Once()
	assert.NoError(t, notificationMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/notifications/invoice_paid.go"))

	mockContext.EXPECT().Argument(0).Return("InvoicePaid").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Notification created successfully").Once()
	assert.NoError(t, notificationMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/notifications/invoice_paid.go"))

	mockContext.EXPECT().Argument(0).Return("InvoicePaid").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the notification already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, notificationMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Billing/InvoicePaid").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Notification created successfully").Once()
	assert.NoError(t, notificationMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/notifications/Billing/invoice_paid.go"))
	assert.True(t, file.Contain("app/notifications/Billing/invoice_paid.go", "package Billing"))
	assert.True(t, file.Contain("app/notifications/Billing/invoice_paid.go", "type InvoicePaid struct"))

	assert.Nil(t, file.Remove("app"))
}
