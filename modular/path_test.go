package modular

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

func TestResolvePath(t *testing.T) {
	t.Run("a package inside the application is its directory", func(t *testing.T) {
		dir, err := ResolvePath(&plainModule{"blog"}, "github.com/goravel/framework", DefaultDirectory)
		require.NoError(t, err)
		assert.Equal(t, "modular", dir)
	})

	t.Run("a package of another Go module goes under the module directory", func(t *testing.T) {
		dir, err := ResolvePath(&plainModule{"blog"}, "example.com/app", DefaultDirectory)
		require.NoError(t, err)
		assert.Equal(t, "modules/blog", dir)
	})

	t.Run("Located wins and is cleaned", func(t *testing.T) {
		for given, expected := range map[string]string{
			"modules/shop":          "modules/shop",
			"modules/shop/":         "modules/shop",
			"modules/../app/shop":   "app/shop",
			"./packages/shop/../x/": "packages/x",
		} {
			dir, err := ResolvePath(&locatedModule{plainModule{"shop"}, given}, "", DefaultDirectory)
			require.NoError(t, err, given)
			assert.Equal(t, expected, dir, given)
		}
	})

	t.Run("Located outside the project is refused", func(t *testing.T) {
		for _, given := range []string{"", " ", "..", "../shop", "modules/../../shop", "/srv/shop"} {
			_, err := ResolvePath(&locatedModule{plainModule{"shop"}, given}, "example.com/app", DefaultDirectory)
			assert.ErrorIs(t, err, errors.ModularModulePathInvalid, given)
		}
	})
}

func TestDerivePath(t *testing.T) {
	const main = "example.com/app"

	tests := []struct {
		name, pkg, mainModule, directory, expected string
		err                                        error
	}{
		{"blog", "example.com/app/modules/blog", main, "modules", "modules/blog", nil},
		{"blog", "example.com/app/features/blog-posts", main, "modules", "features/blog-posts", nil},
		{"root", "example.com/app", main, "modules", ".", nil},
		{"billing", "github.com/acme/billing", main, "modules", "modules/billing", nil},
		{"billing", "github.com/acme/billing/module", main, "app/modules/", "app/modules/billing", nil},
		// A prefix match must end at a path separator.
		{"billing", "example.com/application/billing", main, "modules", "modules/billing", nil},
		{"blog", "main", main, "modules", "", errors.ModularModulePathUnknown},
		{"blog", "", main, "modules", "", errors.ModularModulePathUnknown},
		{"blog", "example.com/app/blog", "", "modules", "", errors.ModularModulePathUnknown},
		{"a/b", "github.com/acme/billing", main, "modules", "", errors.ModularModulePathInvalid},
		{"..", "github.com/acme/billing", main, "modules", "", errors.ModularModulePathInvalid},
		{"", "github.com/acme/billing", main, "modules", "", errors.ModularModulePathInvalid},
		{"billing", "github.com/acme/billing", main, "../modules", "", errors.ModularModulePathInvalid},
	}
	for _, tt := range tests {
		dir, err := derivePath(tt.name, tt.pkg, tt.mainModule, tt.directory)
		if tt.err != nil {
			assert.ErrorIs(t, err, tt.err, tt.pkg)

			continue
		}
		require.NoError(t, err, tt.pkg)
		assert.Equal(t, tt.expected, dir, tt.pkg)
	}

	_, err := derivePath("blog", "main", main, "modules")
	assert.EqualError(t, err, `[modular] module blog: cannot derive its directory from package "main", implement Path() to set it`)
}
