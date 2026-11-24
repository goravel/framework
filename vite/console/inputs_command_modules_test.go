package console

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksmodule "github.com/goravel/framework/mocks/module"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/module"
	"github.com/goravel/framework/support"
)

type testModule struct{ name string }

func (m *testModule) Name() string { return m.name }

type viteModule struct {
	testModule
	entries []string
	aliases map[string]string
}

func (m *viteModule) Frontend() contractsmodule.Frontend {
	return contractsmodule.Frontend{Entries: m.entries, Aliases: m.aliases}
}

type sourcesModule struct {
	testModule
	sources fs.FS
	entries []string
	dir     string
}

func (m *sourcesModule) Frontend() contractsmodule.Frontend {
	return contractsmodule.Frontend{Files: contractsmodule.Files{FS: m.sources, Dir: m.dir}, Entries: m.entries}
}

// localBuild makes this package's test modules part of the application.
func localBuild() (*debug.BuildInfo, bool) {
	return &debug.BuildInfo{Main: debug.Module{Path: "github.com/goravel/framework"}}, true
}

// thirdPartyBuild makes them third-party, linked at version.
func thirdPartyBuild(version string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Path: "example.com/app"},
			Deps: []*debug.Module{{Path: "github.com/goravel/framework", Version: version, Sum: "h1:abc"}},
		}, true
	}
}

// moduleRegistry answers for modules placed at modules/<name>.
func moduleRegistry(t *testing.T, modules ...contractsmodule.Module) *mocksmodule.Registry {
	t.Helper()

	registry := mocksmodule.NewRegistry(t)
	registry.EXPECT().All().Return(modules)
	for _, module := range modules {
		registry.EXPECT().Path(module.Name()).Return("modules/"+module.Name(), true).Maybe()
	}

	return registry
}

// moduleCommand builds a command over the given views; compiled is false when the
// command is expected to fail before it compiles them.
func moduleCommand(t *testing.T, registry contractsmodule.Registry, views fs.FS, compiled bool) *InputsCommand {
	t.Helper()

	view := mocksview.NewView(t)
	vite := defaultVite(t)
	if compiled {
		var registered []fs.FS
		if views != nil {
			registered = []fs.FS{views}
		}
		view.EXPECT().Funcs().Return(template.FuncMap{"vite": func(...string) template.HTML { return "" }}).Once()
		view.EXPECT().RegisteredViews().Return(nil).Once()
		view.EXPECT().RegisteredViewFS().Return(registered).Once()
	} else {
		vite.ExpectedCalls = nil
	}

	cmd := NewInputsCommand(view, vite, nil, registry)
	cmd.buildInfo = localBuild

	return cmd
}

func TestInputsCommand_Modules(t *testing.T) {
	setupAppViews(t, nil)

	blog := &viteModule{
		testModule: testModule{"blog"},
		entries:    []string{"resources/js/app.js", "./resources/css/app.css"},
		aliases:    map[string]string{"@blog-assets": "resources/assets", "@blog": "./resources"},
	}
	shop := &testModule{"shop"}
	views := fstest.MapFS{"blog/index.tmpl": {Data: []byte(`{{ vite "@blog/app.js" }}`)}}

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":["@blog/app.js"],"assets":[],"dynamic":[],"paths":{"public_directory":"public","build_directory":"build","hot_file":"public/hot","manifest":".vite/manifest.json"},` +
		`"modules":[{"name":"blog","path":"modules/blog","entries":["modules/blog/resources/js/app.js","modules/blog/resources/css/app.css"],"aliases":{"@blog":"modules/blog/resources","@blog-assets":"modules/blog/resources/assets"}},` +
		`{"name":"shop","path":"modules/shop","entries":[],"aliases":{}}]}`).Once()

	assert.NoError(t, moduleCommand(t, moduleRegistry(t, blog, shop), views, true).Handle(ctx))
}

