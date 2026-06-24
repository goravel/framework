package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestMiddlewareMakeCommand(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return(nil).Once()
	mockContext.EXPECT().Ask("Enter the middleware name", mock.Anything).Return("", errors.New("the middleware name cannot be empty")).Once()
	mockContext.EXPECT().Error("the middleware name cannot be empty").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"VerifyCsrfToken"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/verify_csrf_token.go"))

	mockContext.EXPECT().Arguments().Return([]string{"VerifyCsrfToken"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the middleware already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"User/Auth"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Middleware created successfully").Once()
	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/http/middleware/User/auth.go"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "func Auth() http.Middleware {"))
	assert.Nil(t, file.Remove("app"))
}

func TestMiddlewareMakeCommand_MultipleNames(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"Authenticate", "Throttle"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Middleware created successfully").Times(2)

	assert.NoError(t, middlewareMakeCommand.Handle(mockContext))

	assert.True(t, file.Exists("app/http/middleware/authenticate.go"))
	assert.True(t, file.Exists("app/http/middleware/throttle.go"))
	assert.Nil(t, file.Remove("app"))
}
