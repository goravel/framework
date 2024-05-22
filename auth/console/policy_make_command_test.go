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

func TestPolicyMakeCommand(t *testing.T) {
	policyMakeCommand := &PolicyMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the policy name", mock.Anything).Return("", errors.New("the policy name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, policyMakeCommand.Handle(mockContext))
	}), "the policy name cannot be empty")

	mockContext.On("Argument", 0).Return("UserPolicy").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/policies/user_policy.go"))

	mockContext.On("Argument", 0).Return("UserPolicy").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, policyMakeCommand.Handle(mockContext))
	}), "the policy already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/AuthPolicy").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/policies/User/auth_policy.go"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "package User"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "type AuthPolicy struct {"))

	assert.Nil(t, file.Remove("app"))
}
