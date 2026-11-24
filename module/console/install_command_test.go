package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/file"
)

type InstallCommandTestSuite struct {
	suite.Suite
	process *mocksprocess.Process
}

func TestInstallCommandTestSuite(t *testing.T) {
	suite.Run(t, new(InstallCommandTestSuite))
}

func (s *InstallCommandTestSuite) SetupTest() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	s.process = mocksprocess.NewProcess(s.T())
}

func (s *InstallCommandTestSuite) TearDownTest() {
	removeScaffold(s.T())
}

func (s *InstallCommandTestSuite) context(packages ...string) *mocksconsole.Context {
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Arguments().Return(packages).Once()

	return ctx
}

// result is a finished process: failed with err, or successful with output.
func (s *InstallCommandTestSuite) result(output string, err error) *mocksprocess.Result {
	result := mocksprocess.NewResult(s.T())
	result.EXPECT().Failed().Return(err != nil).Once()
	if err != nil {
		result.EXPECT().Error().Return(err).Once()
	}
	if output != "" {
		result.EXPECT().Output().Return(output).Once()
	}

	return result
}

// expectGet expects go get and go list for pkg, answering name as its Go package.
func (s *InstallCommandTestSuite) expectGet(pkg, importPath, name string) {
	spinner := mocksprocess.NewProcess(s.T())
	s.process.EXPECT().WithSpinner("Getting " + pkg).Return(spinner).Once()
	spinner.EXPECT().Run("go", "get", pkg).Return(s.result("", nil)).Once()
	s.process.EXPECT().Run("go", "list", "-f", "{{.Name}}", importPath).Return(s.result(name+"\n", nil)).Once()
}

// expectPublish expects go mod tidy and the publish run printing output.
func (s *InstallCommandTestSuite) expectPublish(output string, err error) {
	s.process.EXPECT().Run("go", "mod", "tidy").Return(s.result("", nil)).Once()

	result := mocksprocess.NewResult(s.T())
	result.EXPECT().Output().Return(output).Once()
	result.EXPECT().Failed().Return(err != nil).Once()
	if err != nil {
		result.EXPECT().Error().Return(err).Once()
	}
	spinner := mocksprocess.NewProcess(s.T())
	s.process.EXPECT().WithSpinner("Publishing modules").Return(spinner).Once()
	spinner.EXPECT().Run("go", "run", ".", "artisan", "module:publish").Return(result).Once()
}

func (s *InstallCommandTestSuite) TestInstall() {
	s.expectGet("github.com/acme/billing@v1.2.0", "github.com/acme/billing", "billing")
	s.expectGet("github.com/acme/go-shop/v2", "github.com/acme/go-shop/v2", "shop")
	s.expectPublish("  billing ... PUBLISHED\n", nil)

	ctx := s.context("github.com/acme/billing@v1.2.0", "github.com/acme/go-shop/v2")
	ctx.EXPECT().Success("Module installed successfully: github.com/acme/billing@v1.2.0").Once()
	ctx.EXPECT().Success("Module installed successfully: github.com/acme/go-shop/v2").Once()
	ctx.EXPECT().Line("billing ... PUBLISHED").Once()

	s.NoError(NewInstallCommand(s.process).Handle(ctx))

	s.True(file.Contain("bootstrap/app.go", "WithModules(Modules)"))
	modules, err := file.GetContent("bootstrap/modules.go")
	s.NoError(err)
	s.Contains(modules, "\t\"github.com/acme/billing\"\n")
	s.Contains(modules, "\tshop \"github.com/acme/go-shop/v2\"\n")
	s.Contains(modules, "&billing.Module{}")
	s.Contains(modules, "&shop.Module{}")

	// Again: listed once, published again.
	s.expectGet("github.com/acme/billing", "github.com/acme/billing", "billing")
	s.expectPublish("", nil)
	ctx = s.context("github.com/acme/billing")
	ctx.EXPECT().Success("Module installed successfully: github.com/acme/billing").Once()

	s.NoError(NewInstallCommand(s.process).Handle(ctx))

	again, err := file.GetContent("bootstrap/modules.go")
	s.NoError(err)
	s.Equal(modules, again)
}

func (s *InstallCommandTestSuite) TestPackageNameTaken() {
	s.expectGet("github.com/acme/billing", "github.com/acme/billing", "billing")
	s.expectPublish("", nil)
	ctx := s.context("github.com/acme/billing")
	ctx.EXPECT().Success("Module installed successfully: github.com/acme/billing").Once()
	s.NoError(NewInstallCommand(s.process).Handle(ctx))

	s.expectGet("github.com/other/billing", "github.com/other/billing", "billing")

	err := NewInstallCommand(s.process).Handle(s.context("github.com/other/billing"))

	s.EqualError(err, errors.ModuleImportConflict.Args("github.com/other/billing", "billing", "github.com/acme/billing").Error())
}

func (s *InstallCommandTestSuite) TestFailures() {
	s.Run("no package", func() {
		s.Equal(errors.ModuleInstallPackageEmpty, NewInstallCommand(s.process).Handle(s.context()))
	})

	s.Run("go get", func() {
		spinner := mocksprocess.NewProcess(s.T())
		s.process.EXPECT().WithSpinner("Getting example.com/x").Return(spinner).Once()
		spinner.EXPECT().Run("go", "get", "example.com/x").Return(s.result("", assert.AnError)).Once()

		err := NewInstallCommand(s.process).Handle(s.context("example.com/x"))

		s.EqualError(err, errors.ModuleInstallFailed.Args("go get example.com/x", assert.AnError).Error())
		s.False(file.Exists("bootstrap/modules.go"))
	})

	s.Run("go list", func() {
		spinner := mocksprocess.NewProcess(s.T())
		s.process.EXPECT().WithSpinner("Getting example.com/x").Return(spinner).Once()
		spinner.EXPECT().Run("go", "get", "example.com/x").Return(s.result("", nil)).Once()
		s.process.EXPECT().Run("go", "list", "-f", "{{.Name}}", "example.com/x").Return(s.result("", assert.AnError)).Once()

		err := NewInstallCommand(s.process).Handle(s.context("example.com/x"))

		s.EqualError(err, errors.ModuleInstallFailed.Args("go list example.com/x", assert.AnError).Error())
	})

	s.Run("publish", func() {
		s.expectGet("example.com/x", "example.com/x", "x")
		s.expectPublish("  x ... FAILED", assert.AnError)
		ctx := s.context("example.com/x")
		ctx.EXPECT().Success("Module installed successfully: example.com/x").Once()
		ctx.EXPECT().Line("x ... FAILED").Once()

		err := NewInstallCommand(s.process).Handle(ctx)

		s.EqualError(err, errors.ModuleInstallFailed.Args("module:publish", assert.AnError).Error())
	})
}

func (s *InstallCommandTestSuite) TestRequiresBootstrapSetup() {
	s.Require().NoError(file.Remove("bootstrap"))

	err := NewInstallCommand(s.process).Handle(mocksconsole.NewContext(s.T()))

	s.EqualError(err, errors.ModuleRequiresBootstrapSetup.Args("module:install").Error())
}
