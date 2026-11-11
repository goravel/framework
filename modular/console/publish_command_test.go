package console

import (
	"io/fs"
	"runtime/debug"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmodular "github.com/goravel/framework/mocks/modular"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type testModule struct{ name string }

func (m *testModule) Name() string { return m.name }

type sourcesModule struct {
	testModule
	sources fs.FS
}

func (m *sourcesModule) Sources() fs.FS { return m.sources }

type locatedModule struct{ sourcesModule }

func (m *locatedModule) Path() string { return "modules/local" }

// thirdPartyBuild makes this package's test modules third-party: the app is another Go module.
func thirdPartyBuild() (*debug.BuildInfo, bool) {
	return &debug.BuildInfo{
		Main: debug.Module{Path: "example.com/app"},
		Deps: []*debug.Module{{Path: "github.com/goravel/framework", Version: "v1.3.0", Sum: "h1:abc"}},
	}, true
}

type PublishCommandTestSuite struct {
	suite.Suite
}

func TestPublishCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PublishCommandTestSuite))
}

func (s *PublishCommandTestSuite) TearDownTest() {
	removeScaffold(s.T())
}

func (s *PublishCommandTestSuite) command(registry contractsmodular.Registry) *PublishCommand {
	cmd := NewPublishCommand(registry)
	cmd.buildInfo = thirdPartyBuild

	return cmd
}

func (s *PublishCommandTestSuite) TestSignature() {
	cmd := NewPublishCommand(nil)

	s.Equal("modular:publish", cmd.Signature())
	s.Equal("Copy the sources of third-party modules into the project", cmd.Description())
	s.Equal("modular", cmd.Extend().Category)
	s.Len(cmd.Extend().Flags, 1)
}

func (s *PublishCommandTestSuite) TestRegistryNotSet() {
	s.Equal(errors.ModularRegistryNotSet, NewPublishCommand(nil).Handle(mocksconsole.NewContext(s.T())))
}

func (s *PublishCommandTestSuite) TestPublishesThirdPartyModules() {
	billing := &sourcesModule{testModule{"billing"}, fstest.MapFS{"resources/js/app.js": {Data: []byte("app")}}}
	plain := &testModule{"plain"}
	local := &locatedModule{sourcesModule{testModule{"local"}, fstest.MapFS{"app.js": {Data: []byte("app")}}}}

	registry := mocksmodular.NewRegistry(s.T())
	registry.EXPECT().All().Return([]contractsmodular.Module{billing, plain, local})
	registry.EXPECT().Path("billing").Return("modules/billing", true)
	registry.EXPECT().Path("plain").Return("modules/plain", true)

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("").Once()
	ctx.EXPECT().TwoColumnDetail("billing", statusPublished).Once()
	ctx.EXPECT().TwoColumnDetail("plain", statusSkippedNone).Once()
	ctx.EXPECT().TwoColumnDetail("local", statusSkippedLocal).Once()

	s.NoError(s.command(registry).Handle(ctx))

	s.True(file.Exists("modules/billing/resources/js/app.js"))
	s.True(file.Contain("modules/billing/"+modular.PublishMarker, `"version": "v1.3.0"`))
	s.False(file.Exists("modules/plain"))
	s.False(file.Exists("modules/local"))
	s.Equal(modular.PublishPublished, modular.PublishState(billing, modular.Dependency{Path: "github.com/goravel/framework", Version: "v1.3.0", Sum: "h1:abc"}, path.Base("modules", "billing")))

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("").Once()
	ctx.EXPECT().TwoColumnDetail("billing", statusUnchanged).Once()
	ctx.EXPECT().TwoColumnDetail("plain", statusSkippedNone).Once()
	ctx.EXPECT().TwoColumnDetail("local", statusSkippedLocal).Once()

	s.NoError(s.command(registry).Handle(ctx))
}

func (s *PublishCommandTestSuite) TestModuleOption() {
	billing := &sourcesModule{testModule{"billing"}, fstest.MapFS{"app.js": {Data: []byte("app")}}}

	registry := mocksmodular.NewRegistry(s.T())
	registry.EXPECT().All().Return([]contractsmodular.Module{billing, &testModule{"plain"}})
	registry.EXPECT().Find("billing").Return(billing, true).Once()
	registry.EXPECT().Find("shop").Return(nil, false).Once()
	registry.EXPECT().Path("billing").Return("modules/billing", true)

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("billing").Once()
	ctx.EXPECT().TwoColumnDetail("billing", statusPublished).Once()

	s.NoError(s.command(registry).Handle(ctx))

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("shop").Once()

	s.Equal(errors.ModularModuleNotFound.Args("shop"), s.command(registry).Handle(ctx))
}

func (s *PublishCommandTestSuite) TestFailureDoesNotStopOtherModules() {
	billing := &sourcesModule{testModule{"billing"}, fstest.MapFS{"app.js": {Data: []byte("app")}}}
	shop := &sourcesModule{testModule{"shop"}, fstest.MapFS{"app.js": {Data: []byte("app")}}}
	s.Require().NoError(file.PutContent("modules/billing/module.go", "package billing\n"))

	registry := mocksmodular.NewRegistry(s.T())
	registry.EXPECT().All().Return([]contractsmodular.Module{billing, shop})
	registry.EXPECT().Path("billing").Return("modules/billing", true)
	registry.EXPECT().Path("shop").Return("modules/shop", true)

	notOwned := errors.ModularPublishTargetNotOwned.Args("billing", "modules/billing", "github.com/goravel/framework")
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("").Once()
	ctx.EXPECT().Error(notOwned.Error()).Once()
	ctx.EXPECT().TwoColumnDetail("billing", statusFailed).Once()
	ctx.EXPECT().TwoColumnDetail("shop", statusPublished).Once()

	s.Equal(errors.ModularPublishFailed.Args(1, 2), s.command(registry).Handle(ctx))
	s.False(file.Exists("modules/billing/app.js"))
	s.True(file.Exists("modules/shop/app.js"))
}

func (s *PublishCommandTestSuite) TestWriteFailure() {
	billing := &sourcesModule{testModule{"billing"}, fstest.MapFS{"resources/js/app.js": {Data: []byte("app")}}}
	// An earlier publish owns the directory, but a file is where resources/ goes.
	s.Require().NoError(file.PutContent("modules/billing/"+modular.PublishMarker, `{"module": "github.com/goravel/framework"}`))
	s.Require().NoError(file.PutContent("modules/billing/resources", "file"))

	registry := mocksmodular.NewRegistry(s.T())
	registry.EXPECT().All().Return([]contractsmodular.Module{billing})
	registry.EXPECT().Path("billing").Return("modules/billing", true)

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Option("module").Return("").Once()
	ctx.EXPECT().Error(mock.Anything).Once()
	ctx.EXPECT().TwoColumnDetail("billing", statusFailed).Once()

	s.Equal(errors.ModularPublishFailed.Args(1, 1), s.command(registry).Handle(ctx))
}
