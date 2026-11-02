package view

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseSet(t *testing.T, files map[string]string) *template.Template {
	t.Helper()

	root := template.New("").Funcs(template.FuncMap{
		"asset": func(...string) string { return "" },
		"upper": strings.ToUpper,
	})
	for name, content := range files {
		_, err := root.New(name).Parse(content)
		require.NoError(t, err, name)
	}

	return root
}

func TestCalls(t *testing.T) {
	set := parseSet(t, map[string]string{
		"home.tmpl": `<head>{{ asset "a.css" "a.js" }}</head><p>{{ upper "shout" }}</p>`,
		"admin.tmpl": `{{ define "admin.tmpl" }}
			{{ if .admin }}{{ asset "admin.js" }}{{ else }}{{ asset "user.js" }}{{ end }}
			{{ range .items }}{{ asset "item.png" }}{{ end }}
			{{ with .x }}{{ "with.js" | asset }}{{ end }}
			{{ template "partial" (asset "partial-arg.js") }}
			{{ end }}
			{{ define "partial" }}{{ . }}{{ end }}`,
		"layout.tmpl": `{{ $entry := "var.js" }}{{ asset $entry }}{{ asset (printf "%s" .name) }}{{ asset "ok.js" .dyn }}{{ .Field "asset" }}{{ asset }}`,
	})

	assert.Equal(t, []Call{
		{Template: "admin.tmpl", Args: []string{"admin.js"}},
		{Template: "admin.tmpl", Args: []string{"user.js"}},
		{Template: "admin.tmpl", Args: []string{"item.png"}},
		{Template: "admin.tmpl", Args: []string{"with.js"}},
		{Template: "admin.tmpl", Args: []string{"partial-arg.js"}},
		{Template: "home.tmpl", Args: []string{"a.css", "a.js"}},
		{Template: "layout.tmpl", Dynamic: true},
		{Template: "layout.tmpl", Dynamic: true},
		{Template: "layout.tmpl", Dynamic: true},
		{Template: "layout.tmpl", Args: []string{}},
	}, Calls(set, "asset"))

	assert.Equal(t, []Call{{Template: "home.tmpl", Args: []string{"shout"}}}, Calls(set, "upper"))
	assert.Nil(t, Calls(set, "missing"))
	assert.Nil(t, Calls(nil, "asset"))
}

func TestCalls_AfterRender(t *testing.T) {
	// The escaper rewrites the trees on first execution; the calls are still found afterwards.
	set := parseSet(t, map[string]string{"page.tmpl": `<a href="{{ asset "a.svg" }}">{{ asset "b.js" }}</a>`})

	var out strings.Builder
	require.NoError(t, set.ExecuteTemplate(&out, "page.tmpl", nil))

	assert.Equal(t, []Call{
		{Template: "page.tmpl", Args: []string{"a.svg"}},
		{Template: "page.tmpl", Args: []string{"b.js"}},
	}, Calls(set, "asset"))
}
