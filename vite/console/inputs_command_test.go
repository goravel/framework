package console

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksview "github.com/goravel/framework/mocks/view"
	mocksvite "github.com/goravel/framework/mocks/vite"
	"github.com/goravel/framework/support"
)

// setupAppViews points the project root at a temporary directory holding the given app views.
func setupAppViews(t *testing.T, files map[string]string) {
	t.Helper()

	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	for name, content := range files {
		path := filepath.Join(support.RelativePath, support.Config.Paths.Resources, "views", name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
}

// defaultVite answers like a Vite left on its defaults, with the hot file inside the project root.
func defaultVite(t *testing.T) *mocksvite.Vite {
	t.Helper()

	vite := mocksvite.NewVite(t)
	vite.EXPECT().HotFile().Return(filepath.Join(support.RelativePath, "public", "hot")).Once()
	vite.EXPECT().BuildDirectory().Return("build").Once()
	vite.EXPECT().ManifestFilename().Return(".vite/manifest.json").Once()

	return vite
}

const defaultPaths = `"paths":{"public_directory":"public","build_directory":"build","hot_file":"public/hot","manifest":".vite/manifest.json"},"modules":[]`

func TestInputsCommand_Metadata(t *testing.T) {
	cmd := NewInputsCommand(nil, nil, nil, nil)
	assert.Equal(t, "vite:inputs", cmd.Signature())
	assert.Equal(t, "Print the Vite entry points and assets the views and config declare, and the build paths, as JSON", cmd.Description())
	assert.Equal(t, "vite", cmd.Extend().Category)
}

func TestInputsCommand_Handle(t *testing.T) {
	setupAppViews(t, map[string]string{
		"home.tmpl": `{{ vite "resources/css/app.css" "resources/js/app.js" }}<img src="{{ viteAsset "resources/images/logo.svg" }}">`,
		"dyn.tmpl":  `{{ vite .entry }}{{ vite "resources/js/mixed.js" .extra }}`,
	})
	pkg := fstest.MapFS{"admin.tmpl": {Data: []byte(`{{ vite "resources/js/admin.js" "resources/js/app.js" }}`)}}

	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(template.FuncMap{
		"vite":      func(...string) template.HTML { return "" },
		"viteAsset": func(string) string { return "" },
	}).Once()
	view.EXPECT().RegisteredViews().Return(nil).Once()
	view.EXPECT().RegisteredViewFS().Return([]fs.FS{pkg}).Once()

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetStringSlice("vite.inputs", []string{}).Return([]string{"resources/inertia/app.ts", "resources/js/app.js"}).Once()
	config.EXPECT().GetStringSlice("vite.assets", []string{}).Return([]string{"resources/images/logo.svg"}).Once()

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":["resources/css/app.css","resources/inertia/app.ts","resources/js/admin.js","resources/js/app.js","resources/js/mixed.js"],"assets":["resources/images/logo.svg"],"dynamic":[{"template":"dyn.tmpl","function":"vite"},{"template":"dyn.tmpl","function":"vite"}],` + defaultPaths + `}`).Once()

	assert.NoError(t, NewInputsCommand(view, defaultVite(t), config, nil).Handle(ctx))
}

func TestInputsCommand_Paths(t *testing.T) {
	setupAppViews(t, nil)

	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(template.FuncMap{}).Times(2)
	view.EXPECT().RegisteredViews().Return(nil).Times(2)
	view.EXPECT().RegisteredViewFS().Return(nil).Times(2)

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetStringSlice("vite.inputs", []string{}).Return(nil).Times(2)
	config.EXPECT().GetStringSlice("vite.assets", []string{}).Return(nil).Times(2)

	// Moved directories are reported as configured, the hot file relative to the project root.
	vite := mocksvite.NewVite(t)
	vite.EXPECT().HotFile().Return(filepath.Join(support.RelativePath, "storage", "vite.hot")).Once()
	vite.EXPECT().BuildDirectory().Return("assets").Once()
	vite.EXPECT().ManifestFilename().Return("manifest.json").Once()

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":[],"assets":[],"dynamic":[],"paths":{"public_directory":"public","build_directory":"assets","hot_file":"storage/vite.hot","manifest":"manifest.json"},"modules":[]}`).Once()
	assert.NoError(t, NewInputsCommand(view, vite, config, nil).Handle(ctx))

	// A hot file outside the project root stays absolute.
	outside := filepath.Join(t.TempDir(), "hot")
	vite = mocksvite.NewVite(t)
	vite.EXPECT().HotFile().Return(outside).Once()
	vite.EXPECT().BuildDirectory().Return("build").Once()
	vite.EXPECT().ManifestFilename().Return(".vite/manifest.json").Once()

	ctx = mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":[],"assets":[],"dynamic":[],"paths":{"public_directory":"public","build_directory":"build","hot_file":"` + filepath.ToSlash(outside) + `","manifest":".vite/manifest.json"},"modules":[]}`).Once()
	assert.NoError(t, NewInputsCommand(view, vite, config, nil).Handle(ctx))
}

func TestInputsCommand_NoViews(t *testing.T) {
	setupAppViews(t, nil)

	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(template.FuncMap{}).Once()
	view.EXPECT().RegisteredViews().Return(nil).Once()
	view.EXPECT().RegisteredViewFS().Return(nil).Once()

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetStringSlice("vite.inputs", []string{}).Return(nil).Once()
	config.EXPECT().GetStringSlice("vite.assets", []string{}).Return(nil).Once()

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":[],"assets":[],"dynamic":[],` + defaultPaths + `}`).Once()

	assert.NoError(t, NewInputsCommand(view, defaultVite(t), config, nil).Handle(ctx))
}

func TestInputsCommand_ConfiguredEntries(t *testing.T) {
	setupAppViews(t, nil)

	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(template.FuncMap{}).Once()
	view.EXPECT().RegisteredViews().Return(nil).Once()
	view.EXPECT().RegisteredViewFS().Return(nil).Once()

	config := mocksconfig.NewConfig(t)
	inputs := []string{"resources/inertia/app.ts", "resources/inertia/app.ts", "resources/js/app.ts"}
	config.EXPECT().GetStringSlice("vite.inputs", []string{}).Return(inputs).Once()
	config.EXPECT().GetStringSlice("vite.assets", []string{}).Return([]string{"resources/images/logo.svg"}).Once()

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":["resources/inertia/app.ts","resources/js/app.ts"],"assets":["resources/images/logo.svg"],"dynamic":[],` + defaultPaths + `}`).Once()

	assert.NoError(t, NewInputsCommand(view, defaultVite(t), config, nil).Handle(ctx))

	// The command sorts its own copy, never the value the config stores.
	assert.Equal(t, []string{"resources/inertia/app.ts", "resources/inertia/app.ts", "resources/js/app.ts"}, inputs)
}

func TestInputsCommand_NilConfig(t *testing.T) {
	setupAppViews(t, nil)

	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(template.FuncMap{}).Once()
	view.EXPECT().RegisteredViews().Return(nil).Once()
	view.EXPECT().RegisteredViewFS().Return(nil).Once()

	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(`{"inputs":[],"assets":[],"dynamic":[],` + defaultPaths + `}`).Once()

	assert.NoError(t, NewInputsCommand(view, defaultVite(t), nil, nil).Handle(ctx))
}

func TestInputsCommand_Errors(t *testing.T) {
	ctx := mocksconsole.NewContext(t)
	assert.EqualError(t, NewInputsCommand(nil, nil, nil, nil).Handle(ctx), errors.ViewFacadeNotSet.SetModule(errors.ModuleVite).Error())
	assert.EqualError(t, NewInputsCommand(mocksview.NewView(t), nil, nil, nil).Handle(ctx), errors.ViteFacadeNotSet.Error())

	// A view that does not parse, here because vite is not registered, fails the command.
	setupAppViews(t, map[string]string{"broken.tmpl": `{{ vite "a" }}`})
	view := mocksview.NewView(t)
	view.EXPECT().Funcs().Return(nil).Once()
	view.EXPECT().RegisteredViews().Return(nil).Once()
	view.EXPECT().RegisteredViewFS().Return(nil).Once()

	err := NewInputsCommand(view, mocksvite.NewVite(t), nil, nil).Handle(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[vite] failed to compile views")
	assert.Contains(t, err.Error(), `function "vite" not defined`)
}
