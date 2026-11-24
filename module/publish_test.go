package module

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
)

// Test modules live in this package, so an app with another main module sees them as third-party.
const publishPackage = "github.com/goravel/framework/module"

type sourcesModule struct {
	plainModule
	sources fs.FS
	views   fs.FS
	dir     string
}

func (m *sourcesModule) Frontend() contractsmodule.Frontend {
	return contractsmodule.Frontend{Files: contractsmodule.FS(m.sources).Dir(m.dir)}
}
func (m *sourcesModule) Views() contractsmodule.Files {
	if m.views == nil {
		return contractsmodule.Files{}
	}

	return contractsmodule.FS(m.views).Dir(".")
}

// diskModule declares frontend files without an FS: they are read from its package directory.
type diskModule struct {
	plainModule
	dir     string
	entries []string
}

func (m *diskModule) Frontend() contractsmodule.Frontend {
	return contractsmodule.Frontend{Files: contractsmodule.Path(m.dir), Entries: m.entries}
}

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

	t.Run("frontend files of a directory keep it", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			sources:     fstest.MapFS{"web/js/app.js": {Data: []byte("app")}, "other/x.js": {Data: []byte("x")}},
			dir:         "web",
		}

		_, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.Equal(t, "app", read(t, dir, "web/js/app.js"))
		assert.NoDirExists(t, filepath.Join(dir, "other"))

		module.dir = "../outside"
		_, err = Publish(module, dependency, dir)
		assert.ErrorIs(t, err, errors.ModuleFilesDirInvalid)
	})

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
		assert.Equal(t, errors.ModulePublishTargetNotOwned.Args("billing", dir, dependency.Path), err)
		assert.NoFileExists(t, filepath.Join(dir, "app.js"))

		// Published from another Go module.
		dir = t.TempDir()
		_, err = Publish(module, Dependency{Path: "github.com/other/billing"}, dir)
		require.NoError(t, err)
		_, err = Publish(module, dependency, dir)
		assert.Equal(t, errors.ModulePublishTargetNotOwned.Args("billing", dir, dependency.Path), err)
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

// unreadableFS lists inner but fails to open the file named fail.
type unreadableFS struct {
	inner fs.FS
	fail  string
}

func (r unreadableFS) Open(name string) (fs.File, error) {
	if name == r.fail {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}

	return r.inner.Open(name)
}

func TestPublish_Failures(t *testing.T) {
	dependency := Dependency{Path: "github.com/acme/billing", Version: "v1.2.0"}
	app := fstest.MapFS{"resources/js/app.js": {Data: []byte("app")}}
	// An owned directory: published before, so Publish may write into it.
	owned := func(t *testing.T, files ...string) string {
		dir := t.TempDir()
		content, err := json.Marshal(publishMarker{Dependency: dependency, Files: files})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, PublishMarker), content, 0o644))

		return dir
	}

	t.Run("unreadable sources", func(t *testing.T) {
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			sources:     unreadableFS{inner: app, fail: "resources/js/app.js"},
		}

		_, err := Publish(module, dependency, t.TempDir())
		assert.ErrorIs(t, err, fs.ErrPermission)
		assert.Equal(t, PublishNone, PublishState(module, dependency, t.TempDir()))
	})

	t.Run("unreadable views", func(t *testing.T) {
		module := &sourcesModule{
			plainModule: plainModule{"billing"},
			views:       unreadableFS{inner: fstest.MapFS{"index.tmpl": {Data: []byte("index")}}, fail: "index.tmpl"},
		}

		_, err := Publish(module, dependency, t.TempDir())
		assert.ErrorIs(t, err, fs.ErrPermission)
	})

	t.Run("a file is in the way of a directory", func(t *testing.T) {
		dir := owned(t)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "resources"), []byte("file"), 0o644))

		_, err := Publish(&sourcesModule{plainModule: plainModule{"billing"}, sources: app}, dependency, dir)
		assert.Error(t, err)
	})

	t.Run("a directory is in the way of a file", func(t *testing.T) {
		dir := owned(t)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "resources", "js", "app.js"), 0o755))

		_, err := Publish(&sourcesModule{plainModule: plainModule{"billing"}, sources: app}, dependency, dir)
		assert.Error(t, err)
	})

	t.Run("a dropped file became a directory with content", func(t *testing.T) {
		dir := owned(t, "legacy")
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "legacy"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "legacy", "mine.txt"), []byte("mine"), 0o644))

		_, err := Publish(&sourcesModule{plainModule: plainModule{"billing"}, sources: app}, dependency, dir)
		assert.Error(t, err)
		assert.FileExists(t, filepath.Join(dir, "legacy", "mine.txt"))
	})

	t.Run("marker paths outside the directory are ignored", func(t *testing.T) {
		parent := t.TempDir()
		dir := filepath.Join(parent, "billing")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("keep"), 0o644))
		content, err := json.Marshal(publishMarker{Dependency: dependency, Files: []string{"../outside.txt", "."}})
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, PublishMarker), content, 0o644))

		changed, err := Publish(&sourcesModule{plainModule: plainModule{"billing"}, sources: app}, dependency, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.FileExists(t, filepath.Join(parent, "outside.txt"))
		assert.DirExists(t, dir)
	})
}

