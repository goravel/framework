package view

import (
	"bytes"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockslog "github.com/goravel/framework/mocks/log"
)

// failingFS fails to open one path of its base filesystem.
type failingFS struct {
	base     fs.FS
	failPath string
}

func (f failingFS) Open(name string) (fs.File, error) {
	if name == f.failPath {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}

	return f.base.Open(name)
}

func render(t *testing.T, tmpl *template.Template, name string, data any) string {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, tmpl.ExecuteTemplate(&buf, name, data), name)

	return buf.String()
}

func writeFiles(t *testing.T, dir string, files map[string]string) string {
	t.Helper()

	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}

	return dir
}

func TestCompile_NoSources(t *testing.T) {
	setupAppViews(t, nil)

	tmpl, err := Compile(CompileOptions{})
	assert.NoError(t, err)
	assert.Nil(t, tmpl, "no view file anywhere yields no template set")

	tmpl, err = Compile(CompileOptions{Paths: []string{t.TempDir()}, FS: []fs.FS{fstest.MapFS{}}})
	assert.NoError(t, err)
	assert.Nil(t, tmpl, "empty sources yield no template set")
}

func TestCompile_AppViewsNamedByRelativePath(t *testing.T) {
	setupAppViews(t, map[string]string{
		"index.html":        "Root",
		"admin/index.html":  "Admin",
		"public/index.html": "Public",
	})

	tmpl, err := Compile(CompileOptions{})
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	assert.Equal(t, "Root", render(t, tmpl, "index.html", nil))
	assert.Equal(t, "Admin", render(t, tmpl, "admin/index.html", nil))
	assert.Equal(t, "Public", render(t, tmpl, "public/index.html", nil))
}

func TestCompile_Extensions(t *testing.T) {
	setupAppViews(t, map[string]string{
		"page.tmpl":  `{{ define "page.tmpl" }}Page{{ end }}`,
		"notes.md":   "# Not a template {{",
		"style.css":  "body {}",
		"other.html": "Other",
	})

	t.Run("no filter parses every file and fails on the broken one", func(t *testing.T) {
		_, err := Compile(CompileOptions{})
		assert.ErrorContains(t, err, "failed to parse view")
		assert.ErrorContains(t, err, "notes.md")
	})

	t.Run("filter parses only the listed extensions", func(t *testing.T) {
		tmpl, err := Compile(CompileOptions{Extensions: []string{".tmpl", ".html"}})
		require.NoError(t, err)
		require.NotNil(t, tmpl)

		assert.Equal(t, "Page", render(t, tmpl, "page.tmpl", nil))
		assert.Equal(t, "Other", render(t, tmpl, "other.html", nil))
		assert.Nil(t, tmpl.Lookup("style.css"))
		assert.Nil(t, tmpl.Lookup("notes.md"))
	})
}

func TestCompile_Precedence(t *testing.T) {
	setupAppViews(t, map[string]string{
		"page.tmpl":        `{{ define "page.tmpl" }}App Page{{ end }}`,
		"raw.tmpl":         "App Raw",
		"nested/leaf.tmpl": "App Leaf",
	})
	pkgDir := writeFiles(t, t.TempDir(), map[string]string{
		"page.tmpl":        `{{ define "page.tmpl" }}Dir Page{{ end }}`,
		"raw.tmpl":         "Dir Raw",
		"dir.tmpl":         `{{ define "dir.tmpl" }}Dir Only{{ end }}`,
		"shared.tmpl":      `{{ define "shared.tmpl" }}From Dir{{ end }}`,
		"nested/leaf.tmpl": "Dir Leaf",
	})
	pkgFS := fstest.MapFS{
		"shared.tmpl": {Data: []byte(`{{ define "shared.tmpl" }}From FS{{ end }}`)},
		"fs.tmpl":     {Data: []byte(`{{ define "fs.tmpl" }}FS Only{{ end }}`)},
		"raw.tmpl":    {Data: []byte("FS Raw")},
	}

	tmpl, err := Compile(CompileOptions{Paths: []string{pkgDir}, FS: []fs.FS{pkgFS}})
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	assert.Equal(t, "App Page", render(t, tmpl, "page.tmpl", nil), "app define wins over package define")
	assert.Equal(t, "App Raw", render(t, tmpl, "raw.tmpl", nil), "app define-less file wins by path")
	assert.Equal(t, "App Leaf", render(t, tmpl, "nested/leaf.tmpl", nil), "nested define-less app file wins by path")
	assert.Equal(t, "Dir Only", render(t, tmpl, "dir.tmpl", nil))
	assert.Equal(t, "From Dir", render(t, tmpl, "shared.tmpl", nil), "directories rank above filesystems")
	assert.Equal(t, "FS Only", render(t, tmpl, "fs.tmpl", nil))
}

