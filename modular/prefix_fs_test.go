package modular_test

import (
	"bytes"
	"io"
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

// stubFS opens file for every name, or fails with err.
type stubFS struct {
	file fs.File
	err  error
}

func (r stubFS) Open(string) (fs.File, error) { return r.file, r.err }

// stubFile is a file that cannot be listed; Stat fails with err.
type stubFile struct{ err error }

func (r stubFile) Stat() (fs.FileInfo, error) { return nil, r.err }
func (r stubFile) Read([]byte) (int, error)   { return 0, io.EOF }
func (r stubFile) Close() error               { return nil }

func TestPrefixFS_Root(t *testing.T) {
	fsys := modular.NewPrefixFS("blog", blogViews())

	root, err := fsys.Open(".")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, root.Close()) })

	info, err := root.Stat()
	require.NoError(t, err)
	assert.Equal(t, ".", info.Name())
	assert.Zero(t, info.Size())
	assert.Equal(t, fs.ModeDir|0o555, info.Mode())
	assert.True(t, info.IsDir())
	assert.Nil(t, info.Sys())

	mount, err := fs.Stat(fsys, "blog")
	require.NoError(t, err)
	assert.Equal(t, mount.ModTime(), info.ModTime())

	_, err = root.Read(make([]byte, 1))
	assert.ErrorIs(t, err, fs.ErrInvalid)

	// More entries are asked for than the single mount point.
	dir, ok := root.(fs.ReadDirFile)
	require.True(t, ok)
	entries, err := dir.ReadDir(5)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "blog", entries[0].Name())
	_, err = dir.ReadDir(5)
	assert.ErrorIs(t, err, io.EOF)
}

func TestPrefixFS_InvalidPaths(t *testing.T) {
	fsys := modular.NewPrefixFS("blog", blogViews())

	_, err := fs.ReadDir(fsys, "../blog")
	assert.ErrorIs(t, err, fs.ErrInvalid)

	_, err = fs.ReadFile(fsys, "../blog/index.tmpl")
	assert.ErrorIs(t, err, fs.ErrInvalid)

	// Directories are not files.
	for _, name := range []string{".", "blog"} {
		_, err = fs.ReadFile(fsys, name)
		assert.ErrorIs(t, err, fs.ErrInvalid, name)
	}
}

func TestPrefixFS_UnreadableInner(t *testing.T) {
	fsys := modular.NewPrefixFS("blog", stubFS{err: fs.ErrPermission})

	_, err := fsys.Open(".")
	assert.ErrorIs(t, err, fs.ErrPermission)
	_, err = fsys.Open("blog")
	assert.ErrorIs(t, err, fs.ErrPermission)
	_, err = fs.ReadDir(fsys, ".")
	assert.ErrorIs(t, err, fs.ErrPermission)
	_, err = fs.Stat(fsys, ".")
	assert.ErrorIs(t, err, fs.ErrPermission)
	_, err = fs.Stat(fsys, "blog")
	assert.ErrorIs(t, err, fs.ErrPermission)
}

func TestPrefixFS_MountFile(t *testing.T) {
	// The inner root cannot be described or listed.
	fsys := modular.NewPrefixFS("blog", stubFS{file: stubFile{err: fs.ErrPermission}})

	mount, err := fsys.Open("blog")
	require.NoError(t, err)

	_, err = mount.Stat()
	assert.ErrorIs(t, err, fs.ErrPermission)

	dir, ok := mount.(fs.ReadDirFile)
	require.True(t, ok)
	_, err = dir.ReadDir(-1)
	assert.ErrorIs(t, err, fs.ErrInvalid)
}
