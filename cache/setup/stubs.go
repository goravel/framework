package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("cache", map[string]any{
		// Default Cache Store
		//
		// This option controls the default cache connection that gets used while
		// using this caching library. This connection is used when another is
		// not explicitly specified when executing a given caching function.
		"default": "memory",

		// Cache Stores
		//
		// Here you may define all the cache "stores" for your application as
		// well as their drivers. You may even define multiple stores for the
		// same cache driver to group types of items stored in your caches.
		// Available Drivers: "memory", "custom"
		"stores": map[string]any{
			"memory": map[string]any{
				"driver": "memory",
			},
		},

		// Cache Key Prefix
		//
		// When utilizing a RAM based store such as APC or Memcached, there might
		// be other applications utilizing the same cache. So, we'll specify a
		// value to get prefixed to all our keys, so we can avoid collisions.
		// Must: a-zA-Z0-9_-
		"prefix": config.GetString("APP_NAME", "goravel") + "_cache",
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) CacheFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/cache"
)

func Cache() cache.Cache {
	return App().MakeCache()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
