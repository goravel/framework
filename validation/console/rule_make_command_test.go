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

func TestRuleMakeCommand(t *testing.T) {
	requestMakeCommand := &RuleMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the rule name", mock.Anything).Return("", errors.New("the rule name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the rule name cannot be empty")

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the rule already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/Phone").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/rules/User/phone.go"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/rules/User/phone.go", "type Phone struct"))
	assert.Nil(t, file.Remove("app"))
}
