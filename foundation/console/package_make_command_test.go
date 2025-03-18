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

type PackageMakeCommandTestSuite struct {
	suite.Suite
}

func TestPackageMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageMakeCommandTestSuite))
}

func (s *PackageMakeCommandTestSuite) TestSignature() {
	expected := "make:package"
	s.Require().Equal(expected, NewPackageMakeCommand().Signature())
}

func (s *PackageMakeCommandTestSuite) TestDescription() {
	expected := "Create a package template"
	s.Require().Equal(expected, NewPackageMakeCommand().Description())
}

func (s *PackageMakeCommandTestSuite) TestExtend() {
	cmd := NewPackageMakeCommand()
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "make"
		s.Require().Equal(expected, got.Category)
	})

	if len(got.Flags) > 0 {
		s.Run("should have correctly configured StringFlag", func() {
			managerFlag, ok := got.Flags[0].(*command.BoolFlag)
			if !ok {
				s.Fail("First flag is not BoolFlag (got type: %T)", got.Flags[0])
			}

			rootFlag, ok := got.Flags[1].(*command.StringFlag)
			if !ok {
				s.Fail("First flag is not StringFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      interface{}
				expected interface{}
			}{
				{"Name", rootFlag.Name, "root"},
				{"Aliases", rootFlag.Aliases, []string{"r"}},
				{"Usage", rootFlag.Usage, "The root path of package, default: packages"},
				{"Value", rootFlag.Value, "packages"},
				{"Name", managerFlag.Name, "manager"},
				{"Aliases", managerFlag.Aliases, []string{"m"}},
				{"Usage", managerFlag.Usage, "Create a package manager"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *PackageMakeCommandTestSuite) TestHandle() {
	var (
		mockContext *mocksconsole.Context
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(s.T())
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "name is empty",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Ask("Enter the package name", mock.Anything).Return("", errors.New("the package name cannot be empty")).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
			},
		},
		{
			name: "name is sms and use default root(hasn't manager)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("sms").Once()
				mockContext.EXPECT().Option("root").Return("packages").Once()
				mockContext.EXPECT().OptionBool("manager").Return(false).Once()
				mockContext.EXPECT().Success("Package created successfully: packages/sms").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
				s.True(file.Exists("packages/sms/README.md"))
				s.True(file.Exists("packages/sms/service_provider.go"))
				s.True(file.Exists("packages/sms/sms.go"))
				s.True(file.Exists("packages/sms/config/sms.go"))
				s.True(file.Exists("packages/sms/contracts/sms.go"))
				s.True(file.Exists("packages/sms/facades/sms.go"))
				s.True(file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms"))
				s.True(file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms/contracts"))
				s.NoError(file.Remove("packages"))
			},
		},
		{
			name: "name is sms and use default root(has manager)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("sms").Once()
				mockContext.EXPECT().Option("root").Return("packages").Once()
				mockContext.EXPECT().OptionBool("manager").Return(true).Once()
				mockContext.EXPECT().Success("Package created successfully: packages/sms").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
				s.True(file.Exists("packages/sms/README.md"))
				s.True(file.Exists("packages/sms/service_provider.go"))
				s.True(file.Exists("packages/sms/sms.go"))
				s.True(file.Exists("packages/sms/config/sms.go"))
				s.True(file.Exists("packages/sms/contracts/sms.go"))
				s.True(file.Exists("packages/sms/facades/sms.go"))
				s.True(file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms"))
				s.True(file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms/contracts"))
				s.True(file.Exists("packages/sms/manager/manager.go"))
				s.NoError(file.Remove("packages"))
			},
		},
		{
			name: "name is github.com/goravel/sms and use other root",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("github.com/goravel/sms-aws").Once()
				mockContext.EXPECT().Option("root").Return("package").Once()
				mockContext.EXPECT().OptionBool("manager").Return(false).Once()
				mockContext.EXPECT().Success("Package created successfully: package/github_com_goravel_sms_aws").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
				s.True(file.Exists("package/github_com_goravel_sms_aws/README.md"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/service_provider.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/config/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/contracts/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/facades/github_com_goravel_sms_aws.go"))
				s.NoError(file.Remove("package"))
			},
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()
			test.assert()
		})
	}
}
func (s *PackageMakeCommandTestSuite) TestPackageName() {
	input := "github.com/example/package-name"
	expected := "package_name"
	s.Equal(expected, packageName(input))

	input2 := "example.com/another_package.name"
	expected2 := "another_package_name"
	s.Equal(expected2, packageName(input2))
}
