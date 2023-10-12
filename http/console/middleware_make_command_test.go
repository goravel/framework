package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestMiddlewareMakeCommand(t *testing.T) {
	middlewareMakeCommand := &MiddlewareMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := middlewareMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("VerifyCsrfToken").Once()
	err = middlewareMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/http/middleware/verify_csrf_token.go"))

	mockContext.On("Argument", 0).Return("User/Auth").Once()
	err = middlewareMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/http/middleware/User/auth.go"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/middleware/User/auth.go", "func Auth() http.Middleware {"))
	assert.Nil(t, file.Remove("app"))
}
