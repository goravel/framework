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
	"github.com/goravel/framework/foundation/json"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
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
			{Key: "All facades", Value: "all"},
			{Key: "Select facades", Value: "select"},
			{Key: "Third-party package", Value: "third"},
		}
		facadeOptions = []console.Choice{
			{Key: fmt.Sprintf("%-11s", "Auth") + color.Gray().Sprintf(" - %s", "Description"), Value: "Auth"},
			{Key: "Orm", Value: "Orm"},
		}
		installedBindings = []any{binding.Config}
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(s.T())
		bindings = map[string]binding.Info{
			binding.Auth: {
				Description:  "Description",
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
	}{
		{
			name: "go get failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("failed to get package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package install failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("failed to install package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("failed to tidy go.mod file: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " installed successfully").Once()
			},
		},
		{
			name: "facade name is empty, failed to choice what to install",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Choice("Which facades or package do you want to install?", options).
					Return("", assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "facade name is empty, failed to select facades",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Choice("Which facades or package do you want to install?", options).
					Return("select", nil).Once()
				mockContext.EXPECT().MultiSelect("Select the facades to install", facadeOptions, mock.Anything).
					Return(nil, assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "facade name is empty, failed to input third package",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Choice("Which facades or package do you want to install?", options).
					Return("third", nil).Once()
				mockContext.EXPECT().Ask("Enter the package", console.AskOption{
					Description: "E.g.: github.com/goravel/framework or github.com/goravel/framework@master",
				}).Return("", assert.AnError).Once()
				mockContext.EXPECT().Error(assert.AnError.Error()).Once()
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Arguments().Return([]string{}).Once()
				mockContext.EXPECT().OptionBool("all-facades").Return(false).Once()
				mockContext.EXPECT().Choice("Which facades or package do you want to install?", options).
					Return("select", nil).Once()
				mockContext.EXPECT().MultiSelect("Select the facades to install", facadeOptions, mock.Anything).
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
				mockContext.EXPECT().Error(fmt.Sprintf("failed to install facade %s: %s", "Auth", assert.AnError)).Once()
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
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
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

			packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings, json.New())
			packageInstallCommand.installedFacadesInTheCurrentCommand = test.installedFacadesInTheCurrentCommand

			s.NoError(packageInstallCommand.Handle(mockContext))
		})
	}
}

func (s *PackageInstallCommandTestSuite) Test_installFacade_TwoFacadesHaveTheSameDrivers_ShouldOnlyInstallOnce() {
	var (
		mockContext = mocksconsole.NewContext(s.T())
		pkg         = "github.com/goravel/postgres"
		drivers     = []binding.Driver{
			{
				Name:    "Postgres",
				Package: pkg,
			},
			{
				Name:    "MySQL",
				Package: "github.com/goravel/mysql",
			},
		}
		bindings = map[string]binding.Info{
			binding.DB: {
				PkgPath: "github.com/goravel/framework/database",
				Drivers: drivers,
			},
			binding.Orm: {
				PkgPath: "github.com/goravel/framework/database",
				Drivers: drivers,
			},
		}
		dbFacade          = "DB"
		ormFacade         = "Orm"
		installedBindings = []any{}
	)

	packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings, json.New())

	// Install DB facade
	mockContext.EXPECT().Spinner("> @go run "+bindings[binding.DB].PkgPath+"/setup install --facade=DB --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", dbFacade), []console.Choice{
		{Key: "Postgres", Value: pkg},
		{Key: "MySQL", Value: "github.com/goravel/mysql"},
		{Key: "Custom", Value: "Custom"},
	}, console.ChoiceOption{
		Description: fmt.Sprintf("A driver is required for %s, please select one to install.", dbFacade),
	}).Return(pkg, nil).Once()
	mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Twice()
	mockContext.EXPECT().Success("Package " + pkg + " installed successfully").Once()
	mockContext.EXPECT().Success("Facade DB installed successfully").Once()

	s.NoError(packageInstallCommand.installFacade(mockContext, dbFacade))

	// Install Orm facade, should not install the driver again
	mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Orm].PkgPath+"/setup install --facade=Orm --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Success("Facade Orm installed successfully").Once()

	s.NoError(packageInstallCommand.installFacade(mockContext, ormFacade))
}

