package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestPolicyMakeCommand(t *testing.T) {
	policyMakeCommand := &PolicyMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Arguments().Return([]string{""}).Once()
	mockContext.EXPECT().Ask("Enter the policy name", mock.Anything).Return("", errors.New("the policy name cannot be empty")).Once()
	mockContext.EXPECT().Error("the policy name cannot be empty").Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"UserPolicy"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Policy created successfully").Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/policies/user_policy.go"))

	mockContext.EXPECT().Arguments().Return([]string{"UserPolicy"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the policy already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Arguments().Return([]string{"User/AuthPolicy"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Policy created successfully").Once()
	assert.Nil(t, policyMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/policies/User/auth_policy.go"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "package User"))
	assert.True(t, file.Contain("app/policies/User/auth_policy.go", "type AuthPolicy struct {"))

	assert.Nil(t, file.Remove("app"))
}

func TestPolicyMakeCommand_MultipleNames(t *testing.T) {
	policyMakeCommand := &PolicyMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"PostPolicy", "CommentPolicy"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Policy created successfully").Times(2)
	assert.Nil(t, policyMakeCommand.Handle(mockContext))

	assert.True(t, file.Exists("app/policies/post_policy.go"))
	assert.True(t, file.Exists("app/policies/comment_policy.go"))
	assert.Nil(t, file.Remove("app"))
}
