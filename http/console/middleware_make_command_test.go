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

func TestMiddlewareMakeCommand(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the middleware name", mock.Anything).Return("", errors.New("the middleware name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, middlewareMakeCommand.Handle(mockContext))
	}), "the middleware name cannot be empty")

	mockContext.On("Argument", 0).Return("VerifyCsrfToken").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/verify_csrf_token.go"))

	mockContext.On("Argument", 0).Return("VerifyCsrfToken").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, middlewareMakeCommand.Handle(mockContext))
	}), "the middleware already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/Auth").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/User/auth.go"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "func Auth() http.Middleware {"))
	assert.Nil(t, file.Remove("app"))
}
