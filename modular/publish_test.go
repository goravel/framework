package modular

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

// Test modules live in this package, so an app with another main module sees them as third-party.
const publishPackage = "github.com/goravel/framework/modular"

type sourcesModule struct {
	plainModule
	sources fs.FS
	views   fs.FS
}

func (m *sourcesModule) Sources() fs.FS { return m.sources }
func (m *sourcesModule) Views() fs.FS   { return m.views }

type locatedSourcesModule struct{ sourcesModule }

func (m *locatedSourcesModule) Path() string { return "vendor/billing" }

func buildInfo(deps ...*debug.Module) *debug.BuildInfo {
	return &debug.BuildInfo{Main: debug.Module{Path: "example.com/app"}, Deps: deps}
}

func TestDependencyOf(t *testing.T) {
	module := &sourcesModule{plainModule: plainModule{"billing"}}
	framework := &debug.Module{Path: "github.com/goravel/framework", Version: "v1.3.0", Sum: "h1:abc"}

	t.Run("third-party", func(t *testing.T) {
		dependency, ok := DependencyOf(module, buildInfo(&debug.Module{Path: "github.com/goravel"}, framework, &debug.Module{Path: "github.com/goravel/framework/mod"}))

		assert.True(t, ok)
		assert.Equal(t, Dependency{Path: "github.com/goravel/framework", Version: "v1.3.0", Sum: "h1:abc"}, dependency)
	})

	t.Run("replaced", func(t *testing.T) {
		replaced := &debug.Module{Path: "github.com/goravel/framework", Version: "v1.3.0", Replace: &debug.Module{Path: "../framework", Version: "(devel)"}}
		dependency, ok := DependencyOf(module, buildInfo(replaced))

		assert.True(t, ok)
		assert.Equal(t, Dependency{Path: "github.com/goravel/framework", Version: "(devel)"}, dependency)
	})

	t.Run("no matching dependency", func(t *testing.T) {
		dependency, ok := DependencyOf(module, buildInfo())

		assert.True(t, ok)
		assert.Equal(t, Dependency{Path: publishPackage}, dependency)
	})

	t.Run("application module", func(t *testing.T) {
		_, ok := DependencyOf(module, &debug.BuildInfo{Main: debug.Module{Path: "github.com/goravel/framework"}})
		assert.False(t, ok)
	})

	t.Run("located module", func(t *testing.T) {
		_, ok := DependencyOf(&locatedSourcesModule{}, buildInfo(framework))
		assert.False(t, ok)
	})

	t.Run("no build info", func(t *testing.T) {
		_, ok := DependencyOf(module, nil)
		assert.False(t, ok)
	})
}

