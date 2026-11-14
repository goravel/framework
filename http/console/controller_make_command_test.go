package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/file"
)

func TestControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return(nil).Once()
	mockContext.EXPECT().Ask("Enter the controller name", mock.Anything).Return("", errors.New("the controller name cannot be empty")).Once()
	mockContext.EXPECT().Error("the controller name cannot be empty").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"UsersController"}).Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/users_controller.go"))

	mockContext.EXPECT().Arguments().Return([]string{"UsersController"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the controller already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"User/AuthController"}).Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/controllers/User/auth_controller.go"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "package User"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "type AuthController struct"))
	assert.True(t, file.Contain("app/http/controllers/User/auth_controller.go", "func (r *AuthController) Index(ctx http.Context) http.Response {"))
	assert.Nil(t, file.Remove("app"))
}

func TestControllerMakeCommand_MultipleNames(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"UserController", "OrderController"}).Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Times(2)
	mockContext.EXPECT().Option("module").Return("").Times(2)
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Controller created successfully").Times(2)

	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	assert.FileExists(t, "app/http/controllers/user_controller.go")
	assert.FileExists(t, "app/http/controllers/order_controller.go")
	assert.Nil(t, file.Remove("app"))
}

// TestControllerMakeCommand_MultipleNamesPartialFailure verifies that when
// creating several controllers at once, a failure for one name (e.g. the file
// already exists) does not stop the remaining names from being processed,
// mirroring the behavior of tools like `mkdir`.
func TestControllerMakeCommand_MultipleNamesPartialFailure(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	// First create a controller so it already exists, then try to create it again
	// alongside a new one.
	mockContext.EXPECT().Arguments().Return([]string{"UserController"}).Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"UserController", "OrderController"}).Once()
	// UserController: NewMake checks force, then fails because the file exists,
	// so it never reaches the resource flag. OrderController: checks force, then
	// resource, then succeeds.
	mockContext.EXPECT().Option("module").Return("").Times(2)
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().Error("the controller already exists. Use the --force or -f flag to overwrite").Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))

	assert.FileExists(t, "app/http/controllers/user_controller.go")
	assert.FileExists(t, "app/http/controllers/order_controller.go")
	assert.Nil(t, file.Remove("app"))
}

func TestResourceControllerMakeCommand(t *testing.T) {
	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{"User/AuthController"}).Once()
	mockContext.EXPECT().Option("module").Return("").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().OptionBool("resource").Return(true).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
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

// blogModule sets its own directory, so it is a module of the application.
type blogModule struct{}

func (r *blogModule) Name() string { return "blog" }
func (r *blogModule) Path() string { return "modules/blog" }

func TestControllerMakeCommand_Module(t *testing.T) {
	registry := modular.NewRegistry("")
	assert.NoError(t, registry.Register(&blogModule{}))
	modular.SetDefault(registry)
	defer modular.SetDefault(nil)
	defer func() { assert.Nil(t, file.Remove("modules")) }()

	controllerMakeCommand := &ControllerMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"admin/PostController"}).Once()
	mockContext.EXPECT().OptionBool("resource").Return(false).Once()
	mockContext.EXPECT().Option("module").Return("Blog").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Controller created successfully").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
	assert.True(t, file.Contain("modules/blog/http/controllers/admin/post_controller.go", "package admin"))
	assert.False(t, file.Exists("app"))

	mockContext.EXPECT().Arguments().Return([]string{"PostController"}).Once()
	mockContext.EXPECT().Option("module").Return("news").Once()
	mockContext.EXPECT().Error("[modular] module news is not registered").Once()
	assert.Nil(t, controllerMakeCommand.Handle(mockContext))
}
