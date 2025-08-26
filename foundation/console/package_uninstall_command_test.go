package console

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
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

		facade             = "Auth"
		pkg                = "github.com/goravel/package"
		pkgWithVersion     = "github.com/goravel/package@unknown"
		facadeDependencies = map[string][]string{
			"Auth": {
				"Config",
				"Orm",
			},
		}
		facadeToPath = map[string]string{
			"Auth":   "github.com/goravel/framework/auth",
			"Config": "github.com/goravel/framework/config",
			"Orm":    "github.com/goravel/framework/database",
		}
		baseFacades      = []string{"Config"}
		installedFacades = []string{"Auth", "Config", "Orm"}
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
				mockContext.EXPECT().Ask("Enter the package name to uninstall", mock.Anything).
					RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
						return "", option[0].Validate("")
					}).Once()
				mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
		},
		{
			name: "package uninstall failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
				mockContext.EXPECT().Error("failed to uninstall package: no required module provides package github.com/goravel/package/setup; to add it:\n\tgo get github.com/goravel/package/setup").Once()
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
				mockContext.EXPECT().Error("failed to tidy go.mod file: go: modules disabled by GO111MODULE=off; see 'go help modules'").Once()
			},
		},
		{
			name: "package uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkgWithVersion}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkgWithVersion + " uninstalled successfully").Once()
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
		},
		{
			name: "facades uninstall failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Confirm("Do you want to remove the dependency facades as well: Orm?").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Orm"]+"/setup uninstall", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Error(mock.MatchedBy(func(message string) bool {
					return s.Contains(message, "Failed to uninstall facade Orm, error:")
				})).Once()
			},
		},
		{
			name: "facades uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Confirm("Do you want to remove the dependency facades as well: Orm?").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Auth"]+"/setup uninstall", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
			},
		},
		{
			name: "facades uninstall partial success",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Confirm("Do you want to remove the dependency facades as well: Orm?").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Orm"]+"/setup uninstall", mock.Anything).
					RunAndReturn(func(s string, option console.SpinnerOption) error {
						return option.Action()
					}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Auth"]+"/setup uninstall", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Error(mock.MatchedBy(func(message string) bool {
					return s.Contains(message, "Failed to uninstall facade Orm, error:")
				})).Once()
				mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
			},
		},
		{
			name: "install package and facade simultaneously",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg, facade}).Once()

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " uninstalled successfully").Once()

				mockContext.EXPECT().Confirm("Do you want to remove the dependency facades as well: Orm?").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Orm"]+"/setup uninstall", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Orm uninstalled successfully").Once()
				mockContext.EXPECT().Spinner("> @go run "+facadeToPath["Auth"]+"/setup uninstall", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
			},
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			s.NoError(NewPackageUninstallCommand(facadeDependencies, facadeToPath, baseFacades, installedFacades).Handle(mockContext))
		})
	}
}

func (s *PackageUninstallCommandTestSuite) TestGetFacadesThatNeedUninstall() {
	facadeDependencies := map[string][]string{
		"Auth": {
			"Config",
			"Orm",
			"Log",
		},
		"DB": {
			"Log",
		},
		"Orm": {
			"Config",
		},
	}
	baseFacades := []string{"Config"}
	installedFacades := []string{"Auth", "Config", "DB", "Orm", "Log"}

	packageUninstallCommand := NewPackageUninstallCommand(facadeDependencies, nil, baseFacades, installedFacades)

	s.ElementsMatch([]string{"Auth", "Orm"}, packageUninstallCommand.getFacadesThatNeedUninstall("Auth"))
}
