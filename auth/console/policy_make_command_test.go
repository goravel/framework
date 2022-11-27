package console

import (
	"testing"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"

	"github.com/stretchr/testify/assert"
)

func TestEventMakeCommand(t *testing.T) {
	policyMakeCommand := &PolicyMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := policyMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("UserPolicy").Once()
	err = policyMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/policies/user_policy.go"))
	assert.True(t, file.Remove("app"))
}
