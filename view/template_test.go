package view

import (
	"fmt"
	"html/template"
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
	assert.ErrorContains(t, err, "broken.tmpl")
	assert.NotErrorIs(t, err, errors.ViewTemplateNotExist)
}

// TestTemplateRender_PointerFields pins the struct field handling toMap does for pointers.
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

func TestTemplateRender_UsesRegisteredFuncs(t *testing.T) {
	setupAppViews(t, map[string]string{
		"loud.tmpl": `{{ define "loud.tmpl" }}{{ shout .Name }}{{ end }}`,
	})

	view := NewView()
	data := map[string]any{"Name": "goravel"}

	// The function is unknown, so the view does not parse at all.
	_, err := view.Make("loud.tmpl", data).Render()
	require.Error(t, err)
	assert.ErrorContains(t, err, "shout")

	// Registering the function drops the compiled set, so the same view parses now.
	view.AddFuncs(template.FuncMap{"shout": strings.ToUpper})

	html, err := view.Make("loud.tmpl", data).Render()
	require.NoError(t, err)
	assert.Equal(t, "GORAVEL", html)
}

func TestTemplateRender_Concurrent(t *testing.T) {
	setupAppViews(t, map[string]string{
		"greeting.tmpl": `{{ define "greeting.tmpl" }}Hello, {{ .Name }}{{ end }}`,
	})

	view := NewView()
	view.Share("Name", "Goravel")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			html, err := view.Make("greeting.tmpl", map[string]any{"Name": "Goravel"}).Render()
			assert.NoError(t, err)
			assert.Equal(t, "Hello, Goravel", html)
		}()
	}
	wg.Wait()
}

func TestTemplateRender_ConcurrentWithRegistration(t *testing.T) {
	setupAppViews(t, map[string]string{
		"greeting.tmpl": `{{ define "greeting.tmpl" }}Hello, {{ .Name }}{{ end }}`,
	})

	view := NewView()

	// Both groups wait on one channel, so registrations overlap with the renders instead of
	// possibly finishing before the first render starts.
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			html, err := view.Make("greeting.tmpl", map[string]any{"Name": "Goravel"}).Render()
			assert.NoError(t, err)
			assert.Equal(t, "Hello, Goravel", html)
		}()
	}

	// Registering a source drops the compiled set, which the renders above are reading.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			view.LoadViewsFromFS(fstest.MapFS{
				fmt.Sprintf("views/pkg%d.tmpl", i): {Data: []byte(`package`)},
			}, "views")
		}(i)
	}

	close(start)
	wg.Wait()

	// Every registration must have reached the next compile.
	for i := 0; i < 5; i++ {
		html, err := view.Make(fmt.Sprintf("pkg%d.tmpl", i)).Render()
		require.NoError(t, err)
		assert.Equal(t, "package", html)
	}
}

func TestTemplateWith(t *testing.T) {
	setupAppViews(t, map[string]string{
		"greeting.tmpl": `{{ define "greeting.tmpl" }}{{ .Greeting }}, {{ .Name }}{{ end }}`,
	})

	view := NewView()
	view.Share("Greeting", "Hello")
	view.Share("Name", "shared")

	tmpl := view.Make("greeting.tmpl", map[string]any{"Name": "data"})
	assert.Equal(t, "greeting.tmpl", tmpl.Name())
	assert.Equal(t, map[string]any{"Name": "data"}, tmpl.Data())

	html, err := tmpl.With("Name", "Goravel").With("Greeting", "Hi").Render()
	require.NoError(t, err)
	assert.Equal(t, "Hi, Goravel", html)
	assert.Equal(t, map[string]any{"Name": "Goravel", "Greeting": "Hi"}, tmpl.Data())
	assert.Equal(t, "shared", view.Shared("Name"), "With must not change the shared data")

	html, err = view.Make("greeting.tmpl").With("Name", "Goravel").Render()
	require.NoError(t, err)
	assert.Equal(t, "Hello, Goravel", html)
}

func TestTemplateData_IsACopy(t *testing.T) {
	setupAppViews(t, map[string]string{"greet.tmpl": `{{ define "greet.tmpl" }}{{ .Name }}{{ end }}`})

	view := NewView()
	tmpl := view.Make("greet.tmpl", map[string]any{"Name": "Goravel"})

	data := tmpl.Data()
	data["Name"] = "changed"

	assert.Equal(t, "Goravel", tmpl.Data()["Name"])
	html, err := tmpl.Render()
	require.NoError(t, err)
	assert.Equal(t, "Goravel", html)
}

