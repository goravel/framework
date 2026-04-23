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
			flag, ok := got.Flags[0].(*command.StringFlag)
			if !ok {
				s.Fail("First flag is not StringFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "root"},
				{"Aliases", flag.Aliases, []string{"r"}},
				{"Usage", flag.Usage, "The root path of package, default: packages"},
				{"Value", flag.Value, "packages"},
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
		// Create bootstrap/app.go to trigger IsBootstrapSetup() == true
		_ = file.Create("bootstrap/app.go", `package bootstrap

import "github.com/goravel/framework/foundation"

func Boot() {
	foundation.Setup().Start()
}
`)
	}

	s.T().Cleanup(func() {
		_ = file.Remove("bootstrap")
	})

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
			name: "name is sms and use default root",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("sms").Once()
				mockContext.EXPECT().Option("root").Return("packages").Once()
				mockContext.EXPECT().Success("Package created successfully: packages/sms").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
				s.True(file.Exists("packages/sms/README.md"))
				s.True(file.Exists("packages/sms/service_provider.go"))
				s.True(file.Exists("packages/sms/sms.go"))
				s.True(file.Exists("packages/sms/contracts/sms.go"))
				s.True(file.Exists("packages/sms/facades/sms.go"))
				s.True(file.Contain("packages/sms/facades/sms.go", "github.com/goravel/framework/packages/sms"))
				s.True(file.Contain("packages/sms/facades/sms.go", "github.com/goravel/framework/packages/sms/contracts"))
				s.True(file.Exists("packages/sms/setup/stubs.go"))
				s.True(file.Exists("packages/sms/setup/setup.go"))
				s.NoError(file.Remove("packages"))
			},
		},
		{
			name: "name is github.com/goravel/sms and use other root",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("github.com/goravel/sms-aws").Once()
				mockContext.EXPECT().Option("root").Return("package").Once()
				mockContext.EXPECT().Success("Package created successfully: package/github_com_goravel_sms_aws").Once()
			},
			assert: func() {
				s.NoError(NewPackageMakeCommand().Handle(mockContext))
				s.True(file.Exists("package/github_com_goravel_sms_aws/README.md"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/service_provider.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/contracts/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/facades/github_com_goravel_sms_aws.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/setup/stubs.go"))
				s.True(file.Exists("package/github_com_goravel_sms_aws/setup/setup.go"))
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
