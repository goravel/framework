package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestMailMakeCommand(t *testing.T) {
	modelMakeCommand := &MailMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the mail name", mock.Anything).Return("", errors.New("the mail name cannot be empty")).Once()
	mockContext.EXPECT().Error("the mail name cannot be empty").Once()
	assert.NoError(t, modelMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/mails/new_user.go"))

	mockContext.EXPECT().Argument(0).Return("NewUser").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Mail created successfully").Once()
	assert.NoError(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/mails/new_user.go"))

	mockContext.EXPECT().Argument(0).Return("NewUser").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the mail already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, modelMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/VerifyEmail").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Mail created successfully").Once()
	assert.NoError(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/mails/User/verify_email.go"))
	assert.True(t, file.Contain("app/mails/User/verify_email.go", "package User"))
	assert.True(t, file.Contain("app/mails/User/verify_email.go", "type VerifyEmail struct"))

	assert.Nil(t, file.Remove("app"))
}