func TestCompile_PackageCannotOverrideApp(t *testing.T) {
	t.Run("define-less package file cannot override an app define", func(t *testing.T) {
		setupAppViews(t, map[string]string{"widget.tmpl": `{{ define "widget.tmpl" }}App Widget{{ end }}`})
		pkg := fstest.MapFS{"widget.tmpl": {Data: []byte("Package Raw")}}

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{pkg}})
		require.NoError(t, err)
		assert.Equal(t, "App Widget", render(t, tmpl, "widget.tmpl", nil))
	})

	t.Run("package file with a define cannot override an app template at the same path", func(t *testing.T) {
		setupAppViews(t, map[string]string{"page.tmpl": "App Page"})
		pkg := fstest.MapFS{"page.tmpl": {Data: []byte(`{{ define "widget" }}Widget{{ end }}Package Raw`)}}

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{pkg}})
		require.NoError(t, err)
		assert.Equal(t, "App Page", render(t, tmpl, "page.tmpl", nil))
		assert.Nil(t, tmpl.Lookup("widget"), "the whole file is skipped, not just the colliding name")
	})
}

func TestCompile_PackageCollisions(t *testing.T) {
	setupAppViews(t, nil)

	t.Run("same define name in two packages keeps the first and warns", func(t *testing.T) {
		dir1 := writeFiles(t, t.TempDir(), map[string]string{"layout.tmpl": `{{ define "layout.tmpl" }}First{{ end }}`})
		dir2 := writeFiles(t, t.TempDir(), map[string]string{"layout.tmpl": `{{ define "layout.tmpl" }}Second{{ end }}`})

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view collision: %q defined in %q and %q, using first", "layout.tmpl", filepath.Join(dir1, "layout.tmpl"), filepath.Join(dir2, "layout.tmpl")).Return().Once()

		tmpl, err := Compile(CompileOptions{Paths: []string{dir1, dir2}, Log: mockLog})
		require.NoError(t, err)
		assert.Equal(t, "First", render(t, tmpl, "layout.tmpl", nil))
	})

	t.Run("same path in two filesystems keeps the first, warns, and drops the loser's define", func(t *testing.T) {
		fs1 := fstest.MapFS{"page.tmpl": {Data: []byte("First Raw")}}
		fs2 := fstest.MapFS{"page.tmpl": {Data: []byte(`{{ define "loser" }}Loser{{ end }}Second Raw`)}}

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view collision: %q defined in %q and %q, using first", "page.tmpl", "fs[0]/page.tmpl", "fs[1]/page.tmpl").Return().Once()

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{fs1, fs2}, Log: mockLog})
		require.NoError(t, err)
		assert.Equal(t, "First Raw", render(t, tmpl, "page.tmpl", nil))
		assert.Nil(t, tmpl.Lookup("loser"))
	})

	t.Run("without a log the collision is silent", func(t *testing.T) {
		fs1 := fstest.MapFS{"page.tmpl": {Data: []byte("First")}}
		fs2 := fstest.MapFS{"page.tmpl": {Data: []byte("Second")}}

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{fs1, fs2}})
		require.NoError(t, err)
		assert.Equal(t, "First", render(t, tmpl, "page.tmpl", nil))
	})
}

func TestCompile_SkippedSources(t *testing.T) {
	setupAppViews(t, nil)
	good := fstest.MapFS{"ok.tmpl": {Data: []byte("ok")}}

	t.Run("missing directory and regular file as a path are skipped", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "not-a-dir.tmpl")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))

		tmpl, err := Compile(CompileOptions{Paths: []string{filepath.Join(t.TempDir(), "missing"), file}, FS: []fs.FS{good}})
		require.NoError(t, err)
		assert.Equal(t, "ok", render(t, tmpl, "ok.tmpl", nil))
	})

	t.Run("nil filesystem is skipped and warned about, keeping its index", func(t *testing.T) {
		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view source fs[%d] is nil, skipping", 0).Return().Once()

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{nil, good}, Log: mockLog})
		require.NoError(t, err)
		assert.Equal(t, "ok", render(t, tmpl, "ok.tmpl", nil))
	})

	t.Run("unreadable filesystem is skipped and warned about", func(t *testing.T) {
		unreadable := failingFS{base: good, failPath: "."}

		mockLog := mockslog.NewLog(t)
		mockLog.EXPECT().Warningf("view source fs[%d] is unreadable, skipping: %v", 1, &fs.PathError{Op: "open", Path: ".", Err: fs.ErrPermission}).Return().Once()

		tmpl, err := Compile(CompileOptions{FS: []fs.FS{good, unreadable}, Log: mockLog})
		require.NoError(t, err)
		assert.Equal(t, "ok", render(t, tmpl, "ok.tmpl", nil))
	})

	t.Run("nil and unreadable filesystems without a log", func(t *testing.T) {
		tmpl, err := Compile(CompileOptions{FS: []fs.FS{nil, failingFS{base: good, failPath: "."}}})
		assert.NoError(t, err)
		assert.Nil(t, tmpl)
	})
}

