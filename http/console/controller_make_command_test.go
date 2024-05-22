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

func TestControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the controller name", mock.Anything).Return("", errors.New("the controller name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	}), "the controller name cannot be empty")

	mockContext.On("Argument", 0).Return("UsersController").Once()
	mockContext.On("OptionBool", "resource").Return(false).Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/users_controller.go"))

	mockContext.On("Argument", 0).Return("UsersController").Once()
	mockContext.On("OptionBool", "resource").Return(false).Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	}), "the controller already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/AuthController").Once()
	mockContext.On("OptionBool", "resource").Return(false).Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/User/auth_controller.go"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "package User"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "type AuthController struct"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Index(ctx http.Context) http.Response {"))
	assert.Nil(t, file.Remove("app"))
}

func TestResourceControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("User/AuthController").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.On("OptionBool", "resource").Return(true).Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/User/auth_controller.go"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "package User"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "type AuthController struct"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Index(ctx http.Context) http.Response {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Show(ctx http.Context) http.Response {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Store(ctx http.Context) http.Response {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Update(ctx http.Context) http.Response {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Destroy(ctx http.Context) http.Response {"))
	assert.Nil(t, file.Remove("app"))
}
