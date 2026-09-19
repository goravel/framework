package view

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

func TestTemplateRender(t *testing.T) {
	setupAppViews(t, map[string]string{
		"foo.tmpl":      `{{ define "foo.tmpl" }}app foo{{ end }}`,
		"greeting.tmpl": `{{ define "greeting.tmpl" }}{{ .Greeting }}, {{ .Name }}{{ end }}`,
		"raw.tmpl":      `raw {{ .Name }}`,
	})

	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "shared.tmpl"), []byte(`{{ define "shared.tmpl" }}from dir{{ end }}`), 0o644))

	pkgB := fstest.MapFS{
		"views/bar.tmpl":    {Data: []byte(`{{ define "bar.tmpl" }}package bar{{ end }}`)},
		"views/shared.tmpl": {Data: []byte(`{{ define "shared.tmpl" }}from embedded{{ end }}`)},
	}

	view := NewView()
	view.Share("Greeting", "Hello")
	view.Share("Name", "shared")
	view.LoadViewsFrom(pkgDir)
	view.LoadViewsFromFS(testViews, "testdata/views")
	view.LoadViewsFromFS(pkgB, "views")

	type page struct {
		Title   string
		Nav     string
		private string
	}

	tests := []struct {
		name   string
		view   string
		data   []any
		expect string
	}{
		{name: "application overrides embedded package", view: "foo.tmpl", expect: "app foo"},
		{name: "filesystem package overrides embedded package", view: "shared.tmpl", expect: "from dir"},
		{name: "embedded package fallback", view: "bar.tmpl", expect: "package bar"},
		{name: "file without define block", view: "raw.tmpl", data: []any{map[string]any{"Name": "Goravel"}}, expect: "raw Goravel"},
		{name: "shared data only", view: "greeting.tmpl", expect: "Hello, shared"},
		{name: "map overrides shared data", view: "greeting.tmpl", data: []any{map[string]string{"Name": "Goravel"}}, expect: "Hello, Goravel"},
		{name: "nil data", view: "greeting.tmpl", data: []any{nil}, expect: "Hello, shared"},
		{name: "nil pointer data", view: "greeting.tmpl", data: []any{(*page)(nil)}, expect: "Hello, shared"},
		{
			name:   "struct data with nested layout, partial and block",
			view:   "pages/home.tmpl",
			data:   []any{page{Title: "Home", Nav: "Menu", private: "hidden"}},
			expect: "<html><body><nav>Menu</nav><main><h1>Home</h1></main></body></html>",
		},
		{
			name:   "pointer to struct",
			view:   "pages/home.tmpl",
			data:   []any{&page{Title: "Home", Nav: "Menu"}},
			expect: "<html><body><nav>Menu</nav><main><h1>Home</h1></main></body></html>",
		},
		{
			name:   "values are HTML escaped",
			view:   "greeting.tmpl",
			data:   []any{map[string]any{"Name": "<script>"}},
			expect: "Hello, &lt;script&gt;",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			html, err := view.Make(test.view, test.data...).Render()
			require.NoError(t, err)
			assert.Equal(t, test.expect, html)
		})
	}

	t.Run("caller data is not modified", func(t *testing.T) {
		data := map[string]any{"Name": "Goravel"}
		_, err := view.Make("greeting.tmpl", data).With("Greeting", "Hi").Render()
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"Name": "Goravel"}, data)
	})

	t.Run("missing template", func(t *testing.T) {
		_, err := view.Make("missing.tmpl").Render()
		assert.ErrorIs(t, err, errors.ViewTemplateNotExist)
	})

	t.Run("invalid data", func(t *testing.T) {
		_, err := view.Make("greeting.tmpl", 1).Render()
		assert.ErrorIs(t, err, errors.ViewInvalidData)

		_, err = view.Make("greeting.tmpl", map[int]string{1: "a"}).Render()
		assert.ErrorIs(t, err, errors.ViewInvalidData)
	})
}

func TestTemplateRender_NoViews(t *testing.T) {
	setupAppViews(t, nil)

	_, err := NewView().Make("welcome.tmpl").Render()
	assert.ErrorIs(t, err, errors.ViewTemplateNotExist)
}

func TestTemplateRender_ParseError(t *testing.T) {
	setupAppViews(t, map[string]string{
		"broken.tmpl": `{{ define "broken.tmpl" }}{{ .Name }`,
	})

	_, err := NewView().Make("broken.tmpl").Render()
	assert.Error(t, err)
}

func TestTemplateRender_PointerFields(t *testing.T) {
	setupAppViews(t, map[string]string{
		"post.tmpl": `{{ define "post.tmpl" }}[{{ .Title }}][{{ .Author }}]{{ end }}`,
	})

	type post struct {
		Title  *string
		Author *string
	}

	title := "Hello"
	html, err := NewView().Make("post.tmpl", post{Title: &title}).Render()
	require.NoError(t, err)
	assert.Equal(t, "[Hello][]", html)
}

