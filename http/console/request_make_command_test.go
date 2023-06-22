package console

import (
	"testing"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"

	"github.com/stretchr/testify/assert"
)

func TestRequestMakeCommand(t *testing.T) {
	requestMakeCommand := &RequestMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := requestMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("CreateUser").Once()
	err = requestMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/http/requests/create_user.go"))

	mockContext.On("Argument", 0).Return("User/Auth").Once()
	err = requestMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/http/requests/User/auth.go"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "package User"))
	assert.True(t, file.Contain("app/http/requests/User/auth.go", "type Auth struct"))
	assert.Nil(t, file.Remove("app"))
}
