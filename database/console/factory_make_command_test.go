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

func TestFactoryMakeCommand(t *testing.T) {
	factoryMakeCommand := &FactoryMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the factory name", mock.Anything).Return("", errors.New("the factory name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	}), "the factory name cannot be empty")

	mockContext.EXPECT().Argument(0).Return("UserFactory").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Factory created successfully").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/factories/user_factory.go"))
	assert.True(t, file.Contain("database/factories/user_factory.go", "package factories"))
	assert.True(t, file.Contain("database/factories/user_factory.go", "type UserFactory struct"))

	mockContext.EXPECT().Argument(0).Return("UserFactory").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Factory created successfully").Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	}), "the factory already exists. Use the --force or -f flag to overwrite")
	assert.Nil(t, file.Remove("database"))

	mockContext.EXPECT().Argument(0).Return("subdir/DemoFactory").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Factory created successfully").Once()
	assert.Nil(t, factoryMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("database/factories/subdir/demo_factory.go"))
	assert.True(t, file.Contain("database/factories/subdir/demo_factory.go", "package subdir"))
	assert.True(t, file.Contain("database/factories/subdir/demo_factory.go", "type DemoFactory struct"))
	assert.Nil(t, file.Remove("database"))
}
