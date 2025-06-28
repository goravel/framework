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

type PackageInstallCommandTestSuite struct {
	suite.Suite
}

func TestPackageInstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageInstallCommandTestSuite))
}

func (s *PackageInstallCommandTestSuite) TestSignature() {
	expected := "package:install"
	s.Require().Equal(expected, NewPackageInstallCommand(nil).Signature())
}

func (s *PackageInstallCommandTestSuite) TestDescription() {
	expected := "Install a package"
	s.Require().Equal(expected, NewPackageInstallCommand(nil).Description())
}

func (s *PackageInstallCommandTestSuite) TestExtend() {
	cmd := NewPackageInstallCommand(nil)
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "package"
		s.Require().Equal(expected, got.Category)
	})

	s.Run("should return correct args usage", func() {
		expected := " <package@version>"
		s.Require().Equal(expected, got.ArgsUsage)
	})
}

func (s *PackageInstallCommandTestSuite) TestHandle() {
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
				mockContext.EXPECT().Ask("Enter the package name to install", mock.Anything).
					RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
						return "", option[0].Validate("")
					}).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
			assert: func() {
				s.NoError(NewPackageInstallCommand(nil).Handle(mockContext))
			},
		},
		{
			name: "go get failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().Spinner("> @go get package@unknown", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(nil).Handle(mockContext))
				})
				s.Contains(captureOutput, "failed to get package:")
				s.Contains(captureOutput, `go: package@unknown: malformed module path "package": missing dot in first path element`)

			},
		},
		{
			name: "package install failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().Spinner("> @go get package@unknown", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup install", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(nil).Handle(mockContext))
				})
				s.Contains(captureOutput, "failed to install package:")
				s.Contains(captureOutput, `package package/setup is not in std`)

			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().Spinner("> @go get package@unknown", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(nil).Handle(mockContext))
				})
				s.Contains(captureOutput, "failed to tidy go.mod file:")
				s.Contains(captureOutput, `go: modules disabled by GO111MODULE=off`)

			},
		},
		{
			name: "package install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return("package@unknown").Once()
				mockContext.EXPECT().Spinner("> @go get package@unknown", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run package/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(nil).Handle(mockContext))
				}), "Package package@unknown installed successfully")
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
