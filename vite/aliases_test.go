package vite

import (
	"html/template"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

const aliasManifest = `{
  "goravel:aliases": {"@": "resources/js", "@blog": "modules/blog/resources/js", "@blog-assets": "modules/blog/resources/assets"},
  "resources/js/app.js": {"file": "assets/app-1.js", "src": "resources/js/app.js", "isEntry": true},
  "modules/blog/resources/js/app.js": {"file": "assets/blog-2.js", "src": "modules/blog/resources/js/app.js", "isEntry": true},
  "modules/blog/resources/assets/x.jpg": {"file": "assets/x-3.jpg", "src": "modules/blog/resources/assets/x.jpg"}
}`

var testAliases = map[string]string{"@": "resources/js", "@blog": "modules/blog/resources/js"}

func TestResolveEntry(t *testing.T) {
	tests := []struct {
		entry    string
		expected string
		err      error
	}{
		{entry: "a/./b", expected: "a/b"},
		{entry: "/resources/js/app.js", expected: "resources/js/app.js"},
		{entry: "@/app.js", expected: "resources/js/app.js"},
		{entry: "@blog/app.js", expected: "modules/blog/resources/js/app.js"},
		{entry: "/@blog/app.js", expected: "modules/blog/resources/js/app.js"},
		{entry: "@blog", expected: "modules/blog/resources/js"},
		{entry: "@blog/../assets/x.jpg", expected: "modules/blog/resources/assets/x.jpg"},
		{entry: "..", err: errors.ViteEntryOutsideProject.Args("..", "..")},
		{entry: "../x", err: errors.ViteEntryOutsideProject.Args("../x", "../x")},
		{entry: "@blog/../../../../../.env", err: errors.ViteEntryOutsideProject.Args("@blog/../../../../../.env", "../.env")},
		{entry: "@/../../../x", err: errors.ViteEntryOutsideProject.Args("@/../../../x", "../x")},
		// Leaving the alias directory is fine while the path stays in the project.
		{entry: "@/../../x", expected: "x"},
		// "@" only matches at a path boundary, so it never captures another alias.
		{entry: "@shop/app.js", err: errors.ViteAliasNotFound.Args("@shop/app.js")},
		{entry: "@blogs/app.js", err: errors.ViteAliasNotFound.Args("@blogs/app.js")},
	}

	for _, test := range tests {
		t.Run(test.entry, func(t *testing.T) {
			resolved, err := resolveEntry(test.entry, testAliases)
			if test.err != nil {
				assert.EqualError(t, err, test.err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, resolved)
		})
	}

	// Without a table plain paths still resolve.
	resolved, err := resolveEntry("resources/js/app.js", nil)
	require.NoError(t, err)
	assert.Equal(t, "resources/js/app.js", resolved)

	_, err = resolveEntry("@/app.js", nil)
	assert.EqualError(t, err, errors.ViteAliasNotFound.Args("@/app.js").Error())
}

func TestAliases_Build(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), aliasManifest)
	writeFile(t, filepath.Join(public, "build", "assets", "blog-2.js"), "blog()")
	vite := newTestVite(t, "")

	tags, err := vite.Tags("@/app.js", "@blog/app.js", "modules/blog/resources/js/app.js")
	require.NoError(t, err)
	assert.Equal(t, template.HTML(
		`<script type="module" src="/build/assets/app-1.js"></script>`+
			`<script type="module" src="/build/assets/blog-2.js"></script>`,
	), tags)

	for _, asset := range []string{"@blog-assets/x.jpg", "@blog/../assets/x.jpg", "modules/blog/resources/assets/x.jpg"} {
		url, err := vite.Asset(asset)
		require.NoError(t, err)
		assert.Equal(t, "/build/assets/x-3.jpg", url)
	}

	content, err := vite.Content("@blog/app.js")
	require.NoError(t, err)
	assert.Equal(t, "blog()", content)

	_, err = vite.Tags("@shop/app.js")
	assert.EqualError(t, err, errors.ViteAliasNotFound.Args("@shop/app.js").Error())
	_, err = vite.Asset("@blog/../../../../../.env")
	assert.EqualError(t, err, errors.ViteEntryOutsideProject.Args("@blog/../../../../../.env", "../.env").Error())
	_, err = vite.Content("../x")
	assert.EqualError(t, err, errors.ViteEntryOutsideProject.Args("../x", "../x").Error())

	// The table is not a chunk.
	_, err = vite.Asset("goravel:aliases")
	assert.EqualError(t, err, errors.ViteEntryNotFound.Args("goravel:aliases").Error())
}

func TestAliases_ManifestWithoutTable(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), sampleManifest)
	vite := newTestVite(t, "")

	url, err := vite.Asset("resources/images/logo.svg")
	require.NoError(t, err)
	assert.Equal(t, "/build/assets/logo-11223344.svg", url)

	_, err = vite.Tags("@/app.js")
	assert.EqualError(t, err, errors.ViteAliasNotFound.Args("@/app.js").Error())
}

func TestAliases_ManifestInvalidTable(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), `{"goravel:aliases": ["@"]}`)

	_, err := newTestVite(t, "").Tags("resources/js/app.js")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is invalid")
}

func TestAliases_Hot(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")
	writeFile(t, filepath.Join(public, "hot"), `{"url":"http://localhost:5173/","aliases":{"@":"resources/js","@blog":"modules/blog/resources/js"}}`+"\n")

	tags, err := vite.Tags("@blog/app.js", "@/app.js", "resources/css/app.css")
	require.NoError(t, err)
	assert.Equal(t, template.HTML(
		`<script type="module" src="http://localhost:5173/@vite/client"></script>`+
			`<script type="module" src="http://localhost:5173/modules/blog/resources/js/app.js"></script>`+
			`<script type="module" src="http://localhost:5173/resources/js/app.js"></script>`+
			`<link rel="stylesheet" href="http://localhost:5173/resources/css/app.css">`,
	), tags)

	url, err := vite.Asset("@blog/../assets/x.jpg")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:5173/modules/blog/resources/assets/x.jpg", url)

	_, err = vite.Asset("@blog/../../../../../.env")
	assert.EqualError(t, err, errors.ViteEntryOutsideProject.Args("@blog/../../../../../.env", "../.env").Error())
	_, err = vite.Tags("@shop/app.js")
	assert.EqualError(t, err, errors.ViteAliasNotFound.Args("@shop/app.js").Error())

	assert.Contains(t, string(vite.ReactRefresh()), "'http://localhost:5173/@react-refresh'")
	assert.True(t, vite.IsRunningHot())

	// A bare URL, as an older plugin writes it, has no aliases.
	writeFile(t, filepath.Join(public, "hot"), "http://localhost:5173")
	url, err = vite.Asset("resources/images/logo.svg")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:5173/resources/images/logo.svg", url)
	_, err = vite.Asset("@/app.js")
	assert.EqualError(t, err, errors.ViteAliasNotFound.Args("@/app.js").Error())
}
