package console

import (
	"io"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
)

type PackageUninstallCommandTestSuite struct {
	suite.Suite
}

func TestPackageUninstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageUninstallCommandTestSuite))
}

func (s *PackageUninstallCommandTestSuite) TestSignature() {
	expected := "package:uninstall"
	s.Require().Equal(expected, NewPackageUninstallCommand().Signature())
}

func (s *PackageUninstallCommandTestSuite) TestDescription() {
	expected := "Uninstall a package"
	s.Require().Equal(expected, NewPackageUninstallCommand().Description())
}

func (s *PackageUninstallCommandTestSuite) TestExtend() {
	cmd := NewPackageUninstallCommand()
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "package"
		s.Require().Equal(expected, got.Category)
	})

	s.Run("should return correct args usage", func() {
		expected := " <package>"
		s.Require().Equal(expected, got.ArgsUsage)
	})
}

func (s *PackageUninstallCommandTestSuite) TestHandle() {
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
			name: "package name is empty",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("").Once()
				mockContext.EXPECT().Ask("Enter the package name to uninstall", mock.Anything).
						RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
							return "", option[0].Validate("")
						}).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
			assert: func() {
				s.NoError(NewPackageUninstallCommand().Handle(mockContext))
			},
		},

		{
			name: "package uninstall failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup uninstall", mock.Anything).
						RunAndReturn(func(s string, option console.SpinnerOption) error {
							return option.Action()
						}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, "failed to uninstall package:")
				s.Contains(captureOutput, `package package/setup is not in std`)

			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
						RunAndReturn(func(s string, option console.SpinnerOption) error {
							return option.Action()
						}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, "failed to tidy go.mod file:")
				s.Contains(captureOutput, `go: modules disabled by GO111MODULE=off`)

			},
		},
		{
			name: "package uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				}), "Package package@unknown uninstalled successfully")
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
