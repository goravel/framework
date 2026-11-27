package module

import (
	contractsmodule "github.com/goravel/framework/contracts/module"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

func TestStaticFS(t *testing.T) {
	embedded := fstest.MapFS{
		"public/robots.txt":    {Data: []byte("User-agent: *")},
		"public/img/logo.png":  {Data: []byte("png")},
		"resources/secret.txt": {Data: []byte("secret")},
	}

	fsys, ok, err := StaticFS("blog", contractsmodule.FS(embedded).Sub("public"))
	require.NoError(t, err)
	require.True(t, ok)

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

	_, ok, err = StaticFS("blog", contractsmodule.FS(embedded).Sub("../shop"))
	assert.False(t, ok)
	assert.ErrorIs(t, err, errors.ModuleFilesDirInvalid)

	_, ok, err = StaticFS("blog", contractsmodule.Path("public"))
	assert.False(t, ok)
	assert.ErrorIs(t, err, errors.ModuleFilesFSRequired)
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

// statFailFS opens files that cannot be described.
type statFailFS struct{}

func (statFailFS) Open(string) (fs.File, error) { return statFailFile{}, nil }

type statFailFile struct{}

func (statFailFile) Stat() (fs.FileInfo, error) { return nil, fs.ErrPermission }
func (statFailFile) Read([]byte) (int, error)   { return 0, fs.ErrPermission }
func (statFailFile) Close() error               { return nil }

func TestNoListingFS_StatFails(t *testing.T) {
	_, err := noListingFS{statFailFS{}}.Open("robots.txt")
	assert.ErrorIs(t, err, fs.ErrPermission)
}