func TestInputsCommand_ModuleAliasErrors(t *testing.T) {
	setupAppViews(t, nil)

	module := func(name string, aliases map[string]string) *viteModule {
		return &viteModule{testModule: testModule{name}, aliases: aliases}
	}

	tests := []struct {
		name    string
		modules []contractsmodule.Module
		err     error
	}{
		{"escaping target", []contractsmodule.Module{module("blog", map[string]string{"@x": "../shop"})}, errors.ModuleViteAliasInvalid.Args("blog", "@x", "../shop")},
		{"absolute target", []contractsmodule.Module{module("blog", map[string]string{"@x": "/etc"})}, errors.ModuleViteAliasInvalid.Args("blog", "@x", "/etc")},
		{"bare @", []contractsmodule.Module{module("blog", map[string]string{"@": "."})}, errors.ModuleViteAliasInvalid.Args("blog", "@", ".")},
		{"no @", []contractsmodule.Module{module("blog", map[string]string{"bad": "."})}, errors.ModuleViteAliasInvalid.Args("blog", "bad", ".")},
		{"slash in key", []contractsmodule.Module{module("blog", map[string]string{"@a/b": "."})}, errors.ModuleViteAliasInvalid.Args("blog", "@a/b", ".")},
		{
			"two modules, one key",
			[]contractsmodule.Module{module("blog", map[string]string{"@ui": "ui"}), module("shop", map[string]string{"@ui": "ui"})},
			errors.ModuleViteAliasDuplicate.Args("@ui", "shop", "blog"),
		},
		{
			"another module's implicit alias",
			[]contractsmodule.Module{module("blog", map[string]string{"@shop": "resources"}), module("shop", nil)},
			errors.ModuleViteAliasDuplicate.Args("@shop", "blog", "shop"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := moduleCommand(t, moduleRegistry(t, test.modules...), nil, false)

			assert.EqualError(t, cmd.Handle(mocksconsole.NewContext(t)), test.err.Error())
		})
	}
}

func TestInputsCommand_ModuleEntryErrors(t *testing.T) {
	setupAppViews(t, nil)

	for _, entry := range []string{"../shop/app.js", "resources/../../shop/app.js", "/etc/app.js", "", "."} {
		t.Run(entry, func(t *testing.T) {
			blog := &viteModule{testModule: testModule{"blog"}, entries: []string{"resources/js/app.js", entry}}
			cmd := moduleCommand(t, moduleRegistry(t, blog), nil, false)

			assert.EqualError(t, cmd.Handle(mocksconsole.NewContext(t)), errors.ModuleViteEntryInvalid.Args("blog", entry).Error())
		})
	}
}

func TestInputsCommand_ModulePublished(t *testing.T) {
	setupAppViews(t, nil)

	billing := &sourcesModule{testModule: testModule{"billing"}, sources: fstest.MapFS{"resources/js/app.js": {Data: []byte("app")}}}
	plain := &testModule{"plain"}
	dir := filepath.Join(support.RelativePath, "modules", "billing")
	dependency := func(version string) module.Dependency {
		return module.Dependency{Path: "github.com/goravel/framework", Version: version, Sum: "h1:abc"}
	}
	run := func(version string, compiled bool) error {
		ctx := mocksconsole.NewContext(t)
		if compiled {
			ctx.EXPECT().Line(`{"inputs":[],"assets":[],"dynamic":[],"paths":{"public_directory":"public","build_directory":"build","hot_file":"public/hot","manifest":".vite/manifest.json"},` +
				`"modules":[{"name":"plain","path":"modules/plain","entries":[],"aliases":{}},{"name":"billing","path":"modules/billing","entries":[],"aliases":{}}]}`).Once()
		}
		cmd := moduleCommand(t, moduleRegistry(t, plain, billing), nil, compiled)
		cmd.buildInfo = thirdPartyBuild(version)

		return cmd.Handle(ctx)
	}

	// Not published yet; a module without sources is never checked.
	assert.EqualError(t, run("v1.2.0", false), errors.ModuleNotPublished.Args("billing", "github.com/goravel/framework v1.2.0", "modules/billing").Error())

	_, err := module.Publish(billing, dependency("v1.2.0"), dir)
	require.NoError(t, err)
	assert.NoError(t, run("v1.2.0", true))

	// Another version is linked than the one published.
	assert.EqualError(t, run("v1.3.0", false), errors.ModulePublishStale.Args("billing", "modules/billing", "github.com/goravel/framework v1.3.0").Error())

	// The directory holds something module:publish did not write.
	require.NoError(t, os.Remove(filepath.Join(dir, module.PublishMarker)))
	assert.EqualError(t, run("v1.2.0", false), errors.ModulePublishTargetNotOwned.Args("billing", "modules/billing", "github.com/goravel/framework").Error())

	// In the application itself nothing is published.
	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(mock.Anything).Once()
	assert.NoError(t, moduleCommand(t, moduleRegistry(t, billing), nil, true).Handle(ctx))
}

