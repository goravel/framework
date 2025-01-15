package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestTestMakeCommand(t *testing.T) {
	testMakeCommand := &TestMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the test name", mock.Anything).Return("", errors.New("the test name cannot be empty")).Once()
	mockContext.EXPECT().Error("the test name cannot be empty").Once()
	assert.Nil(t, testMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Test created successfully").Once()
	assert.NoError(t, testMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("tests/user_test.go"))

	mockContext.EXPECT().Argument(0).Return("UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the test already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, testMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("user/UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Test created successfully").Once()
	assert.NoError(t, testMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("tests/user/user_test.go"))
	assert.True(t, file.Contain("tests/user/user_test.go", "package user"))
	assert.True(t, file.Contain("tests/user/user_test.go", "type UserTestSuite struct"))
	assert.True(t, file.Contain("tests/user/user_test.go", "func (s *UserTestSuite) SetupTest() {"))
	assert.NoError(t, file.Remove("tests"))
}
