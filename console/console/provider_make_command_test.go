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

type ProviderMakeCommandTestSuite struct {
	suite.Suite
}

func TestProviderMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderMakeCommandTestSuite))
}

func (s *ProviderMakeCommandTestSuite) TestSignature() {
	expected := "make:provider"
	cmd := &ProviderMakeCommand{}
	
	s.Require().Equal(expected, cmd.Signature())
}

func (s *ProviderMakeCommandTestSuite) TestDescription() {
	expected := "Create a new service provider class"
	cmd := &ProviderMakeCommand{}
	
	s.Require().Equal(expected, cmd.Description())
}

func (s *ProviderMakeCommandTestSuite) TestExtend() {
	cmd := &ProviderMakeCommand{}
	extend := cmd.Extend()
	
	s.Run("should return correct category", func() {
		expected := "make"
		s.Require().Equal(expected, extend.Category)
	})
	
	s.Run("should have correct number of flags", func() {
		s.Require().Len(extend.Flags, 1)
	})
	
	if len(extend.Flags) > 0 {
		s.Run("should have correctly configured BoolFlag", func() {
			flag, ok := extend.Flags[0].(*command.BoolFlag)
			if !ok {
				s.Fail("First flag is not BoolFlag (got type: %T)", extend.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "force"},
				{"Aliases", flag.Aliases, []string{"f"}},
				{"Usage", flag.Usage, "Create the provider even if it already exists"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *ProviderMakeCommandTestSuite) TestHandle() {
	cmd := &ProviderMakeCommand{}
	mockContext := mocksconsole.NewContext(s.T())

	// Test empty name
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the provider name", mock.Anything).Return("", errors.New("the provider name cannot be empty")).Once()
	mockContext.EXPECT().Error("the provider name cannot be empty").Once()
	s.Nil(cmd.Handle(mockContext))

	// Test successful creation
	mockContext.EXPECT().Argument(0).Return("UserServiceProvider").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/user_service_provider.go"))

	// Test file already exists without force
	mockContext.EXPECT().Argument(0).Return("UserServiceProvider").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the provider already exists. Use the --force or -f flag to overwrite").Once()
	s.Nil(cmd.Handle(mockContext))

	// Test file already exists with force
	mockContext.EXPECT().Argument(0).Return("UserServiceProvider").Once()
	mockContext.EXPECT().OptionBool("force").Return(true).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/user_service_provider.go"))

	// Test nested provider creation
	mockContext.EXPECT().Argument(0).Return("auth/AuthServiceProvider").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Provider created successfully").Once()
	s.NoError(cmd.Handle(mockContext))
	s.True(file.Exists("app/providers/auth/auth_service_provider.go"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "package auth"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "type AuthServiceProvider struct{}"))
	s.True(file.Contain("app/providers/auth/auth_service_provider.go", "func (r *AuthServiceProvider) Register(app foundation.Application) {"))
	
	// Clean up test files
	s.NoError(file.Remove("app"))
}