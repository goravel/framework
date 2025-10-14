package console

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/file"
)

type PackageUninstallCommandTestSuite struct {
	suite.Suite
}

func TestPackageUninstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageUninstallCommandTestSuite))
}

func (s *PackageUninstallCommandTestSuite) TestHandle() {
	var (
		mockApp     *mocksfoundation.Application
		mockContext *mocksconsole.Context

		facade   = "Auth"
		pkg      = "github.com/goravel/package"
		bindings = map[string]binding.Info{
			binding.Auth: {
				PkgPath:      "github.com/goravel/framework/auth",
				Dependencies: []string{binding.Config, binding.Orm},
			},
			binding.Config: {
				PkgPath: "github.com/goravel/framework/config",
				IsBase:  true,
			},
			binding.Orm: {
				PkgPath:      "github.com/goravel/framework/database",
				Dependencies: []string{binding.Config},
			},
		}
		installedBindings = []any{binding.Auth, binding.Config, binding.Orm}
	)

	beforeEach := func() {
		mockApp = mocksfoundation.NewApplication(s.T())
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
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("failed to uninstall package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("failed to tidy go.mod file: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup uninstall --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " uninstalled successfully").Once()
			},
		},
		{
			name: "facade is a base facade",
			setup: func() {
				facade := "Config"
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Warning(fmt.Sprintf("Facade %s is a base facade, cannot be uninstalled", facade)).Once()
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(bindings), ", ")))
			},
		},
		{
			name: "facades can't be uninstalled because of existing upper dependencies",
			setup: func() {
				facade := "Orm"

				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()

				s.NoError(file.PutContent("test_auth.go", "package facades\n"))

				mockApp.EXPECT().FacadesPath("auth.go").Return("test_auth.go").Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Facade %s is depended on %s facades, cannot be uninstalled", facade, "Auth")).Once()
			},
		},
		{
			name: "facades uninstall failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().OptionBool("force").Return(false).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup uninstall --facade=Auth", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to uninstall facade %s, error: %s", "Auth", assert.AnError)).Once()
			},
		},
		{
			name: "facades uninstall success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup uninstall --facade=Auth --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
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

				mockContext.EXPECT().OptionBool("force").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup uninstall --facade=Auth --force", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			s.NoError(NewPackageUninstallCommand(mockApp, bindings, installedBindings).Handle(mockContext))

			s.NoError(file.Remove("test_auth.go"))
		})
	}
}

func (s *PackageUninstallCommandTestSuite) TestGetBindingsThatNeedUninstall() {
	bindings := map[string]binding.Info{
		binding.Auth: {
			PkgPath:      "github.com/goravel/framework/auth",
			Dependencies: []string{binding.Config, binding.Orm},
		},
		binding.Config: {
			PkgPath: "github.com/goravel/framework/config",
			IsBase:  true,
		},
		binding.Orm: {
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{binding.Config},
		},
	}

	installedBindings := []any{binding.Auth, binding.Config, binding.DB, binding.Orm, binding.Log}

	packageUninstallCommand := NewPackageUninstallCommand(nil, bindings, installedBindings)

	s.ElementsMatch([]string{binding.Auth, binding.Orm}, packageUninstallCommand.getBindingsThatNeedUninstall(binding.Auth))
}

func (s *PackageUninstallCommandTestSuite) TestGetExistingUpperDependencyFacades() {
	bindings := map[string]binding.Info{
		binding.Auth: {
			PkgPath:      "github.com/goravel/framework/auth",
			Dependencies: []string{binding.Config, binding.Orm},
		},
		binding.Config: {
			PkgPath: "github.com/goravel/framework/config",
			IsBase:  true,
		},
		binding.Orm: {
			PkgPath:      "github.com/goravel/framework/database",
			Dependencies: []string{binding.Config},
		},
	}

	installedBindings := []any{binding.Auth, binding.Config, binding.DB, binding.Orm, binding.Log}

	s.Run("upper dependencies exist", func() {
		s.NoError(file.PutContent("test_auth.go", "package facades\n"))
		defer func() {
			s.NoError(file.Remove("test_auth.go"))
		}()

		mockApp := mocksfoundation.NewApplication(s.T())
		mockApp.EXPECT().FacadesPath("auth.go").Return("test_auth.go").Once()
		packageUninstallCommand := NewPackageUninstallCommand(mockApp, bindings, installedBindings)

		s.ElementsMatch([]string{"Auth"}, packageUninstallCommand.getExistingUpperDependencyFacades(binding.Orm))
	})

	s.Run("upper dependencies do not exist", func() {
		mockApp := mocksfoundation.NewApplication(s.T())
		mockApp.EXPECT().FacadesPath("auth.go").Return("test_auth.go").Once()
		packageUninstallCommand := NewPackageUninstallCommand(mockApp, bindings, installedBindings)

		s.Empty(packageUninstallCommand.getExistingUpperDependencyFacades(binding.Orm))
	})
}
