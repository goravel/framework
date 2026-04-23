package console

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

type TestMakeCommandTestSuite struct {
	suite.Suite
}

func TestTestMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(TestMakeCommandTestSuite))
}

func (s *TestMakeCommandTestSuite) TestSignature() {
	cmd := NewTestMakeCommand()
	expected := "make:test"
	s.Require().Equal(expected, cmd.Signature())
}

func (s *TestMakeCommandTestSuite) TestDescription() {
	cmd := NewTestMakeCommand()
	expected := "Create a new test class"
	s.Require().Equal(expected, cmd.Description())
}

func (s *TestMakeCommandTestSuite) TestExtend() {
	cmd := NewTestMakeCommand()
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "make"
		s.Require().Equal(expected, got.Category)
	})

	if len(got.Flags) > 0 {
		s.Run("should have correctly configured StringFlag", func() {
			flag, ok := got.Flags[0].(*command.BoolFlag)
			if !ok {
				s.Fail("First flag is not BoolFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "force"},
				{"Aliases", flag.Aliases, []string{"f"}},
				{"Usage", flag.Usage, "Create the test even if it already exists"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *TestMakeCommandTestSuite) TestTestHandle() {
	cmd := NewTestMakeCommand()
	mockContext := mocksconsole.NewContext(s.T())

	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the test name", mock.Anything).Return("", errors.New("the test name cannot be empty")).Once()
	mockContext.EXPECT().Error("the test name cannot be empty").Once()
	s.Nil(cmd.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Test created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("tests/user_test.go"))

	mockContext.EXPECT().Argument(0).Return("UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the test already exists. Use the --force or -f flag to overwrite").Once()
	s.Nil(cmd.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("user/UserTest").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Test created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("tests/user/user_test.go"))
	s.True(file.Contain("tests/user/user_test.go", "package user"))
	s.True(file.Contain("tests/user/user_test.go", "type UserTestSuite struct"))
	s.True(file.Contain("tests/user/user_test.go", "func (s *UserTestSuite) SetupTest() {"))
	s.True(file.Contain("tests/user/user_test.go", "framework/tests"))
	s.True(file.Contain("tests/user/user_test.go", "tests.TestCase"))
	s.NoError(file.Remove("tests"))
}
