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

	s.Equal("module:init", cmd.Signature())
	s.Equal("Prepare the application for modules", cmd.Description())
	s.Equal("module", cmd.Extend().Category)
}

func (s *InitCommandTestSuite) TestRequiresBootstrapSetup() {
	err := NewInitCommand(nil).Handle(mocksconsole.NewContext(s.T()))

	s.Equal(errors.ModuleRequiresBootstrapSetup.Args("module:init"), err)
	s.False(file.Exists("modules"))
}

func (s *InitCommandTestSuite) TestReturnsUnparsableApp() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", "package bootstrap\n\nvar app = foundation.Setup().Create(\n"))

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()

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
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	s.True(file.Exists("modules/.gitkeep"))
	s.True(file.Contain("bootstrap/app.go", "WithModules(Modules)"))
	modules, err := file.GetContent("bootstrap/modules.go")
	s.NoError(err)
	s.Equal(modify.ModulesStub(), modules)
	s.False(file.Exists("config/module.go"))

	app, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusExists).Once()

	s.NoError(NewInitCommand(nil).Handle(ctx))

	again, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(app, again)

	main, err := file.GetContent("main.go")
	s.NoError(err)
	s.Equal(mainFile, main, "main.go is left alone")
}

func (s *InitCommandTestSuite) TestConfiguredPath() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	s.Require().NoError(file.PutContent("main.go", mainFile))

	config := mocksconfig.NewConfig(s.T())
	config.EXPECT().GetString("module.path", "modules").Return("src/modules").Once()

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("src/modules/.gitkeep", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusCreated).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusCreated).Once()

	s.NoError(NewInitCommand(config).Handle(ctx))

	s.True(file.Exists("src/modules/.gitkeep"))
	s.False(file.Exists("config/module.go"))
}

func (s *InitCommandTestSuite) TestAssignedChainIsManual() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", assignedApp))
	s.Require().NoError(file.PutContent("main.go", mainFile))

	warning := "Add .WithModules(Modules) to foundation.Setup() in bootstrap/app.go"
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().TwoColumnDetail("modules/.gitkeep", statusCreated).Once()
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
	ctx.EXPECT().TwoColumnDetail("bootstrap/modules.go", statusExists).Once()
	ctx.EXPECT().TwoColumnDetail("bootstrap/app.go WithModules(Modules)", statusManual).Once()
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
		s.False(file.Exists("bootstrap/modules.go"))
		removeScaffold(s.T())
	})
}
