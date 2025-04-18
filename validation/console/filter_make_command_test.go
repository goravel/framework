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

var validationServiceProvider = `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/facades"
)

type ValidationServiceProvider struct {
}

func (receiver *ValidationServiceProvider) Register(app foundation.Application) {

}

func (receiver *ValidationServiceProvider) Boot(app foundation.Application) {
	if err := facades.Validation().AddRules(receiver.rules()); err != nil {
		facades.Log().Errorf("add rules error: %+v", err)
	}
	if err := facades.Validation().AddFilters(receiver.filters()); err != nil {
		facades.Log().Errorf("add filters error: %+v", err)
	}
}

func (receiver *ValidationServiceProvider) rules() []validation.Rule {
	return []validation.Rule{}
}

func (receiver *ValidationServiceProvider) filters() []validation.Filter {
	return []validation.Filter{}
}`

func TestFilterMakeCommand(t *testing.T) {
	requestMakeCommand := &FilterMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the filter name", mock.Anything).Return("", errors.New("the filter name cannot be empty")).Once()
	mockContext.EXPECT().Error("the filter name cannot be empty").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Uppercase").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Filter created successfully").Once()
	mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
		return strings.HasPrefix(msg, "filter register failed:")
	})).Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the filter already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Custom/Append").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Filter created successfully").Once()
	mockContext.EXPECT().Success("Filter registered successfully").Once()
	assert.NoError(t, file.PutContent("app/providers/validation_service_provider.go", validationServiceProvider))
	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/Custom/append.go"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "package Custom"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "type Append struct"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "custom_append"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "app/filters/Custom"))
	assert.True(t, file.Contain("app/providers/validation_service_provider.go", "&Custom.Append{}"))
	assert.Nil(t, file.Remove("app"))
}
