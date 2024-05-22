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

func TestRequestMakeCommand(t *testing.T) {
	requestMakeCommand := &RequestMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the request name", mock.Anything).Return("", errors.New("the request name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the request name cannot be empty")

	mockContext.On("Argument", 0).Return("CreateUser").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/requests/create_user.go"))

	mockContext.On("Argument", 0).Return("CreateUser").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the request already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/Auth").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/requests/User/auth.go"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "type Auth struct"))
	assert.Nil(t, file.Remove("app"))
}