func TestInputsCommand_ModuleFrontendFiles(t *testing.T) {
	setupAppViews(t, nil)

	sources := fstest.MapFS{"web/js/app.js": {Data: []byte("app")}}
	run := func(module contractsmodule.Module, build func() (*debug.BuildInfo, bool), line string) error {
		ctx := mocksconsole.NewContext(t)
		if line != "" {
			ctx.EXPECT().Line(mock.MatchedBy(func(output string) bool { return strings.Contains(output, line) })).Once()
		}
		cmd := moduleCommand(t, moduleRegistry(t, module), nil, line != "")
		cmd.buildInfo = build

		return cmd.Handle(ctx)
	}

	t.Run("entries and aliases are relative to Files.Dir", func(t *testing.T) {
		blog := &sourcesModule{testModule: testModule{"blog"}, sources: sources, dir: "web", entries: []string{"js/app.js"}}

		assert.NoError(t, run(blog, localBuild, `"entries":["modules/blog/web/js/app.js"]`))
	})

	t.Run("an entry must be a file of an embedded Files", func(t *testing.T) {
		for _, entry := range []string{"js/missing.js", "js"} {
			blog := &sourcesModule{testModule: testModule{"blog"}, sources: sources, dir: "web", entries: []string{entry}}

			assert.EqualError(t, run(blog, localBuild, ""), errors.ModuleViteEntryMissing.Args("blog", entry).Error())
		}
	})

	t.Run("a directory outside the module", func(t *testing.T) {
		blog := &sourcesModule{testModule: testModule{"blog"}, dir: "../shop"}

		assert.EqualError(t, run(blog, localBuild, ""), errors.ModuleFilesDirInvalid.Args("blog", "Frontend()", "../shop").Error())
	})

	t.Run("a third-party module without an FS is published from its package directory", func(t *testing.T) {
		const pkg = "github.com/goravel/framework/vite/console"
		packageDir := module.PackageDir
		t.Cleanup(func() { module.PackageDir = packageDir })

		source := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(source, "resources", "js"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(source, "resources", "js", "app.js"), []byte("app"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(source, "app.js"), []byte("outside resources"), 0o644))
		module.PackageDir = func(given string) (string, error) {
			assert.Equal(t, pkg, given)

			return source, nil
		}

		billing := &viteModule{testModule: testModule{"billing"}, entries: []string{"resources/js/app.js"}}
		build := thirdPartyBuild("v1.2.0")
		assert.EqualError(t, run(billing, build, ""), errors.ModuleNotPublished.Args("billing", "github.com/goravel/framework v1.2.0", "modules/billing").Error())

		dependency := module.Dependency{Path: "github.com/goravel/framework", Version: "v1.2.0", Sum: "h1:abc"}
		_, err := module.Publish(billing, dependency, filepath.Join(support.RelativePath, "modules", "billing"))
		require.NoError(t, err)
		assert.NoError(t, run(billing, build, `"entries":["modules/billing/resources/js/app.js"]`))

		// Only resources/ is published by default.
		for _, entry := range []string{"resources/js/missing.js", "app.js", "resources"} {
			missing := &viteModule{testModule: testModule{"billing"}, entries: []string{"resources/js/app.js", entry}}
			assert.EqualError(t, run(missing, build, ""), errors.ModuleViteEntryMissing.Args("billing", entry).Error())
		}

		// Without frontend nothing is located or published.
		module.PackageDir = func(string) (string, error) { return "", assert.AnError }
		assert.NoError(t, run(&viteModule{testModule: testModule{"plain"}}, build, `"entries":[]`))
		assert.EqualError(t, run(&viteModule{testModule: testModule{"shop"}, entries: []string{"app.js"}}, build, ""),
			errors.ModuleFrontendSourcesNotFound.Args("shop", pkg, assert.AnError).Error())
	})
}