func TestCompile_Errors(t *testing.T) {
	setupAppViews(t, nil)
	base := fstest.MapFS{
		"page.tmpl":     {Data: []byte(`{{ define "page.tmpl" }}Page{{ end }}`)},
		"sub/deep.tmpl": {Data: []byte(`{{ define "sub/deep.tmpl" }}Deep{{ end }}`)},
	}

	for failPath, expected := range map[string]string{
		"sub":       "failed to walk view source fs[0]/sub: ",
		"page.tmpl": "failed to read view fs[0]/page.tmpl: ",
	} {
		t.Run(failPath, func(t *testing.T) {
			tmpl, err := Compile(CompileOptions{FS: []fs.FS{failingFS{base: base, failPath: failPath}}})
			assert.ErrorContains(t, err, expected)
			assert.ErrorContains(t, err, "permission denied")
			assert.Nil(t, tmpl)
		})
	}

	t.Run("parse error names the source", func(t *testing.T) {
		dir := writeFiles(t, t.TempDir(), map[string]string{"broken.tmpl": `{{ define "broken.tmpl" }}{{ .Unclosed`})

		tmpl, err := Compile(CompileOptions{Paths: []string{dir}})
		assert.ErrorContains(t, err, "failed to parse view "+filepath.Join(dir, "broken.tmpl")+": ")
		assert.Nil(t, tmpl)
	})
}

func TestCompile_DelimsAndFuncs(t *testing.T) {
	setupAppViews(t, nil)

	t.Run("custom delimiters apply to parsing and define extraction", func(t *testing.T) {
		dir1 := writeFiles(t, t.TempDir(), map[string]string{"page.tmpl": `{[ define "page.tmpl" ]}First{[ end ]}`})
		dir2 := writeFiles(t, t.TempDir(), map[string]string{"page.tmpl": `{[ define "page.tmpl" ]}Second{[ end ]}`})

		tmpl, err := Compile(CompileOptions{Delims: &Delims{Left: "{[", Right: "]}"}, Paths: []string{dir1, dir2}})
		require.NoError(t, err)
		assert.Equal(t, "First", render(t, tmpl, "page.tmpl", nil))
	})

	t.Run("funcs are available to every source", func(t *testing.T) {
		dir := writeFiles(t, t.TempDir(), map[string]string{"greet.tmpl": `{{ define "greet.tmpl" }}{{ upper "hi" }}{{ end }}`})

		tmpl, err := Compile(CompileOptions{FuncMap: template.FuncMap{"upper": strings.ToUpper}, Paths: []string{dir}})
		require.NoError(t, err)
		assert.Equal(t, "HI", render(t, tmpl, "greet.tmpl", nil))
	})
}

func TestCompile_NestedLayoutsFromEmbeddedViews(t *testing.T) {
	setupAppViews(t, nil)
	sub, err := fs.Sub(testViews, "testdata/views")
	require.NoError(t, err)

	tmpl, err := Compile(CompileOptions{FS: []fs.FS{sub}})
	require.NoError(t, err)
	assert.Equal(t,
		"<html><body><nav>Menu</nav><main><h1>Home</h1></main></body></html>",
		render(t, tmpl, "pages/home.tmpl", map[string]any{"Title": "Home", "Nav": "Menu"}))
}

func TestCompile_GlobMetacharactersInNames(t *testing.T) {
	setupAppViews(t, nil)
	pkg := fstest.MapFS{
		"[id].tmpl":    {Data: []byte("bracket")},
		"what?.tmpl":   {Data: []byte("question")},
		"star*.tmpl":   {Data: []byte("star")},
		"plain.tmpl":   {Data: []byte("plain")},
		"a/[b]/c.tmpl": {Data: []byte("nested")},
	}

	tmpl, err := Compile(CompileOptions{FS: []fs.FS{pkg}})
	require.NoError(t, err)
	for name, expected := range map[string]string{
		"[id].tmpl": "bracket", "what?.tmpl": "question", "star*.tmpl": "star", "plain.tmpl": "plain", "a/[b]/c.tmpl": "nested",
	} {
		assert.Equal(t, expected, render(t, tmpl, name, nil), name)
	}
}

func TestExtractDefineName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		left     string
		expected string
	}{
		{name: "single line define", content: `{{ define "index.tmpl" }}hello{{ end }}`, expected: "index.tmpl"},
		{name: "define with extra whitespace", content: `{{ define    "page.tmpl" }}`, expected: "page.tmpl"},
		{name: "no define", content: `hello world`},
		{name: "block instead of define", content: `{{ block "content" . }}{{ end }}`},
		{name: "custom delimiter does not match the standard left", content: `{[ define "cm.tmpl" ]}Custom{[ end ]}`},
		{name: "custom delimiter with matching left", content: `{[ define "cm.tmpl" ]}Custom{[ end ]}`, left: "{[", expected: "cm.tmpl"},
		{name: "custom delimiter is cached", content: `<% define "erb.tmpl" %>`, left: "<%", expected: "erb.tmpl"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			left := test.left
			if left == "" {
				left = "{{"
			}
			assert.Equal(t, test.expected, extractDefineName(test.content, left))
			assert.Equal(t, test.expected, extractDefineName(test.content, left), "second call uses the cached regexp")
		})
	}
}
