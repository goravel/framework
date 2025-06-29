package console

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/maps"
)

type PackageUninstallCommandTestSuite struct {
	suite.Suite
}

func TestPackageUninstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageUninstallCommandTestSuite))
}

func (s *PackageUninstallCommandTestSuite) TestHandle() {
	var (
		mockContext *mocksconsole.Context

		facade         = "auth"
		pkg            = "github.com/goravel/package"
		pkgWithVersion = "github.com/goravel/package@unknown"
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
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `no required module provides package github.com/goravel/package/setup; to add it`)

			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `go: modules disabled by GO111MODULE=off`)

			},
		},
		{
			name: "package uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				}), "Package "+pkgWithVersion+" uninstalled successfully")
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Argument(0).Return(facade).Once()
				mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(maps.Keys(facadeToPath), ", ")))
			},
			assert: func() {
				s.NoError(NewPackageUninstallCommand().Handle(mockContext))
			},
		},
		{
			name: "facades uninstall failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(facade).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath[facade]+"/setup uninstall", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				})
				if env.IsWindows() {
					s.Contains(captureOutput, `foundation\console\config\app.go: The system cannot find the path specified`)
				} else {
					s.Contains(captureOutput, `foundation/console/config/app.go: no such file or directory`)
				}
			},
		},
		{
			name: "facades uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(facade).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath[facade]+"/setup uninstall", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageUninstallCommand().Handle(mockContext))
				}), "Facade "+facade+" uninstalled successfully")
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
