package modular

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/foundation"
	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
)

type plainModule struct{ name string }

func (m *plainModule) Name() string { return m.name }

type locatedModule struct {
	plainModule
	path string
}

func (m *locatedModule) Path() string { return m.path }

type dependentModule struct {
	plainModule
	dependencies []string
}

func (m *dependentModule) Dependencies() []string { return m.dependencies }

type routedModule struct{ plainModule }

func (m *routedModule) Routes()                                        {}
func (m *routedModule) ServiceProviders() []foundation.ServiceProvider { return nil }

func names(modules []contractsmodule.Module) []string {
	var out []string
	for _, m := range modules {
		out = append(out, m.Name())
	}

	return out
}

func TestRegistry_All(t *testing.T) {
	registry := NewRegistry("")
	require.NoError(t, registry.Register(
		&dependentModule{plainModule{"admin"}, []string{"comments", "blog"}},
		&plainModule{"shop"},
		&dependentModule{plainModule{"comments"}, []string{"blog", "blog"}},
		&plainModule{"blog"},
	))

	assert.Equal(t, []string{"blog", "comments", "admin", "shop"}, names(registry.All()))
	assert.NoError(t, registry.Resolve())

	// A later registration is sorted too.
	require.NoError(t, registry.Register(&plainModule{"forum"}))
	assert.Equal(t, []string{"blog", "comments", "admin", "shop", "forum"}, names(registry.All()))
}

func TestRegistry_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		modules []contractsmodule.Module
		err     string
	}{
		{
			name: "missing dependency",
			modules: []contractsmodule.Module{
				&dependentModule{plainModule{"comments"}, []string{"blog"}},
				&plainModule{"shop"},
			},
			err: "[modular] module comments depends on module blog, which is not registered",
		},
		{
			name: "cycle",
			modules: []contractsmodule.Module{
				&plainModule{"shop"},
				&dependentModule{plainModule{"blog"}, []string{"comments"}},
				&dependentModule{plainModule{"comments"}, []string{"tags"}},
				&dependentModule{plainModule{"tags"}, []string{"blog"}},
			},
			err: "[modular] modules depend on each other: blog -> comments -> tags -> blog",
		},
		{
			name: "self dependency",
			modules: []contractsmodule.Module{
				&dependentModule{plainModule{"blog"}, []string{"blog"}},
			},
			err: "[modular] modules depend on each other: blog -> blog",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := NewRegistry("")
			require.NoError(t, registry.Register(test.modules...))

			assert.EqualError(t, registry.Resolve(), test.err)
			// Registration order is kept.
			assert.Equal(t, names(test.modules), names(registry.All()))
		})
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry("")
	require.NoError(t, registry.Register(&plainModule{"blog"}, nil))

	err := registry.Register(&plainModule{"shop"}, &plainModule{"blog"}, &plainModule{"forum"})
	assert.ErrorIs(t, err, errors.ModularModuleDuplicate)
	assert.EqualError(t, err, `[modular] module "blog" is registered more than once`)

	assert.Equal(t, []string{"blog", "shop"}, names(registry.All()))
}

func TestRegistry_RegisterInvalidName(t *testing.T) {
	for _, name := range []string{"", "Blog", "blog.admin", "blog_posts", "-blog", "blog--posts", "1blog"} {
		err := NewRegistry("").Register(&plainModule{name})
		assert.ErrorIs(t, err, errors.ModularModuleNameNotKebab, name)
	}
	assert.EqualError(t, NewRegistry("").Register(&plainModule{"Blog"}),
		`[modular] module *modular.plainModule is named "Blog", a module name is kebab case like blog or blog-posts`)
	assert.NoError(t, NewRegistry("").Register(&plainModule{"blog-posts2"}))
}

func TestRegistry_Path(t *testing.T) {
	registry := NewRegistry("")
	require.NoError(t, registry.Register(&plainModule{"blog"}, &locatedModule{plainModule{"shop"}, "app/shop"}))

	// Under go test the main module is the framework.
	dir, ok := registry.Path("blog")
	assert.True(t, ok)
	assert.Equal(t, "modular", dir)

	dir, ok = registry.Path("shop")
	assert.True(t, ok)
	assert.Equal(t, "app/shop", dir)

	_, ok = registry.Path("forum")
	assert.False(t, ok)

	err := registry.Register(&locatedModule{plainModule{"forum"}, "../forum"}, &plainModule{"news"})
	assert.ErrorIs(t, err, errors.ModularModulePathInvalid)
	assert.False(t, registry.Has("forum"))
	assert.False(t, registry.Has("news"))
}

func TestRegistry_FindHas(t *testing.T) {
	registry := NewRegistry("")
	blog := &plainModule{"blog"}
	require.NoError(t, registry.Register(blog))

	found, ok := registry.Find("blog")
	assert.True(t, ok)
	assert.Same(t, blog, found)
	assert.True(t, registry.Has("blog"))

	_, ok = registry.Find("shop")
	assert.False(t, ok)
	assert.False(t, registry.Has("shop"))
}

func TestRegistry_Capabilities(t *testing.T) {
	registry := NewRegistry("")

	capabilities := registry.Capabilities(&routedModule{plainModule{"blog"}})
	assert.Len(t, capabilities, 15)
	for label, ok := range capabilities {
		assert.Equal(t, label == "routes" || label == "services", ok, label)
	}

	for label, ok := range registry.Capabilities(&plainModule{"plain"}) {
		assert.False(t, ok, label)
	}
}

func TestProviding(t *testing.T) {
	routed := &routedModule{plainModule{"blog"}}
	modules := []contractsmodule.Module{&plainModule{"plain"}, routed}

	assert.Equal(t, []contractsmodule.ProvidesRoutes{routed}, Providing[contractsmodule.ProvidesRoutes](modules))
	assert.Empty(t, Providing[contractsmodule.ProvidesViews](modules))
}

func TestDefault(t *testing.T) {
	t.Cleanup(func() { SetDefault(nil) })
	assert.Nil(t, Default())

	registry := NewRegistry("")
	SetDefault(registry)
	assert.Same(t, registry, Default())
}

func TestLocal(t *testing.T) {
	// A module that sets its own path is in the application.
	assert.True(t, Local(&locatedModule{plainModule{"blog"}, "modules/blog"}))

	info, ok := debug.ReadBuildInfo()
	require.True(t, ok)
	_, thirdParty := DependencyOf(&plainModule{"blog"}, info)
	assert.Equal(t, !thirdParty, Local(&plainModule{"blog"}))
}
