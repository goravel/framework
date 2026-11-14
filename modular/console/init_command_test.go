package console

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
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

// mainFile has a build tag so the go tool ignores it in this package.
const mainFile = `//go:build ignore

package main

import "goravel/bootstrap"

func main() {
	bootstrap.Boot().Start()
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

func (s *InitCommandTestSuite) TestReturnsUnparsableApp() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", "package bootstrap\n\nvar app = foundation.Setup().Create(\n"))

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()

	err := NewInitCommand(nil).Handle(ctx)

	s.Error(err)
	s.False(isManualWiring(err))
	s.False(file.Exists("bootstrap/modules.go"))
}

func (s *InitCommandTestSuite) TestCreatesThenReportsExisting() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	s.Require().NoError(file.PutContent("main.go", mainFile))

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", statusCreated).Once()

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
	ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", statusExists).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	again, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(app, again)

	main, err := file.GetContent("main.go")
	s.NoError(err)
	s.Equal(strings.Replace(mainFile, "\nfunc main(", "\n"+generateDirective+"\n\nfunc main(", 1), main)
}

func (s *InitCommandTestSuite) TestGenerateDirective() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	rows := func(status string) *mocksconsole.Context {
		ctx := mocksconsole.NewContext(s.T())
		ctx.EXPECT().TwoColumnDetail(mock.Anything, mock.Anything).Times(4)
		ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", status).Once()

		return ctx
	}

	// No main.go: left to the user.
	ctx := rows(statusManual)
	ctx.EXPECT().Warning("Add " + generateDirective + " to the main package").Once()
	s.NoError(NewInitCommand(nil).Handle(ctx))
	s.False(file.Exists("main.go"))

	// No func main in main.go: appended.
	s.Require().NoError(file.PutContent("main.go", "//go:build ignore\n\npackage main\n"))
	s.NoError(NewInitCommand(nil).Handle(rows(statusCreated)))
	main, err := file.GetContent("main.go")
	s.NoError(err)
	s.Equal("//go:build ignore\n\npackage main\n\n"+generateDirective+"\n", main)

	s.NoError(NewInitCommand(nil).Handle(rows(statusExists)))
}

func (s *InitCommandTestSuite) TestAppFacadesAndConfiguredPath() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	s.Require().NoError(file.PutContent("main.go", mainFile))
	s.Require().NoError(file.PutContent("app/facades/config.go", "package facades\n"))

	config := mocksconfig.NewConfig(s.T())
	config.EXPECT().GetString("modular.path", "modules").Return("src/modules").Once()

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("src/modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", statusCreated).Once()

	s.NoError(NewInitCommand(config).Handle(ctx))

	s.True(file.Exists("src/modules/.gitkeep"))
	s.True(file.Contain("config/modular.go", `"github.com/goravel/framework/app/facades"`))
}

func (s *InitCommandTestSuite) TestAssignedChainIsManual() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", assignedApp))
	s.Require().NoError(file.PutContent("main.go", mainFile))

	warning := "Add .WithModules(Modules) to foundation.Setup() in bootstrap/app.go"
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusManual).Once()
	ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", statusCreated).Once()
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
	ctx.EXPECT().TwoColumnDetail("main.go go:generate modular:publish", statusExists).Once()
	ctx.EXPECT().Warning(warning).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))
}

// removeScaffold deletes what the commands write into the package directory.
func removeScaffold(t *testing.T) {
	t.Helper()

	for _, dir := range []string{"app", "bootstrap", "config", "modules", "src", "tests", "main.go"} {
		if err := file.Remove(dir); err != nil {
			t.Fatal(err)
		}
	}
}

func (s *InitCommandTestSuite) TestWriteFailures() {
	s.Run("the modules directory is a file", func() {
		s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
		s.Require().NoError(file.PutContent("modules", "file"))

		s.Error(NewInitCommand(nil).Handle(mocksconsole.NewContext(s.T())))
		s.False(file.Exists("config/modular.go"))
		removeScaffold(s.T())
	})

	s.Run("the config directory is a file", func() {
		s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
		s.Require().NoError(file.PutContent("config", "file"))

		ctx := mocksconsole.NewContext(s.T())
		ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()

		s.Error(NewInitCommand(nil).Handle(ctx))
		s.False(file.Exists("bootstrap/modules.go"))
		removeScaffold(s.T())
	})

	s.Run("main.go is read-only", func() {
		if os.Geteuid() == 0 {
			s.T().Skip("root writes read-only files")
		}
		s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
		s.Require().NoError(file.PutContent("main.go", mainFile))
		s.Require().NoError(os.Chmod("main.go", 0o444))
		s.T().Cleanup(func() { _ = os.Chmod("main.go", 0o644) })

		ctx := mocksconsole.NewContext(s.T())
		ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
		ctx.EXPECT().TwoColumnDetail("config/modular.go", statusCreated).Once()
		ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
		ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()

		s.Error(NewInitCommand(nil).Handle(ctx))
		s.False(file.Contain("main.go", generateDirective))
		s.Require().NoError(os.Chmod("main.go", 0o644))
	})
}