func (s *PackageInstallCommandTestSuite) Test_installDriver() {
	var (
		mockContext *mocksconsole.Context

		facade      = "Route"
		bindingInfo = binding.Info{
			PkgPath:      "github.com/goravel/framework/route",
			Dependencies: []string{binding.Config},
			Drivers: []binding.Driver{
				{
					Name:    "Route",
					Package: "route",
				},
				{
					Name:        "Gin",
					Description: "Description",
					Package:     "github.com/goravel/gin",
				},
				{
					Name:    "Fiber",
					Package: "github.com/goravel/fiber",
				},
			},
		}
		bindings = map[string]binding.Info{
			binding.Route: bindingInfo,
			binding.Config: {
				PkgPath: "github.com/goravel/framework/config",
				IsBase:  true,
			},
		}
		installedBindings = []any{binding.Config}
	)

	tests := []struct {
		name        string
		bindingInfo binding.Info
		setup       func()
		expectError error
	}{
		{
			name:  "driver is empty",
			setup: func() {},
		},
		{
			name:        "select driver returns error",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("", assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name:        "select custom driver, but ask returns error",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("Custom", nil).Once()
				mockContext.EXPECT().Ask(fmt.Sprintf("Please enter the %s driver package", facade)).Return("", assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name:        "select custom driver, but input empty",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("Custom", nil).Once()
				mockContext.EXPECT().Ask(fmt.Sprintf("Please enter the %s driver package", facade)).Return("", nil).Once()
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("", assert.AnError).Once()
			},
			expectError: assert.AnError,
		},
		{
			name:        "failed to install driver",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("github.com/goravel/gin", nil).Once()
				mockContext.EXPECT().Spinner("> @go get github.com/goravel/gin", mock.Anything).Return(assert.AnError).Once()
			},
			expectError: fmt.Errorf("failed to get package: %s", assert.AnError),
		},
		{
			name:        "successful to install driver",
			bindingInfo: bindingInfo,
			setup: func() {
				pkg := "github.com/goravel/gin"
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return(pkg, nil).Once()
				mockContext.EXPECT().Spinner("> @go get "+pkg, mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go run "+pkg+"/setup install --module=github.com/goravel/framework", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Spinner("> @go mod tidy", mock.Anything).Return(nil).Once()
				mockContext.EXPECT().Success("Package " + pkg + " installed successfully").Once()
			},
		},
		{
			name:        "failed to install internal driver",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("route", nil).Once()
				mockContext.EXPECT().Spinner("> @go run github.com/goravel/framework/route/setup install --driver=route --module=github.com/goravel/framework", mock.Anything).
					Return(assert.AnError).Once()
			},
			expectError: fmt.Errorf("failed to install driver %s: %s", "route", assert.AnError),
		},
		{
			name:        "successful to install driver",
			bindingInfo: bindingInfo,
			setup: func() {
				mockContext.EXPECT().Choice(fmt.Sprintf("Select the %s driver to install", facade), []console.Choice{
					{Key: "Route", Value: "route"},
					{Key: "Gin" + color.Gray().Sprintf(" - %s", "Description"), Value: "github.com/goravel/gin"},
					{Key: "Fiber", Value: "github.com/goravel/fiber"},
					{Key: "Custom", Value: "Custom"},
				}, console.ChoiceOption{
					Description: fmt.Sprintf("A driver is required for %s, please select one to install.", facade),
				}).Return("route", nil).Once()
				mockContext.EXPECT().Spinner("> @go run github.com/goravel/framework/route/setup install --driver=route --module=github.com/goravel/framework", mock.Anything).
					Return(nil).Once()
				mockContext.EXPECT().Success("Driver route installed successfully").Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockContext = mocksconsole.NewContext(s.T())

			tt.setup()

			packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings, json.New())

			s.Equal(tt.expectError, packageInstallCommand.installDriver(mockContext, facade, tt.bindingInfo))
		})
	}
}

func (s *PackageInstallCommandTestSuite) Test_getBindingsToInstall() {
	bindings := map[string]binding.Info{
		binding.Auth: {
			PkgPath:         "github.com/goravel/framework/auth",
			Dependencies:    []string{binding.Config, binding.Orm},
			InstallTogether: []string{binding.Gate},
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

	packageInstallCommand := NewPackageInstallCommand(bindings, installedBindings, json.New())

	s.ElementsMatch([]string{binding.Orm, binding.Gate}, packageInstallCommand.getBindingsToInstall(binding.Auth))
}

func Test_getAvailableFacades(t *testing.T) {
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

func Test_getFacadeDescription(t *testing.T) {
	bindings := map[string]binding.Info{
		binding.Auth: {
			Description:  "Description",
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

	assert.Equal(t, "Description", getFacadeDescription("Auth", bindings))
	assert.Empty(t, getFacadeDescription("Orm", bindings))
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

func Test_isInternalDriver(t *testing.T) {
	assert.False(t, isInternalDriver(""))
	assert.True(t, isInternalDriver("database"))
	assert.False(t, isInternalDriver("github.com/goravel/redis"))
}
