package modular_test

import (
	"bytes"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/view"
)

func blogViews() fstest.MapFS {
	return fstest.MapFS{
		"index.tmpl":        {Data: []byte(`{{ template "blog/layouts/base.tmpl" . }}`)},
		"layouts/base.tmpl": {Data: []byte(`<h1>{{ .title }}</h1>`)},
	}
}

func TestPrefixFS_TestFS(t *testing.T) {
	assert.NoError(t, fstest.TestFS(modular.NewPrefixFS("blog", blogViews()), "blog/index.tmpl", "blog/layouts/base.tmpl"))
}

func TestPrefixFS_Paths(t *testing.T) {
	fsys := modular.NewPrefixFS("blog", blogViews())

	entries, err := fs.ReadDir(fsys, ".")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "blog", entries[0].Name())
	assert.True(t, entries[0].IsDir())

	info, err := fs.Stat(fsys, "blog")
	require.NoError(t, err)
	assert.Equal(t, "blog", info.Name())
	assert.True(t, info.IsDir())

	content, err := fs.ReadFile(fsys, "blog/layouts/base.tmpl")
	require.NoError(t, err)
	assert.Equal(t, `<h1>{{ .title }}</h1>`, string(content))

	for _, name := range []string{"index.tmpl", "shop/index.tmpl", "blogs/index.tmpl", "blog/missing.tmpl"} {
		_, err := fsys.Open(name)
		assert.ErrorIs(t, err, fs.ErrNotExist, name)
	}

	_, err = fsys.Open("../blog")
	assert.ErrorIs(t, err, fs.ErrInvalid)
}

// Views load and compile from the mounted FS, named under the module.
func TestPrefixFS_Views(t *testing.T) {
	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	fsys := modular.NewPrefixFS("blog", blogViews())

	registry := view.NewView()
	assert.NotPanics(t, func() { registry.LoadViewsFromFS(fsys, ".") })
	assert.True(t, registry.Exists("blog/index.tmpl"))
	assert.False(t, registry.Exists("blog"))
	assert.False(t, registry.Exists("index.tmpl"))

	tmpl, err := view.Compile(view.CompileOptions{FS: registry.RegisteredViewFS()})
	require.NoError(t, err)

	var out bytes.Buffer
	require.NoError(t, tmpl.ExecuteTemplate(&out, "blog/index.tmpl", map[string]any{"title": "Blog"}))
	assert.Equal(t, "<h1>Blog</h1>", out.String())
}
