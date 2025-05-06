package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestRuleMakeCommand(t *testing.T) {
	requestMakeCommand := &RuleMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the rule name", mock.Anything).Return("", errors.New("the rule name cannot be empty")).Once()
	mockContext.EXPECT().Error("the rule name cannot be empty").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
		return strings.HasPrefix(msg, "rule register failed:")
	})).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/uppercase.go"))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the rule already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Rule created successfully").Once()
	mockContext.EXPECT().Success("Rule registered successfully").Once()
	assert.NoError(t, file.PutContent("app/providers/validation_service_provider.go", validationServiceProvider))
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/User/phone.go"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "type Phone struct"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "user_phone"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "app/rules/User"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "&User.Phone{}"))
	assert.Nil(t, file.Remove("app"))
}