func TestPublish(t *testing.T) {
	dependency := Dependency{Path: "github.com/acme/billing", Version: "v1.2.0", Sum: "h1:abc"}
	read := func(t *testing.T, dir, name string) string {
		content, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		require.NoError(t, err)

		return string(content)
	}

	t.Run("writes sources, views, notice and marker", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			sources: fstest.MapFS{
				"resources/js/app.js":    {Data: []byte("app")},
				"resources/js/legacy.js": {Data: []byte("legacy")},
			},
			views: fstest.MapFS{"layouts/base.tmpl": {Data: []byte("base")}},
		}
		assert.Equal(t, PublishMissing, PublishState(module, dependency, dir))

		changed, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, "app", read(t, dir, "resources/js/app.js"))
		assert.Equal(t, "base", read(t, dir, "resources/views/layouts/base.tmpl"))
		assert.Contains(t, read(t, dir, PublishNotice), "`billing` (github.com/acme/billing v1.2.0)")
		assert.JSONEq(t, `{
			"module": "github.com/acme/billing", "version": "v1.2.0", "sum": "h1:abc",
			"files": ["DO-NOT-EDIT.md", "resources/js/app.js", "resources/js/legacy.js", "resources/views/layouts/base.tmpl"]
		}`, read(t, dir, PublishMarker))
		assert.Equal(t, PublishPublished, PublishState(module, dependency, dir))

		// Unchanged the second time.
		changed, err = Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.False(t, changed)

		// A new version drops a file, changes one, and keeps files it did not write.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("mine"), 0o644))
		upgraded := Dependency{Path: dependency.Path, Version: "v1.3.0", Sum: "h1:def"}
		module.sources = fstest.MapFS{"resources/js/app.js": {Data: []byte("app v2")}}
		module.views = nil
		assert.Equal(t, PublishStale, PublishState(module, upgraded, dir))

		changed, err = Publish(module, upgraded, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, "app v2", read(t, dir, "resources/js/app.js"))
		assert.Equal(t, "mine", read(t, dir, "notes.txt"))
		assert.NoFileExists(t, filepath.Join(dir, "resources/js/legacy.js"))
		assert.NoDirExists(t, filepath.Join(dir, "resources/views"))
		assert.Contains(t, read(t, dir, PublishMarker), `"version": "v1.3.0"`)
		assert.Equal(t, PublishPublished, PublishState(module, upgraded, dir))
	})

	t.Run("views inside the sources are not published twice", func(t *testing.T) {
		dir := t.TempDir()
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			sources:     fstest.MapFS{"resources/views/index.tmpl": {Data: []byte("source")}},
			views:       fstest.MapFS{"index.tmpl": {Data: []byte("embedded")}, "extra.tmpl": {Data: []byte("extra")}},
		}

		changed, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, "source", read(t, dir, "resources/views/index.tmpl"))
		assert.NoFileExists(t, filepath.Join(dir, "resources/views/extra.tmpl"))
	})

	t.Run("nothing to publish", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "billing")
		module := &sourcesModule{plainModule: plainModule{"billing"}}

		changed, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.False(t, changed)
		assert.NoDirExists(t, dir)
		assert.Equal(t, PublishNone, PublishState(module, dependency, dir))
	})

	t.Run("refuses a directory it does not own", func(t *testing.T) {
		module := &sourcesModule{plainModule: plainModule{"billing"}, sources: fstest.MapFS{"app.js": {Data: []byte("app")}}}

		// Someone's code without a marker.
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "module.go"), []byte("package billing"), 0o644))
		_, err := Publish(module, dependency, dir)
		assert.Equal(t, errors.ModularPublishTargetNotOwned.Args("billing", dir, dependency.Path), err)
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))

		// Published from another Go module.
		dir = t.TempDir()
		_, err = Publish(module, Dependency{Path: "github.com/other/billing"}, dir)
		require.NoError(t, err)
		_, err = Publish(module, dependency, dir)
		assert.Equal(t, errors.ModularPublishTargetNotOwned.Args("billing", dir, dependency.Path), err)
		assert.Equal(t, PublishConflict, PublishState(module, dependency, dir))
	})

	t.Run("state follows the files, not only the version", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "billing")
		module := &sourcesModule{plainModule: plainModule{"billing"}, sources: fstest.MapFS{"app.js": {Data: []byte("app")}}}
		replaced := Dependency{Path: dependency.Path, Version: "(devel)"}

		_, err := Publish(module, replaced, dir)
		require.NoError(t, err)
		assert.Equal(t, PublishPublished, PublishState(module, replaced, dir))

		// The module changed without a new version.
		module.sources = fstest.MapFS{"app.js": {Data: []byte("app v2")}}
		assert.Equal(t, PublishStale, PublishState(module, replaced, dir))
		changed, err := Publish(module, replaced, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, PublishPublished, PublishState(module, replaced, dir))

		// The copy was edited or lost a file.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "app.js"), []byte("edited"), 0o644))
		assert.Equal(t, PublishStale, PublishState(module, replaced, dir))
		require.NoError(t, os.Remove(filepath.Join(dir, "app.js")))
		assert.Equal(t, PublishStale, PublishState(module, replaced, dir))

		// The module ships another file.
		module.sources = fstest.MapFS{"app.js": {Data: []byte("app v2")}, "extra.js": {Data: []byte("extra")}}
		_, err = Publish(module, replaced, dir)
		require.NoError(t, err)
		module.sources = fstest.MapFS{"app.js": {Data: []byte("app v2")}}
		assert.Equal(t, PublishStale, PublishState(module, replaced, dir))
	})

	t.Run("conflict when the directory is not a publish of the module", func(t *testing.T) {
		module := &sourcesModule{plainModule: plainModule{"billing"}, sources: fstest.MapFS{"app.js": {Data: []byte("app")}}}

		dir := t.TempDir()
		assert.Equal(t, PublishMissing, PublishState(module, dependency, dir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "module.go"), []byte("package billing"), 0o644))
		assert.Equal(t, PublishConflict, PublishState(module, dependency, dir))

		require.NoError(t, os.WriteFile(filepath.Join(dir, PublishMarker), []byte("{"), 0o644))
		assert.Equal(t, PublishConflict, PublishState(module, dependency, dir))
	})

	t.Run("go files are never published", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "billing")
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			sources: fstest.MapFS{
				"module.go":           {Data: []byte("package billing")},
				"go.mod":              {Data: []byte("module x")},
				"go.sum":              {Data: []byte("")},
				"http/controller.go":  {Data: []byte("package http")},
				"resources/js/app.js": {Data: []byte("app")},
				"resources/go.md":     {Data: []byte("kept")},
			},
		}

		_, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		for _, name := range []string{"module.go", "go.mod", "go.sum", "http/controller.go", "http"} {
			assert.NoFileExists(t, filepath.Join(dir, name))
			assert.NoDirExists(t, filepath.Join(dir, name))
		}
		assert.Equal(t, "app", read(t, dir, "resources/js/app.js"))
		assert.Equal(t, "kept", read(t, dir, "resources/go.md"))
		assert.NotContains(t, read(t, dir, PublishMarker), ".go\"")

		// Only Go files: nothing to publish.
		onlyGo := &sourcesModule{plainModule: plainModule{"billing"}, sources: fstest.MapFS{"module.go": {Data: []byte("package billing")}}}
		assert.Equal(t, PublishNone, PublishState(onlyGo, dependency, t.TempDir()))
	})
}
