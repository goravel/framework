package view

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
)

//go:embed testdata/views
var testViews embed.FS

func TestView(t *testing.T) {
	view := NewView()
	view.Share("a", "b")
	assert.Equal(t, "b", view.Shared("a"))
	assert.Equal(t, "c", view.Shared("b", "c"))
	assert.Equal(t, map[string]any{"a": "b"}, view.GetShared())
}

func TestLoadViewsFrom(t *testing.T) {
	view := NewView()
	assert.Empty(t, view.RegisteredViews())
	assert.Empty(t, view.RegisteredViewFS())

	view.LoadViewsFrom("/pkg/a/views")
	view.LoadViewsFrom("/pkg/b/views")

	assert.Equal(t, []string{"/pkg/a/views", "/pkg/b/views"}, view.RegisteredViews())
	assert.Empty(t, view.RegisteredViewFS(), "filesystem paths must not leak into fs sources")

	registered := view.RegisteredViews()
	registered[0] = "mutated"
	assert.Equal(t, "/pkg/a/views", view.RegisteredViews()[0], "RegisteredViews must return a copy")
}

func TestLoadViewsFromFS(t *testing.T) {
	pkgB := fstest.MapFS{
		"views/bar.tmpl": {Data: []byte(`{{ define "bar.tmpl" }}package bar{{ end }}`)},
	}

	view := NewView()
	view.LoadViewsFromFS(testViews, "testdata/views")
	view.LoadViewsFromFS(pkgB, "views")

	filesystems := view.RegisteredViewFS()
	require.Len(t, filesystems, 2)
	assert.Empty(t, view.RegisteredViews(), "fs sources must not leak into filesystem paths")

	// Each source is rooted at its registered root and preserves registration order.
	content, err := fs.ReadFile(filesystems[0], "foo.tmpl")
	require.NoError(t, err)
	assert.Contains(t, string(content), "package foo")

	content, err = fs.ReadFile(filesystems[1], "bar.tmpl")
	require.NoError(t, err)
	assert.Contains(t, string(content), "package bar")

	// Nested directories are reachable relative to the root.
	_, err = fs.Stat(filesystems[0], "layouts/app.tmpl")
	assert.NoError(t, err)

	filesystems[0] = nil
	assert.NotNil(t, view.RegisteredViewFS()[0], "RegisteredViewFS must return a copy")
}

func TestLoadViewsFromFS_RootNormalization(t *testing.T) {
	fsys := fstest.MapFS{
		"views/index.tmpl": {Data: []byte("index")},
	}

	tests := []struct {
		name string
		root string
		file string
	}{
		{name: "plain", root: "views", file: "index.tmpl"},
		{name: "trailing slash", root: "views/", file: "index.tmpl"},
		{name: "leading slash", root: "/views", file: "index.tmpl"},
		{name: "dot prefix", root: "./views", file: "index.tmpl"},
		{name: "os separator", root: filepath.FromSlash("views/"), file: "index.tmpl"},
		{name: "dot root", root: ".", file: "views/index.tmpl"},
		{name: "slash root", root: "/", file: "views/index.tmpl"},
		{name: "empty root", root: "", file: "views/index.tmpl"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			view := NewView()
			view.LoadViewsFromFS(fsys, test.root)

			content, err := fs.ReadFile(view.RegisteredViewFS()[0], test.file)
			require.NoError(t, err)
			assert.Equal(t, "index", string(content))
		})
	}
}

func TestLoadViewsFromFS_Invalid(t *testing.T) {
	t.Run("nil filesystem", func(t *testing.T) {
		view := NewView()
		assert.PanicsWithError(t, errors.ViewFSRequired.Error(), func() {
			view.LoadViewsFromFS(nil, "views")
		})
		assert.Empty(t, view.RegisteredViewFS())
	})

	t.Run("root escapes filesystem", func(t *testing.T) {
		view := NewView()
		assert.PanicsWithError(t, errors.ViewInvalidFSRoot.Args("../views", &fs.PathError{Op: "sub", Path: "../views", Err: fs.ErrInvalid}).Error(), func() {
			view.LoadViewsFromFS(fstest.MapFS{}, "../views")
		})
		assert.Empty(t, view.RegisteredViewFS())
	})
}

func TestExists(t *testing.T) {
	appDir := setupAppViews(t, map[string]string{
		"app.tmpl":              "app",
		"admin/dashboard.tmpl":  "app dashboard",
		"override/foo.tmpl":     "app foo",
		"deep/nested/leaf.tmpl": "leaf",
	})

	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "dir.tmpl"), []byte("dir"), 0o644))

	view := NewView()
	view.LoadViewsFrom(pkgDir)
	view.LoadViewsFromFS(testViews, "testdata/views")
	view.LoadViewsFromFS(fstest.MapFS{
		"views/bar.tmpl": {Data: []byte("bar")},
	}, "views")

	tests := []struct {
		name   string
		view   string
		exists bool
	}{
		{name: "app view", view: "app.tmpl", exists: true},
		{name: "nested app view", view: "admin/dashboard.tmpl", exists: true},
		{name: "deeply nested app view", view: "deep/nested/leaf.tmpl", exists: true},
		{name: "filesystem package view", view: "dir.tmpl", exists: true},
		{name: "embedded package view", view: "foo.tmpl", exists: true},
		{name: "nested embedded package view", view: "layouts/app.tmpl", exists: true},
		{name: "nested embedded package view with os separator", view: filepath.FromSlash("pages/home.tmpl"), exists: true},
		{name: "second embedded package view", view: "bar.tmpl", exists: true},
		{name: "embedded view outside root", view: "views/bar.tmpl", exists: false},
		{name: "embedded view addressed by unrooted path", view: "testdata/views/foo.tmpl", exists: false},
		{name: "missing view", view: "missing.tmpl", exists: false},
		{name: "missing nested view", view: "layouts/missing.tmpl", exists: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.exists, view.Exists(test.view))
		})
	}

	// Sanity check that the app views were really the ones being found.
	require.NoError(t, os.RemoveAll(appDir))
	assert.False(t, view.Exists("app.tmpl"))
	assert.True(t, view.Exists("foo.tmpl"))
}