func TestPublish_FromDisk(t *testing.T) {
	dependency := Dependency{Path: "github.com/acme/billing", Version: "v1.2.0", Sum: "h1:abc"}
	packageDir := PackageDir
	t.Cleanup(func() { PackageDir = packageDir })

	source := t.TempDir()
	for name, content := range map[string]string{
		"resources/js/app.js":        "app",
		"resources/views/index.tmpl": "index",
		"web/app.css":                "css",
		"module.go":                  "package billing",
		"go.mod":                     "module github.com/acme/billing",
		".gitignore":                 "node_modules",
		".github/ci.yml":             "ci",
		"node_modules/vue/index.js":  "vue",
	} {
		target := filepath.Join(source, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
		require.NoError(t, os.WriteFile(target, []byte(content), 0o644))
	}
	PackageDir = func(pkg string) (string, error) {
		assert.Equal(t, publishPackage, pkg)

		return source, nil
	}
	published := func(t *testing.T, dir string) []string {
		marker, ok := readPublishMarker(dir)
		require.True(t, ok)

		return marker.Files
	}

	t.Run("resources by default", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &diskModule{plainModule: plainModule{"billing"}, entries: []string{"resources/js/app.js"}}

		_, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.Equal(t, []string{PublishNotice, "resources/js/app.js", "resources/views/index.tmpl"}, published(t, dir))
	})

	t.Run("the package directory without go files, dot entries and node_modules", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &diskModule{plainModule: plainModule{"billing"}, dir: ".", entries: []string{"resources/js/app.js"}}
		assert.Equal(t, PublishMissing, PublishState(module, dependency, dir))

		changed, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, []string{PublishNotice, "resources/js/app.js", "resources/views/index.tmpl", "web/app.css"}, published(t, dir))
		assert.Equal(t, PublishPublished, PublishState(module, dependency, dir))

		// The source changed, e.g. a replaced module being worked on.
		require.NoError(t, os.WriteFile(filepath.Join(source, "web", "app.css"), []byte("new"), 0o644))
		assert.Equal(t, PublishStale, PublishState(module, dependency, dir))
		require.NoError(t, os.WriteFile(filepath.Join(source, "web", "app.css"), []byte("css"), 0o644))
	})

	t.Run("a directory of it", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &diskModule{plainModule: plainModule{"billing"}, dir: "web"}

		_, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.Equal(t, []string{PublishNotice, "web/app.css"}, published(t, dir))
	})

	t.Run("an FS is used instead of the disk", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		module := &sourcesModule{plainModule: plainModule{"billing"}, sources: fstest.MapFS{"app.js": {Data: []byte("embedded")}}}
		PackageDir = func(string) (string, error) { return "", assert.AnError }
		t.Cleanup(func() { PackageDir = packageDir })

		_, err := Publish(module, dependency, dir)
		require.NoError(t, err)
		assert.Equal(t, []string{PublishNotice, "app.js"}, published(t, dir))
	})

	t.Run("failures", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "modules", "billing")
		PackageDir = func(string) (string, error) { return source, nil }

		for _, sub := range []string{"missing", "module.go"} {
			_, err := Publish(&diskModule{plainModule: plainModule{"billing"}, dir: sub}, dependency, dir)
			assert.ErrorIs(t, err, errors.ModuleFilesDirUnreadable, sub)
		}
		_, err := Publish(&diskModule{plainModule: plainModule{"billing"}, dir: "../shop"}, dependency, dir)
		assert.ErrorIs(t, err, errors.ModuleFilesDirInvalid)

		PackageDir = func(string) (string, error) { return "", assert.AnError }
		_, err = Publish(&diskModule{plainModule: plainModule{"billing"}, entries: []string{"app.js"}}, dependency, dir)
		assert.ErrorIs(t, err, errors.ModuleFrontendSourcesNotFound)

		// The default directory is not optional.
		PackageDir = func(string) (string, error) { return t.TempDir(), nil }
		_, err = Publish(&diskModule{plainModule: plainModule{"billing"}, entries: []string{"app.js"}}, dependency, dir)
		assert.ErrorIs(t, err, errors.ModuleFilesDirUnreadable)

		// Nothing declared: the disk is not asked.
		changed, err := Publish(&diskModule{plainModule: plainModule{"billing"}}, dependency, dir)
		assert.NoError(t, err)
		assert.False(t, changed)
	})
}

func TestPackageDir(t *testing.T) {
	dir, err := PackageDir(publishPackage)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(dir, "publish.go"))

	_, err = PackageDir("example.com/missing/package")
	assert.Error(t, err)
}
