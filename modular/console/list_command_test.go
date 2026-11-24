package console

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/suite"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmodule "github.com/goravel/framework/mocks/module"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/file"
)

type richModule struct{ testModule }

func (m *richModule) Dependencies() []string { return []string{"billing"} }
func (m *richModule) Frontend() contractsmodule.Frontend {
	return contractsmodule.Frontend{
		Entries: []string{"resources/js/app.js"},
		Aliases: map[string]string{"@blog-assets": "resources/assets"},
	}
}
func (m *richModule) Static() map[string]contractsmodule.Files {
	return map[string]contractsmodule.Files{"/blog/static": {Dir: "public"}, "/blog/docs": {Dir: "docs"}}
}

type ListCommandTestSuite struct {
	suite.Suite
}

func TestListCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListCommandTestSuite))
}

func (s *ListCommandTestSuite) TearDownTest() {
	removeScaffold(s.T())
}

func (s *ListCommandTestSuite) TestSignature() {
	cmd := NewListCommand(nil)

	s.Equal("module:list", cmd.Signature())
	s.Equal("List the registered modules", cmd.Description())
	s.Equal("module", cmd.Extend().Category)
	s.Len(cmd.Extend().Flags, 1)
}

func (s *ListCommandTestSuite) TestRegistryNotSet() {
	s.Equal(errors.ModularRegistryNotSet, NewListCommand(nil).Handle(mocksconsole.NewContext(s.T())))
}

func (s *ListCommandTestSuite) TestEmpty() {
	registry := mocksmodule.NewRegistry(s.T())
	registry.EXPECT().All().Return(nil)

	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(false).Once()
	ctx.EXPECT().Warning("No modules are registered. List them in bootstrap/modules.go.").Once()
	s.NoError(NewListCommand(registry).Handle(ctx))

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(true).Once()
	var output bytes.Buffer
	cmd := NewListCommand(registry)
	cmd.out = &output
	s.NoError(cmd.Handle(ctx))
	s.Equal("[]\n", output.String())
}

func (s *ListCommandTestSuite) TestList() {
	blog := &richModule{testModule{"blog"}}
	billing := &sourcesModule{testModule{"billing"}, fstest.MapFS{"app.js": {Data: []byte("app")}}}

	registry := mocksmodule.NewRegistry(s.T())
	registry.EXPECT().All().Return([]contractsmodule.Module{blog, billing})
	registry.EXPECT().Path("blog").Return("modules/blog", true)
	registry.EXPECT().Path("billing").Return("modules/billing", true)
	registry.EXPECT().Capabilities(blog).Return(map[string]bool{"frontend": true, "static": true, "views": false})
	registry.EXPECT().Capabilities(billing).Return(map[string]bool{"frontend": true})

	// Both are modules of this application: nothing is published.
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(false).Once()
	ctx.EXPECT().TwoColumnDetail("<fg=green;op=bold>blog</> modules/blog", "local").Once()
	ctx.EXPECT().Line("    <fg=gray>frontend, static</>").Once()
	ctx.EXPECT().Line("    <fg=gray>depends on: billing</>").Once()
	ctx.EXPECT().Line("    <fg=gray>vite entries: resources/js/app.js</>").Once()
	ctx.EXPECT().Line("    <fg=gray>vite aliases: @blog-assets -> resources/assets</>").Once()
	ctx.EXPECT().Line("    <fg=gray>static: /blog/docs, /blog/static</>").Once()
	ctx.EXPECT().TwoColumnDetail("<fg=green;op=bold>billing</> modules/billing", "local").Once()
	ctx.EXPECT().Line("    <fg=gray>frontend</>").Once()

	s.NoError(NewListCommand(registry).Handle(ctx))

	// As third-party modules: billing has sources that are not published yet.
	var output bytes.Buffer
	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(true).Once()

	cmd := NewListCommand(registry)
	cmd.buildInfo = thirdPartyBuild
	cmd.out = &output
	s.NoError(cmd.Handle(ctx))
	s.JSONEq(`[
		{
			"name": "blog", "path": "modules/blog", "dependencies": ["billing"], "publish": "none",
			"capabilities": {"frontend": true, "static": true, "views": false},
			"vite_entries": ["resources/js/app.js"],
			"vite_aliases": {"@blog-assets": "resources/assets"},
			"static": ["/blog/docs", "/blog/static"]
		},
		{
			"name": "billing", "path": "modules/billing", "dependencies": [], "publish": "missing",
			"capabilities": {"frontend": true},
			"vite_entries": [], "vite_aliases": {}, "static": []
		}
	]`, output.String())

	// Published by module:publish.
	publishCtx := mocksconsole.NewContext(s.T())
	publishCtx.EXPECT().Option("module").Return("billing").Once()
	publishCtx.EXPECT().TwoColumnDetail("billing", statusPublished).Once()
	registry.EXPECT().Find("billing").Return(billing, true).Once()
	publish := NewPublishCommand(registry)
	publish.buildInfo = thirdPartyBuild
	s.NoError(publish.Handle(publishCtx))
	s.True(file.Exists("modules/billing/app.js"))

	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(true).Once()
	output.Reset()
	s.NoError(cmd.Handle(ctx))
	s.Contains(output.String(), `"name":"billing","path":"modules/billing","dependencies":[],"publish":"published"`)

	// Something else in the directory of a third-party module is a conflict.
	s.Require().NoError(file.Remove("modules/billing/" + modular.PublishMarker))
	ctx = mocksconsole.NewContext(s.T())
	ctx.EXPECT().OptionBool("json").Return(true).Once()
	output.Reset()
	s.NoError(cmd.Handle(ctx))
	s.Contains(output.String(), `"name":"billing","path":"modules/billing","dependencies":[],"publish":"conflict"`)
}