func TestTemplateRender_MapPointerData(t *testing.T) {
	setupAppViews(t, map[string]string{"greet.tmpl": `{{ define "greet.tmpl" }}[{{ .Name }}]{{ end }}`})

	view := NewView()

	data := map[string]any{"Name": "Goravel"}
	html, err := view.Make("greet.tmpl", &data).Render()
	require.NoError(t, err)
	assert.Equal(t, "[Goravel]", html)

	// The copy is taken when the template is built, so a later change is not picked up.
	tmpl := view.Make("greet.tmpl", &data)
	data["Name"] = "changed"
	html, err = tmpl.Render()
	require.NoError(t, err)
	assert.Equal(t, "[Goravel]", html)

	html, err = view.Make("greet.tmpl", (*map[string]any)(nil)).Render()
	require.NoError(t, err)
	assert.Equal(t, "[]", html)
}

func TestFirst(t *testing.T) {
	setupAppViews(t, map[string]string{
		"admin.tmpl":  `{{ define "admin.tmpl" }}admin {{ .Name }}{{ end }}`,
		"custom.tmpl": `{{ define "custom.tmpl" }}custom {{ .Name }}{{ end }}`,
	})

	view := NewView()
	data := map[string]any{"Name": "Goravel"}

	tmpl := view.First([]string{"missing.tmpl", "custom.tmpl", "admin.tmpl"}, data)
	assert.Equal(t, "custom.tmpl", tmpl.Name())
	html, err := tmpl.Render()
	require.NoError(t, err)
	assert.Equal(t, "custom Goravel", html)

	html, err = view.First([]string{"missing.tmpl", "admin.tmpl"}, data).Render()
	require.NoError(t, err)
	assert.Equal(t, "admin Goravel", html)

	tmpl = view.First([]string{"missing.tmpl", "other.tmpl"}, data)
	assert.Empty(t, tmpl.Name())
	assert.Equal(t, data, tmpl.Data())
	_, err = tmpl.Render()
	assert.ErrorIs(t, err, errors.ViewNoneExist)
}

func TestFirst_SkipsViewThatCannotRender(t *testing.T) {
	setupAppViews(t, map[string]string{
		// The file exists under this path but names its template something else, so it is not
		// addressable as "partials/menu.tmpl" and First has to carry on to the next candidate.
		"partials/menu.tmpl": `{{ define "nav.tmpl" }}nav{{ end }}`,
		"fallback.tmpl":      `{{ define "fallback.tmpl" }}fallback{{ end }}`,
	})

	view := NewView()
	require.True(t, view.Exists("partials/menu.tmpl"))

	tmpl := view.First([]string{"partials/menu.tmpl", "fallback.tmpl"})
	assert.Equal(t, "fallback.tmpl", tmpl.Name())

	html, err := tmpl.Render()
	require.NoError(t, err)
	assert.Equal(t, "fallback", html)
}

func TestFirst_ReportsCompileErrorBeforeNoneExist(t *testing.T) {
	setupAppViews(t, map[string]string{
		"broken.tmpl": `{{ define "broken.tmpl" }}{{ .Name }`,
		"good.tmpl":   `{{ define "good.tmpl" }}good{{ end }}`,
	})

	// Nothing compiles, so First cannot tell which views are renderable. It has to pick a
	// candidate and let Render report the parse error, rather than claiming that none of the
	// views exist.
	_, err := NewView().First([]string{"good.tmpl"}).Render()
	assert.ErrorContains(t, err, "broken.tmpl")
	assert.NotErrorIs(t, err, errors.ViewNoneExist)
}

func TestFirst_ReportsInvalidDataBeforeMissingViews(t *testing.T) {
	setupAppViews(t, map[string]string{"admin.tmpl": `{{ define "admin.tmpl" }}admin{{ end }}`})

	view := NewView()

	_, err := view.First([]string{"missing.tmpl"}, map[int]string{1: "one"}).Render()
	assert.ErrorIs(t, err, errors.ViewInvalidData)

	// The template has no name, so the error names the candidates instead.
	assert.ErrorContains(t, err, "missing.tmpl")

	// An empty candidate list has no names to offer, but the name slot must not be left blank.
	_, err = view.First(nil, 1).Render()
	assert.ErrorIs(t, err, errors.ViewInvalidData)
	assert.NotContains(t, err.Error(), "  ", "the message must fill the name slot")

	_, err = view.First(nil).Render()
	assert.ErrorIs(t, err, errors.ViewNoneExist)

	// A template that matched nothing still collects data; it just cannot be rendered.
	tmpl := view.First([]string{"missing.tmpl"}).With("Name", "Goravel")
	assert.Equal(t, map[string]any{"Name": "Goravel"}, tmpl.Data())
	_, err = tmpl.Render()
	assert.ErrorIs(t, err, errors.ViewNoneExist)
}

