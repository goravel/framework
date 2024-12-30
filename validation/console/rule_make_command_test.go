package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestRuleMakeCommand(t *testing.T) {
	requestMakeCommand := &RuleMakeCommand{}
	mockContext := consolemocks.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the rule name", mock.Anything).Return("", errors.New("the rule name cannot be empty")).Once()
	mockContext.EXPECT().Error("the rule name cannot be empty").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/uppercase.go"))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the rule already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/User/phone.go"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "type Phone struct"))
	assert.Nil(t, file.Remove("app"))
}
