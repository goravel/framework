package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestModelMakeCommand(t *testing.T) {
	modelMakeCommand := &ModelMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the model name", mock.Anything).Return("", errors.New("the model name cannot be empty")).Once()
	mockContext.EXPECT().Error("the model name cannot be empty").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/models/user.go"))

	mockContext.EXPECT().Argument(0).Return("User").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Model created successfully").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/user.go"))

	mockContext.EXPECT().Argument(0).Return("User").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the model already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Model created successfully").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/User/phone.go"))
	assert.True(t, file.Contain("app/models/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/models/User/phone.go", "type Phone struct"))

	assert.Nil(t, file.Remove("app"))

	mockContext.AssertExpectations(t)
}
