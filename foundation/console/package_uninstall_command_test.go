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
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/file"
)

type PackageUninstallCommandTestSuite struct {
	suite.Suite
	mockContext *mocksconsole.Context
	mockProcess *mocksprocess.Process
	mockJson    *mocksfoundation.Json
}

func TestPackageUninstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PackageUninstallCommandTestSuite))
}

func (s *PackageUninstallCommandTestSuite) SetupTest() {
	s.mockContext = mocksconsole.NewContext(s.T())
	s.mockProcess = mocksprocess.NewProcess(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
}

func (s *PackageUninstallCommandTestSuite) TestHandle() {
	var (
		facade    = "Auth"
		pkg       = "github.com/goravel/package"
		pathsJSON = `{"App":"app"}`
		bindings  = map[string]binding.Info{
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
	)

	beforeEach := func() {
		s.mockJson.EXPECT().MarshalString(mock.Anything).Return(pathsJSON, nil).Once()
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "package name is empty",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{}).Once()
				s.mockContext.EXPECT().Ask("Enter the package name to uninstall", mock.Anything).
					RunAndReturn(func(_ string, option ...console.AskOption) (string, error) {
						return "", option[0].Validate("")
					}).Once()
				s.mockContext.EXPECT().Error("the package name cannot be empty").Once()
			},
		},
		{
			name: "package uninstall failed",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				failedResult := mockFailedResult(s.T(), assert.AnError)
				s.mockProcess.EXPECT().Run("go", "run", pkg+"/setup", "uninstall", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON).Return(failedResult).Once()
				s.mockProcess.EXPECT().WithSpinner("Uninstalling " + pkg).Return(s.mockProcess).Once()
				s.mockContext.EXPECT().Error(fmt.Sprintf("failed to uninstall package: %s", assert.AnError)).Once()
			},
		},
		{
			name: "tidy go.mod file failed",
			setup: func() {
				s.T().Setenv("GO111MODULE", "off")
				s.mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				s.mockContext.EXPECT().OptionBool("force").Return(true).Once()
				s.mockProcess.EXPECT().Run("go", "run", pkg+"/setup", "uninstall", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON, "--force").Return(mockSuccessResult(s.T())).Once()
				s.mockProcess.EXPECT().WithSpinner("Uninstalling " + pkg).Return(s.mockProcess).Once()
				failedResult := mockFailedResult(s.T(), assert.AnError)
				s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(failedResult).Once()
				s.mockContext.EXPECT().Error(fmt.Sprintf("failed to tidy go.mod file: %s", assert.AnError)).Once()
			},
		},
		{
			name: "package uninstall success(simulate)",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{pkg}).Once()
				s.mockContext.EXPECT().OptionBool("force").Return(true).Once()
				s.mockProcess.EXPECT().Run("go", "run", pkg+"/setup", "uninstall", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON, "--force").Return(mockSuccessResult(s.T())).Once()
				s.mockProcess.EXPECT().WithSpinner("Uninstalling " + pkg).Return(s.mockProcess).Once()
				s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockSuccessResult(s.T())).Once()
				s.mockContext.EXPECT().Success("Package " + pkg + " uninstalled successfully").Once()
			},
		},
		{
			name: "facade is a base facade",
			setup: func() {
				facade := "Config"
				s.mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				s.mockContext.EXPECT().Warning(fmt.Sprintf("Facade %s is a base facade, cannot be uninstalled", facade)).Once()
			},
		},
		{
			name: "facade is not found",
			setup: func() {
				facade := "unknown"
				s.mockContext.EXPECT().Arguments().Return([]string{facade}).Once()
				s.mockContext.EXPECT().Warning(errors.PackageFacadeNotFound.Args(facade).Error()).Once()
				s.mockContext.EXPECT().Info(fmt.Sprintf("Available facades: %s", strings.Join(getAvailableFacades(bindings), ", ")))
			},
		},
		{
			name: "facades can't be uninstalled because of existing upper dependencies",
			setup: func() {
				facade := "Orm"

				s.mockContext.EXPECT().Arguments().Return([]string{facade}).Once()

				s.NoError(file.PutContent("app/facades/auth.go", "package facades\n"))
				s.NoError(file.PutContent("app/facades/orm.go", "package facades\n"))

				s.mockContext.EXPECT().Error(fmt.Sprintf("Facade %s is depended on %s facades, cannot be uninstalled", facade, "Auth")).Once()
			},
		},
		{
			name: "facades uninstall failed",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{facade}).Once()

				s.NoError(file.PutContent("app/facades/auth.go", "package facades\n"))

				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				failedResult := mockFailedResult(s.T(), assert.AnError)
				s.mockProcess.EXPECT().Run("go", "run", bindings[binding.Auth].PkgPath+"/setup", "uninstall", "--facade=Auth", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON).Return(failedResult).Once()
				s.mockContext.EXPECT().Error(fmt.Sprintf("Failed to uninstall facade %s, error: %s", "Auth", assert.AnError)).Once()
			},
		},
		{
			name: "facades uninstall success(simulate)",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{facade}).Once()

				s.NoError(file.PutContent("app/facades/auth.go", "package facades\n"))

				s.mockContext.EXPECT().OptionBool("force").Return(true).Once()
				s.mockProcess.EXPECT().Run("go", "run", bindings[binding.Auth].PkgPath+"/setup", "uninstall", "--facade=Auth", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON, "--force").Return(mockSuccessResult(s.T())).Once()
				s.mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
				s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockSuccessResult(s.T())).Once()
			},
		},
		{
			name: "install package and facade simultaneously",
			setup: func() {
				s.mockContext.EXPECT().Arguments().Return([]string{pkg, facade}).Once()

				s.mockContext.EXPECT().OptionBool("force").Return(true).Once()
				s.mockProcess.EXPECT().Run("go", "run", pkg+"/setup", "uninstall", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON, "--force").Return(mockSuccessResult(s.T())).Once()
				s.mockProcess.EXPECT().WithSpinner("Uninstalling " + pkg).Return(s.mockProcess).Once()
				s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockSuccessResult(s.T())).Once()
				s.mockContext.EXPECT().Success("Package " + pkg + " uninstalled successfully").Once()

				s.NoError(file.PutContent("app/facades/auth.go", "package facades\n"))

				s.mockContext.EXPECT().OptionBool("force").Return(true).Once()
				s.mockProcess.EXPECT().Run("go", "run", bindings[binding.Auth].PkgPath+"/setup", "uninstall", "--facade=Auth", "--main-path=github.com/goravel/framework", "--paths="+pathsJSON, "--force").Return(mockSuccessResult(s.T())).Once()
				s.mockContext.EXPECT().Success("Facade Auth uninstalled successfully").Once()
				s.mockProcess.EXPECT().Run("go", "mod", "tidy").Return(mockSuccessResult(s.T())).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			beforeEach()
			test.setup()

			s.NoError(NewPackageUninstallCommand(bindings, s.mockProcess, s.mockJson).Handle(s.mockContext))

			s.NoError(file.Remove("app"))
		})
	}
}

func (s *PackageUninstallCommandTestSuite) TestGetExistingUpperDependencyFacades() {
	var (
		pathsJSON = `{"App":"app"}`
	)

	s.mockJson.EXPECT().MarshalString(mock.Anything).Return(pathsJSON, nil).Once()
	packageUninstallCommand := NewPackageUninstallCommand(binding.Bindings, s.mockProcess, s.mockJson)

	s.Run("upper dependencies exist", func() {
		s.NoError(file.PutContent("app/facades/auth.go", "package facades\n"))
		defer func() {
			s.NoError(file.Remove("app"))
		}()

		s.ElementsMatch([]string{"Auth"}, packageUninstallCommand.getExistingUpperDependencyFacades("Orm"))
	})

	s.Run("upper dependencies do not exist", func() {
		s.Empty(packageUninstallCommand.getExistingUpperDependencyFacades("Orm"))
	})
}
