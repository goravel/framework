package console

import (
	"testing"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"

	"github.com/stretchr/testify/assert"
)

func TestPolicyMakeCommand(t *testing.T) {
	policyMakeCommand := &PolicyMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := policyMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("UserPolicy").Once()
	err = policyMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/policies/user_policy.go"))

	mockContext.On("Argument", 0).Return("User/AuthPolicy").Once()
	err = policyMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/policies/User/auth_policy.go"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "package User"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "type AuthPolicy struct {"))

	assert.Nil(t, file.Remove("app"))
}
