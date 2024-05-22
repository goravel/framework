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

func TestTestMakeCommand(t *testing.T) {
	testMakeCommand := &TestMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the test name", mock.Anything).Return("", errors.New("the test name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, testMakeCommand.Handle(mockContext))
	}), "the test name cannot be empty")

	mockContext.On("Argument", 0).Return("UserTest").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, testMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("tests/user_test.go"))

	mockContext.On("Argument", 0).Return("UserTest").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, testMakeCommand.Handle(mockContext))
	}), "the test already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("user/UserTest").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, testMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("tests/user/user_test.go"))
	assert.True(t, file.Contain("tests/user/user_test.go", "package user"))
	assert.True(t, file.Contain("tests/user/user_test.go", "type UserTestSuite struct"))
	assert.True(t, file.Contain("tests/user/user_test.go", "func (s *UserTestSuite) SetupTest() {"))
	assert.Nil(t, file.Remove("tests"))
}
