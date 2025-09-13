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
		installedBindings = []any{binding.Config}
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
		},
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
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(bindings), ", ")))
			},
		},
		{
			name: "facades install failed",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				mockContext.EXPECT().Info(fmt.Sprintf("%s depends on %s, they will be installed simultaneously", facade, "Orm")).Once()
				mockContext.EXPECT().Spinner("> @go run "+bindings[binding.Orm].PkgPath+"/setup install --facade=Orm --module=github.com/goravel/framework", mock.Anything).Return(assert.AnError).Once()
				mockContext.EXPECT().Error(fmt.Sprintf("Failed to install facade %s: %s", "Orm", assert.AnError)).Once()
			},
		},
		{
			name: "facades install success(simulate)",
			setup: func() {
				mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
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

			s.NoError(NewPackageInstallCommand(bindings, installedBindings).Handle(mockContext))
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