// TestResolveTemplates exercises the registry the way a route driver does: app views win,
// then filesystem packages, then embedded packages, with nested define/template/partials
// resolved within the same template set.
func TestResolveTemplates(t *testing.T) {
	appDir := setupAppViews(t, map[string]string{
		"foo.tmpl":  `{{ define "foo.tmpl" }}app foo{{ end }}`,
		"only.tmpl": `{{ define "only.tmpl" }}app only{{ end }}`,
	})

	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "dir.tmpl"), []byte(`{{ define "dir.tmpl" }}dir pkg{{ end }}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "shared.tmpl"), []byte(`{{ define "shared.tmpl" }}from dir{{ end }}`), 0o644))

	pkgB := fstest.MapFS{
		"views/bar.tmpl":    {Data: []byte(`{{ define "bar.tmpl" }}package bar{{ end }}`)},
		"views/shared.tmpl": {Data: []byte(`{{ define "shared.tmpl" }}from embedded{{ end }}`)},
		"other/nope.tmpl":   {Data: []byte(`{{ define "nope.tmpl" }}outside root{{ end }}`)},
	}

	view := NewView()
	view.LoadViewsFrom(pkgDir)
	view.LoadViewsFromFS(testViews, "testdata/views")
	view.LoadViewsFromFS(pkgB, "views")

	tmpl := parseAll(t, appDir, view)

	tests := []struct {
		name     string
		template string
		data     any
		expect   string
	}{
		{name: "application overrides embedded package", template: "foo.tmpl", expect: "app foo"},
		{name: "application only", template: "only.tmpl", expect: "app only"},
		{name: "filesystem package", template: "dir.tmpl", expect: "dir pkg"},
		{name: "filesystem package overrides embedded package", template: "shared.tmpl", expect: "from dir"},
		{name: "embedded package fallback", template: "bar.tmpl", expect: "package bar"},
		{
			name:     "nested layout, partial and block from embedded package",
			template: "pages/home.tmpl",
			data:     map[string]any{"Title": "Home", "Nav": "Menu"},
			expect:   "<html><body><nav>Menu</nav><main><h1>Home</h1></main></body></html>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tmpl.ExecuteTemplate(&buf, test.template, test.data))
			assert.Equal(t, test.expect, buf.String())
		})
	}

	t.Run("missing template", func(t *testing.T) {
		assert.Nil(t, tmpl.Lookup("missing.tmpl"))
		assert.False(t, view.Exists("missing.tmpl"))
	})

	t.Run("embedded file outside root is not loaded", func(t *testing.T) {
		assert.Nil(t, tmpl.Lookup("nope.tmpl"))
		assert.False(t, view.Exists("nope.tmpl"))
	})
}

// setupAppViews points the application resources path at a temp dir and writes the given views into it.
func setupAppViews(t *testing.T, files map[string]string) string {
	t.Helper()

	relativePath := support.RelativePath
	resources := support.Config.Paths.Resources
	t.Cleanup(func() {
		support.RelativePath = relativePath
		support.Config.Paths.Resources = resources
	})

	support.RelativePath = t.TempDir()
	support.Config.Paths.Resources = "resources"

	dir := filepath.Join(support.RelativePath, "resources", "views")
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}

	return dir
}

// parseAll mirrors a route driver: it parses every source into one template set, in precedence
// order, skipping any file whose define name was already claimed by an earlier source.
func parseAll(t *testing.T, appDir string, view *View) *template.Template {
	t.Helper()

	tmpl := template.New("")
	seen := map[string]bool{}

	sources := []fs.FS{os.DirFS(appDir)}
	for _, dir := range view.RegisteredViews() {
		sources = append(sources, os.DirFS(dir))
	}
	sources = append(sources, view.RegisteredViewFS()...)

	for _, fsys := range sources {
		var files []string
		require.NoError(t, fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			content, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}

			name := defineRe.FindStringSubmatch(string(content))
			if len(name) > 1 {
				if seen[name[1]] {
					return nil
				}
				seen[name[1]] = true
			}

			files = append(files, p)

			return nil
		}))

		if len(files) > 0 {
			_, err := tmpl.ParseFS(fsys, files...)
			require.NoError(t, err)
		}
	}

	return tmpl
}

var defineRe = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)