func TestTemplateRender_DefineOnlyFileIsNotAView(t *testing.T) {
	setupAppViews(t, map[string]string{
		"layouts/app.tmpl": `{{ define "layouts/app.tmpl" }}<html>{{ .Title }}</html>{{ end }}`,
	})

	view := NewView()

	html, err := view.Make("layouts/app.tmpl", map[string]any{"Title": "Home"}).Render()
	require.NoError(t, err)
	assert.Equal(t, "<html>Home</html>", html)

	// Every template set carries an unnamed root, and a name no source ever claimed is not a
	// view either. Neither is renderable.
	for _, name := range []string{"app.tmpl", ""} {
		html, err = view.Make(name).Render()
		assert.ErrorIs(t, err, errors.ViewTemplateNotExist, name)
		assert.Empty(t, html)
	}
}

func TestTemplateRender_EmptyTemplateIsNotAView(t *testing.T) {
	setupAppViews(t, map[string]string{
		"empty.tmpl": `{{ define "empty.name" }}{{ end }}`,
		// The layout is walked before the page here, so the empty block placeholder is claimed
		// first and the page's content has to make the name renderable afterwards.
		"a_layout.tmpl": `{{ define "a_layout.tmpl" }}[{{ block "content" . }}{{ end }}|{{ block "aside" . }}{{ end }}]{{ end }}`,
		"b_page.tmpl":   `{{ define "b_page.tmpl" }}{{ template "a_layout.tmpl" . }}{{ end }}{{ define "content" }}page{{ end }}`,
		"fallback.tmpl": `{{ define "fallback.tmpl" }}fallback{{ end }}`,
	})

	view := NewView()

	for _, name := range []string{"empty.name", "aside"} {
		html, err := view.Make(name).Render()
		assert.ErrorIs(t, err, errors.ViewTemplateNotExist, name)
		assert.Empty(t, html)
	}

	// Empty templates stay in the set, so a layout still resolves its placeholders.
	html, err := view.Make("b_page.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "[page|]", html)

	html, err = view.Make("content").Render()
	require.NoError(t, err)
	assert.Equal(t, "page", html)

	tmpl := view.First([]string{"empty.name", "aside", "fallback.tmpl"})
	assert.Equal(t, "fallback.tmpl", tmpl.Name())
}

func TestTemplateRender_BlockPlaceholderDoesNotWinOverPage(t *testing.T) {
	setupAppViews(t, map[string]string{
		// The page is walked first, so the layout's empty {{ block }} must not replace the
		// content the page just filled in.
		"page.tmpl":    `{{ define "page.tmpl" }}{{ template "zlayout.tmpl" . }}{{ end }}{{ define "content" }}page{{ end }}`,
		"zlayout.tmpl": `{{ define "zlayout.tmpl" }}[{{ block "content" . }}{{ end }}]{{ end }}`,
	})

	html, err := NewView().Make("page.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "[page]", html)
}

func TestTemplateRender_PackageDoesNotOverrideApplication(t *testing.T) {
	setupAppViews(t, map[string]string{
		// The trim marker keeps the define out of the rendered output; it still names a template.
		"primary.tmpl": `{{- define "primary.tmpl" -}}app primary{{- end -}}`,
		"shared.tmpl":  `{{- define "shared.tmpl" -}}app shared{{- end -}}`,
	})

	// Both defines live in one file whose own name the application owns.
	pkg := fstest.MapFS{
		"views/shared.tmpl": {Data: []byte(`{{ define "shared.tmpl" }}package shared{{ end }}` +
			`{{ define "other.tmpl" }}package other{{ end }}`)},
	}

	view := NewView()
	view.LoadViewsFromFS(pkg, "views")

	html, err := view.Make("primary.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "app primary", html)

	html, err = view.Make("shared.tmpl").Render()
	require.NoError(t, err)
	assert.Equal(t, "app shared", html)

	// Compile skips the colliding package file whole, so the sibling define in it is gone too.
	_, err = view.Make("other.tmpl").Render()
	assert.ErrorIs(t, err, errors.ViewTemplateNotExist)
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
			if _, err := view.Make("page.tmpl", data).Render(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if _, err := view.Make("page.tmpl", data).Render(); err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}
