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

func TestModelMakeCommand(t *testing.T) {
	modelMakeCommand := &ModelMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the model name", mock.Anything).Return("", errors.New("the model name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, modelMakeCommand.Handle(mockContext))
	}), "the model name cannot be empty")
	assert.False(t, file.Exists("app/models/user.go"))

	mockContext.On("Argument", 0).Return("User").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/user.go"))

	mockContext.On("Argument", 0).Return("User").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, modelMakeCommand.Handle(mockContext))
	}), "the model already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("User/Phone").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/User/phone.go"))
	assert.True(t, file.Contain("app/models/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/models/User/phone.go", "type Phone struct"))

	assert.Nil(t, file.Remove("app"))

	mockContext.AssertExpectations(t)
}