func TestTemplateRender_ParseErrorIsCached(t *testing.T) {
	dir := setupAppViews(t, map[string]string{
		"broken.tmpl": `{{ define "broken.tmpl" }}{{ .Name }`,
	})

	view := NewView()
	_, err := view.Make("broken.tmpl").Render()
	require.Error(t, err)

	// Fixing the file does not help until a new source is registered, the same way the
	// route drivers compile their views once.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.tmpl"), []byte(`{{ define "broken.tmpl" }}fixed{{ end }}`), 0o644))
	_, err = view.Make("broken.tmpl").Render()
	require.Error(t, err)

	view.LoadViewsFrom(t.TempDir())
	html, err := view.Make("broken.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "fixed", html)
}

func TestTemplateRender_ExecuteError(t *testing.T) {
	setupAppViews(t, map[string]string{
		"exec.tmpl": `{{ define "exec.tmpl" }}{{ template "missing.tmpl" }}{{ end }}`,
	})

	html, err := NewView().Make("exec.tmpl").Render()
	assert.ErrorContains(t, err, "missing.tmpl")
	assert.Empty(t, html)
}

func TestTemplateRender_RecompilesAfterRegisteringSource(t *testing.T) {
	setupAppViews(t, map[string]string{
		"foo.tmpl": `{{ define "foo.tmpl" }}app foo{{ end }}`,
	})

	view := NewView()
	_, err := view.Make("bar.tmpl").Render()
	require.ErrorIs(t, err, errors.ViewTemplateNotExist)

	view.LoadViewsFromFS(fstest.MapFS{
		"views/bar.tmpl": {Data: []byte(`{{ define "bar.tmpl" }}package bar{{ end }}`)},
	}, "views")

	html, err := view.Make("bar.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "package bar", html)

	pkgDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "baz.tmpl"), []byte(`{{ define "baz.tmpl" }}package baz{{ end }}`), 0o644))
	view.LoadViewsFrom(pkgDir)

	html, err = view.Make("baz.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "package baz", html)
}

func TestTemplateRender_Concurrent(t *testing.T) {
	setupAppViews(t, map[string]string{
		"greeting.tmpl": `{{ define "greeting.tmpl" }}Hello, {{ .Name }}{{ end }}`,
	})

	view := NewView()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			view.Share("Name", "Goravel")
			html, err := view.Make("greeting.tmpl", map[string]any{"Name": "Goravel"}).Render()
			assert.NoError(t, err)
			assert.Equal(t, "Hello, Goravel", html)
		}()
	}
	wg.Wait()
}

func TestTemplateWith(t *testing.T) {
	setupAppViews(t, map[string]string{
		"greeting.tmpl": `{{ define "greeting.tmpl" }}{{ .Greeting }}, {{ .Name }}{{ end }}`,
	})

	view := NewView()
	view.Share("Greeting", "Hello")
	view.Share("Name", "shared")

	template := view.Make("greeting.tmpl", map[string]any{"Name": "data"})
	assert.Equal(t, "greeting.tmpl", template.Name())
	assert.Equal(t, map[string]any{"Name": "data"}, template.Data())

	html, err := template.With("Name", "Goravel").With("Greeting", "Hi").Render()
	require.NoError(t, err)
	assert.Equal(t, "Hi, Goravel", html)
	assert.Equal(t, map[string]any{"Name": "Goravel", "Greeting": "Hi"}, template.Data())
	assert.Equal(t, "shared", view.Shared("Name"), "With must not change the shared data")

	html, err = view.Make("greeting.tmpl").With("Name", "Goravel").Render()
	require.NoError(t, err)
	assert.Equal(t, "Hello, Goravel", html)
}

func TestFirst(t *testing.T) {
	setupAppViews(t, map[string]string{
		"admin.tmpl":  `{{ define "admin.tmpl" }}admin {{ .Name }}{{ end }}`,
		"custom.tmpl": `{{ define "custom.tmpl" }}custom {{ .Name }}{{ end }}`,
	})

	view := NewView()
	data := map[string]any{"Name": "Goravel"}

	template := view.First([]string{"missing.tmpl", "custom.tmpl", "admin.tmpl"}, data)
	assert.Equal(t, "custom.tmpl", template.Name())
	html, err := template.Render()
	require.NoError(t, err)
	assert.Equal(t, "custom Goravel", html)

	html, err = view.First([]string{"missing.tmpl", "admin.tmpl"}, data).Render()
	require.NoError(t, err)
	assert.Equal(t, "admin Goravel", html)

	template = view.First([]string{"missing.tmpl", "other.tmpl"}, data)
	assert.Empty(t, template.Name())
	assert.Equal(t, data, template.Data())
	_, err = template.Render()
	assert.ErrorIs(t, err, errors.ViewNoneExist)
}

func BenchmarkTemplateRender(b *testing.B) {
	var page strings.Builder
	page.WriteString(`{{ define "page.tmpl" }}<html><body><h1>{{ .Title }}</h1><ul>`)
	for i := 0; i < 200; i++ {
		page.WriteString(`<li class="item">{{ .Title }} item</li>`)
	}
	page.WriteString(`</ul></body></html>{{ end }}`)

	setupAppViews(b, map[string]string{"page.tmpl": page.String()})

	view := NewView()
	view.Share("Name", "Goravel")
	data := map[string]any{"Title": "Home"}

	_, err := view.Make("page.tmpl", data).Render()
	require.NoError(b, err)

	b.Run("serial", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, _ = view.Make("page.tmpl", data).Render()
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = view.Make("page.tmpl", data).Render()
			}
		})
	})
}
