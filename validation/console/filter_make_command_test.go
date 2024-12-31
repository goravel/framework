package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

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
	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/uppercase.go"))

	mockContext.On("Argument", 0).Return("Uppercase").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the filter already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, requestMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Custom/Append").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Filter created successfully").Once()

	assert.NoError(t, requestMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/filters/Custom/append.go"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "package Custom"))
	assert.True(t, file.Contain("app/filters/Custom/append.go", "type Append struct"))
	assert.Nil(t, file.Remove("app"))
}
