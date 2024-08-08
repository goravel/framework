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

func TestFilterMakeCommand(t *testing.T) {
	requestMakeCommand := &FilterMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the filter name", mock.Anything).Return("", errors.New("the filter name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the filter name cannot be empty")

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, requestMakeCommand.Handle(mockContext))
	}), "the filter already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("Custom/Append").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/Custom/append.go"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "package Custom"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "type Append struct"))
	assert.Nil(t, file.Remove("app"))
}
