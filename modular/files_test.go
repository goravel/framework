package modular

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
)

func TestFiles(t *testing.T) {
	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	dir := filepath.Join(support.RelativePath, "modules", "blog", "resources", "views")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.tmpl"), []byte("disk"), 0o644))

	embedded := fstest.MapFS{
		"resources/views/index.tmpl": {Data: []byte("default")},
		"templates/index.tmpl":       {Data: []byte("sub")},
		"index.tmpl":                 {Data: []byte("root")},
	}
	resolve := func(root string, files contractsmodule.Files) (fs.FS, bool, error) {
		return Files("blog", "Views()", root, ViewsDirectory, files)
	}
	index := func(t *testing.T, files contractsmodule.Files) string {
		fsys, ok, err := resolve("modules/blog", files)
		require.NoError(t, err)
		require.True(t, ok)
		content, err := fs.ReadFile(fsys, "index.tmpl")
		require.NoError(t, err)

		return string(content)
	}

	t.Run("filesystem", func(t *testing.T) {
		assert.Equal(t, "default", index(t, contractsmodule.Files{FS: embedded}), "the default directory")
		assert.Equal(t, "sub", index(t, contractsmodule.Files{FS: embedded, Dir: "templates"}))
		assert.Equal(t, "root", index(t, contractsmodule.Files{FS: embedded, Dir: "."}))

		_, ok, err := resolve("modules/blog", contractsmodule.Files{FS: fstest.MapFS{"index.tmpl": {Data: []byte("root")}}})
		assert.False(t, ok)
		assert.ErrorIs(t, err, errors.ModularFilesDirUnreadable, "the default directory is not optional")
	})

	t.Run("filesystem without a default directory is its root", func(t *testing.T) {
		fsys, ok, err := Files("blog", "Static()", "modules/blog", "", contractsmodule.Files{FS: embedded})
		require.NoError(t, err)
		assert.True(t, ok)
		_, err = fs.Stat(fsys, "templates/index.tmpl")
		assert.NoError(t, err)
	})

	t.Run("path is a directory of the module", func(t *testing.T) {
		assert.Equal(t, "disk", index(t, contractsmodule.Files{Dir: "resources/views"}))
	})

	t.Run("zero value is absent", func(t *testing.T) {
		fsys, ok, err := resolve("modules/blog", contractsmodule.Files{})
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Nil(t, fsys)
	})

	t.Run("directory outside", func(t *testing.T) {
		for _, dir := range []string{"..", "../shop", "resources/../../x", "/etc"} {
			for _, files := range []contractsmodule.Files{{Dir: dir}, {FS: embedded, Dir: dir}} {
				_, ok, err := resolve("modules/blog", files)
				assert.False(t, ok)
				assert.ErrorIs(t, err, errors.ModularFilesDirInvalid, dir)
			}
		}

		_, _, err := resolve("../blog", contractsmodule.Files{Dir: "resources/views"})
		assert.ErrorIs(t, err, errors.ModularFilesDirInvalid, "the module directory must stay inside the project")
	})

	t.Run("missing directory or a file", func(t *testing.T) {
		for _, dir := range []string{"missing", "resources/views/index.tmpl"} {
			for _, files := range []contractsmodule.Files{{Dir: dir}, {FS: embedded, Dir: dir}} {
				_, ok, err := resolve("modules/blog", files)
				assert.False(t, ok)
				assert.ErrorIs(t, err, errors.ModularFilesDirUnreadable, dir)
			}
		}
	})
}
