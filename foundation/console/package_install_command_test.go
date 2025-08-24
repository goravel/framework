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

		facade             = "Auth"
		pkg                = "github.com/goravel/package"
		pkgWithVersion     = "github.com/goravel/package@unknown"
		facadeDependencies = map[string][]string{
			"Auth": {
				"Config",
			},
		}
		facadeToPath = map[string]string{
			"Auth":   "github.com/goravel/framework/auth",
			"Config": "github.com/goravel/framework/config",
		}
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
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().Ask("Enter the package/facade name to install", mock.Anything).
					RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
						return "", option[0].Validate("")
					}).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
			assert: func() {
				s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
			},
		},
		{
			name: "go get failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				})
				s.Contains(captureOutput, `go: github.com/goravel/package@unknown: invalid version`)
			},
		},
		{
			name: "package install failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				})
				s.Contains(captureOutput, `no required module provides package github.com/goravel/package/setup; to add it`)
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				})
				s.Contains(captureOutput, `go: modules disabled by GO111MODULE=off`)

			},
		},
		{
			name: "package install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkgWithVersion, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				}), "Package "+pkgWithVersion+" installed successfully")
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(maps.Keys(facadeDependencies), ", ")))
			},
			assert: func() {
				s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
			},
		},
		{
			name: "facades install failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, strings.Join(facadeDependencies[facade], ", "))).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Config"]+"/setup install", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
				mockContext.EXPECT().Error(mock.MatchedBy(func(message string) bool {
					return s.Equal("Failed to install facade Config, error: no required module provides package github.com/goravel/framework/config/setup; to add it:\n\tgo get github.com/goravel/framework/config/setup", message)
				})).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				})
				s.Empty(captureOutput)
			},
		},
		{
			name: "facades install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, strings.Join(facadeDependencies[facade], ", "))).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Config"]+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Auth"]+"/setup install", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				s.Contains(color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				}), "Facade "+facade+" installed successfully")
			},
		},
		{
			name: "install package and facade simultaneously",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg, facade}).Once()

				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()

				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, strings.Join(facadeDependencies[facade], ", "))).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Config"]+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Auth"]+"/setup install", mock.Anything).Return(nil).Once()
			},
			assert: func() {
				captureOutput := color.CaptureOutput(func(w io.Writer) {
					s.NoError(NewPackageInstallCommand(facadeDependencies, facadeToPath).Handle(mockContext))
				})

				s.Contains(captureOutput, "Package "+pkg+" installed successfully")
				s.Contains(captureOutput, "Facade "+facade+" installed successfully")
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
