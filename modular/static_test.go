package modular

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
)

func TestStaticFS_Embedded(t *testing.T) {
	embedded := fstest.MapFS{
		"public/robots.txt":    {Data: []byte("User-agent: *")},
		"public/img/logo.png":  {Data: []byte("png")},
		"resources/secret.txt": {Data: []byte("secret")},
	}

	fsys, err := StaticFS("blog", "modules/blog", contractsmodular.Path{FS: embedded, Dir: "public"})
	require.NoError(t, err)

	content, err := fs.ReadFile(fsys, "robots.txt")
	require.NoError(t, err)
	assert.Equal(t, "User-agent: *", string(content))

	_, err = fs.ReadFile(fsys, "img/logo.png")
	assert.NoError(t, err)

	for _, name := range []string{".", "img"} {
		_, err := fsys.Open(name)
		assert.ErrorIs(t, err, fs.ErrNotExist, "directory %s must not be listable", name)
	}

	_, err = fsys.Open("../resources/secret.txt")
	assert.Error(t, err)
}

func TestStaticFS_Disk(t *testing.T) {
	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	dir := filepath.Join(support.RelativePath, "modules", "blog", "public")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "robots.txt"), []byte("disk"), 0o644))

	fsys, err := StaticFS("blog", "modules/blog", contractsmodular.Path{Dir: "public"})
	require.NoError(t, err)

	content, err := fs.ReadFile(fsys, "robots.txt")
	require.NoError(t, err)
	assert.Equal(t, "disk", string(content))

	_, err = StaticFS("blog", "modules/blog", contractsmodular.Path{Dir: "missing"})
	assert.ErrorIs(t, err, errors.ModularStaticDirUnreadable)
}

func TestStaticFS_Invalid(t *testing.T) {
	embedded := fstest.MapFS{"public/robots.txt": {Data: []byte("x")}}

	for _, dir := range []string{"..", "../shop", "public/../../x", "/etc"} {
		_, err := StaticFS("blog", "modules/blog", contractsmodular.Path{FS: embedded, Dir: dir})
		assert.ErrorIs(t, err, errors.ModularStaticDirInvalid, dir)
	}

	_, err := StaticFS("blog", "../blog", contractsmodular.Path{Dir: "public"})
	assert.ErrorIs(t, err, errors.ModularStaticDirInvalid, "the module directory must stay inside the project")

	_, err = StaticFS("blog", "modules/blog", contractsmodular.Path{FS: embedded, Dir: "public/robots.txt"})
	assert.ErrorIs(t, err, errors.ModularStaticDirUnreadable, "a file is not a directory")

	_, err = StaticFS("blog", "modules/blog", contractsmodular.Path{FS: embedded, Dir: "missing"})
	assert.ErrorIs(t, err, errors.ModularStaticDirUnreadable)
}

func TestCleanPrefix(t *testing.T) {
	tests := []struct {
		prefix, expected string
		ok               bool
	}{
		{"blog", "/blog", true},
		{"/blog/", "/blog", true},
		{"/blog/static", "/blog/static", true},
		{"/blog/../shop", "/shop", true},
		{"/", "", false},
		{"", "", false},
		{"..", "", false},
	}
	for _, tt := range tests {
		cleaned, ok := CleanPrefix(tt.prefix)
		assert.Equal(t, tt.ok, ok, tt.prefix)
		assert.Equal(t, tt.expected, cleaned, tt.prefix)
	}
}
