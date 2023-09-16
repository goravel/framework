package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := controllerMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("UsersController").Once()
	mockContext.On("OptionBool", "resource").Return(false).Once()
	err = controllerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/http/controllers/users_controller.go"))

	mockContext.On("Argument", 0).Return("User/AuthController").Once()
	mockContext.On("OptionBool", "resource").Return(false).Once()
	err = controllerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
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
	mockContext.On("OptionBool", "resource").Return(true).Once()
	err := controllerMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
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
