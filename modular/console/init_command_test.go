package console

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
)

const chainedApp = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Create()
}
`

const assignedApp = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	builder := foundation.Setup().WithConfig(nil)

	return builder.Create()
}
`

type InitCommandTestSuite struct {
	suite.Suite
}

func TestInitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(InitCommandTestSuite))
}

func (s *InitCommandTestSuite) TearDownTest() {
	removeScaffold(s.T())
}

func (s *InitCommandTestSuite) TestSignature() {
	cmd := NewInitCommand(nil)

	s.Equal("modular:init", cmd.Signature())
	s.Equal("Prepare the application for modules", cmd.Description())
	s.Equal("modular", cmd.Extend().Category)
}

func (s *InitCommandTestSuite) TestRequiresBootstrapSetup() {
	err := NewInitCommand(nil).Handle(mocksconsole.NewContext(s.T()))

	s.Equal(errors.ModularRequiresBootstrapSetup.Args("modular:init"), err)
	s.False(file.Exists("modules"))
}

func (s *InitCommandTestSuite) TestCreatesThenReportsExisting() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	s.True(file.Exists("modules/.gitkeep"))
	s.True(file.Contain("bootstrap/app.go", "WithModules(Modules)"))
	modules, err := file.GetContent("bootstrap/modules.go")
	s.NoError(err)
	s.Equal(modify.ModulesStub(), modules)

	// No app/facades here, so the config file uses the framework facades.
	config, err := file.GetContent("config/modular.go")
	s.NoError(err)
	s.Contains(config, "package config")
	s.Contains(config, `"github.com/goravel/framework/facades"`)
	s.Contains(config, `facades.Config()`)
	s.Contains(config, `config.Add("modular", map[string]any{`)

	app, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusExists).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	again, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(app, again)
}

func (s *InitCommandTestSuite) TestAppFacadesAndConfiguredPath() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	s.Require().NoError(file.PutContent("app/facades/config.go", "package facades\n"))

	config := mocksconfig.NewConfig(s.T())
	config.EXPECT().GetString("modular.path", "modules").Return("src/modules").Once()

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("src/modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()

	s.NoError(NewInitCommand(config).Handle(ctx))

	s.True(file.Exists("src/modules/.gitkeep"))
	s.True(file.Contain("config/modular.go", `"github.com/goravel/framework/app/facades"`))
}

func (s *InitCommandTestSuite) TestAssignedChainIsManual() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", assignedApp))

	warning := "Add .WithModules(Modules) to foundation.Setup() in bootstrap/app.go"
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusManual).Once()
	ctx.EXPECT().Warning(warning).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	s.True(file.Exists("bootstrap/modules.go"))
	app, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(assignedApp, app)

	// modules.go without WithModules: the same manual row again.
	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusManual).Once()
	ctx.EXPECT().Warning(warning).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))
}

// removeScaffold deletes what the commands write into the package directory.
func removeScaffold(t *testing.T) {
	t.Helper()

	for _, dir := range []string{"app", "bootstrap", "config", "modules", "src", "tests"} {
		if err := file.Remove(dir); err != nil {
			t.Fatal(err)
		}
	}
}
