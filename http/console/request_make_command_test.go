package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestRequestMakeCommand(t *testing.T) {
	requestMakeCommand := &RequestMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the request name", mock.Anything).Return("", errors.New("the request name cannot be empty")).Once()
	mockContext.EXPECT().Error("the request name cannot be empty").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("CreateUser").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Request created successfully").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/requests/create_user.go"))

	mockContext.EXPECT().Argument(0).Return("CreateUser").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the request already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Auth").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Request created successfully").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/requests/User/auth.go"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "type Auth struct"))
	assert.Nil(t, file.Remove("app"))
}
