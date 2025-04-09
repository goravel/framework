package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the controller name", mock.Anything).Return("", errors.New("the controller name cannot be empty")).Once()
	mockContext.EXPECT().Error("the controller name cannot be empty").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UsersController").Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/users_controller.go"))

	mockContext.EXPECT().Argument(0).Return("UsersController").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the controller already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/AuthController").Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/User/auth_controller.go"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "package User"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "type AuthController struct"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Index(ctx http.Context) error {"))
	assert.Nil(t, file.Remove("app"))
}

func TestResourceControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("User/AuthController").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().OptionBool("resource").Return(true).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/User/auth_controller.go"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "package User"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "type AuthController struct"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Index(ctx http.Context) error {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Show(ctx http.Context) error {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Store(ctx http.Context) error {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Update(ctx http.Context) error {"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Destroy(ctx http.Context) error {"))
	assert.Nil(t, file.Remove("app"))
}
