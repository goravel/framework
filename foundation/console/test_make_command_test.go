package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestTestMakeCommand(t *testing.T) {
	testMakeCommand := &TestMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := testMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("UserTest").Once()
	err = testMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("tests/user_test.go"))

	mockContext.On("Argument", 0).Return("user/UserTest").Once()
	err = testMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("tests/user/user_test.go"))
	assert.True(t, file.Contain("tests/user/user_test.go", "package user"))
	assert.True(t, file.Contain("tests/user/user_test.go", "type UserTestSuite struct"))
	assert.True(t, file.Contain("tests/user/user_test.go", "func (s *UserTestSuite) SetupTest() {"))
	assert.Nil(t, file.Remove("tests"))
}
