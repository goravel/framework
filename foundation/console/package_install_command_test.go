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
	"github.com/goravel/framework/support/maps"
)

type PackageInstallCommandTestSuite struct {
	suite.Suite
}

func TestPackageInstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageInstallCommandTestSuite))
}

func (s *PackageInstallCommandTestSuite) TestHandle() {
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
				mockContext.EXPECT().Ask("Enter the package/facade name to install", mock.Anything).
					RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
						return "", option[0].Validate("")
					}).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
			assert: func() {
				s.NoError(NewPackageInstallCommand().Handle(mockContext))
			},
		},
		{
			name: "go get failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `go: github.com/goravel/package@unknown: invalid version: git ls-remote -q origin in`)
			},
		},
		{
			name: "package install failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `no required module provides package github.com/goravel/package/setup; to add it`)
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `go: modules disabled by GO111MODULE=off`)

			},
		},
		{
			name: "package install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(pkgWithVersion).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				}), "Package "+pkgWithVersion+" installed successfully")
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
				s.NoError(NewPackageInstallCommand().Handle(mockContext))
			},
		},
		{
			name: "facades install failed",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(facade).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath[facade]+"/setup install", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				})
				s.Contains(captureOutput, `foundation/console/config/app.go: no such file or directory`)
			},
		},
		{
			name: "facades install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Argument(0).Return(facade).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath[facade]+"/setup install", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand().Handle(mockContext))
				}), "Facade "+facade+" installed successfully")
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
