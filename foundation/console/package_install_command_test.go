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
		bindings    map[string]binding.Info

		facade  = "Auth"
		pkg     = "github.com/goravel/package"
		options = []console.Choice{
			{Key: "Auth", Value: "Auth"},
			{Key: "Orm", Value: "Orm"},
		}
		installedBindings = []any{binding.Config}
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(s.T())
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
	}

	tests := []struct {
		name                                string
		installedFacadesInTheCurrentCommand []string
		setup                               func()
		assert                              func()
	}{
		{
			name: "go get failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to get package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package install failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to install package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to tidy go.mod file: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " installed successfully").Once()
			},
		},
		{
			name: "facade name is empty, MultiSelect returns error",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().MultiSelect("Select the facades to install", options, mock.Anything).
					Return(nil, assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().MultiSelect("Select the facades to install", options, mock.Anything).
					Return([]string{facade}, nil).Once()
				mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(bindings), ", ")))
			},
		},
		{
			name: "facades install failed",
			setup: func() {
				bindings = map[string]binding.Info{
					binding.Auth: {
						PkgPath:      "github.com/goravel/framework/auth",
						Dependencies: []string{binding.Config, binding.Orm},
					},
					binding.Config: {
						PkgPath: "github.com/goravel/framework/config",
						IsBase:  true,
					},
				}

				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(true).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup install --facade=Auth --module=github.com/goravel/framework", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to install facade %s: %s", "Auth", assert.AnError)).Once()
			},
		},
		{
			name:                                "The install facade has been installed in the current command",
			installedFacadesInTheCurrentCommand: []string{"Orm"},
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, "Orm")).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup install --facade=Auth --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth installed successfully").Once()
			},
		},
		{
			name: "facades install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, "Orm")).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Orm].PkgPath+"/setup install --facade=Orm --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Orm installed successfully").Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup install --facade=Auth --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth installed successfully").Once()
			},
		},
		{
			name: "install package and facade simultaneously",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg, facade}).Once()

				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " installed successfully").Once()

				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, "Orm")).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Orm].PkgPath+"/setup install --facade=Orm --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Orm installed successfully").Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Auth].PkgPath+"/setup install --facade=Auth --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Facade Auth installed successfully").Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings)
			packageInstallCommand.installedFacadesInTheCurrentCommand = test.installedFacadesInTheCurrentCommand

			s.NoError(packageInstallCommand.Handle(mockContext))
		})
	}
}

func (s *PackageInstallCommandTestSuite) TestGetDependenciesThatNeedInstall() {
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
	installedBindings := []any{binding.Config}

	packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings)

	s.ElementsMatch([]string{binding.Orm}, packageInstallCommand.getDependenciesThatNeedInstall(binding.Auth))
}

func TestGetAvailableFacades(t *testing.T) {
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

	assert.ElementsMatch(t, []string{"Auth", "Orm"}, getAvailableFacades(bindings))
}

func TestGetDependencyBindings(t *testing.T) {
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

	assert.ElementsMatch(t, []string{binding.Orm}, getDependencyBindings(binding.Auth, bindings))
}
