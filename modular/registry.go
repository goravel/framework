// Package modular implements the module registry and module filesystems. It
// imports no framework implementation, so foundation can import it.
package modular

import (
	"slices"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/env"
)

var _ modular.Registry = (*Registry)(nil)

// Registry keeps modules in registration order.
type Registry struct {
	modules []modular.Module
	byName  map[string]modular.Module
	paths   map[string]string

	// mainModule is the app's Go module path; directory holds third-party modules.
	mainModule string
	directory  string
}

// NewRegistry places third-party modules under directory (modular.path), "" for
// DefaultDirectory.
func NewRegistry(directory string) *Registry {
	if directory == "" {
		directory = DefaultDirectory
	}

	return &Registry{
		byName:     make(map[string]modular.Module),
		paths:      make(map[string]string),
		mainModule: env.MainPath(),
		directory:  directory,
	}
}

// Register stops at the first error; earlier modules stay registered.
func (r *Registry) Register(modules ...modular.Module) error {
	for _, module := range modules {
		if module == nil {
			continue
		}

		name := module.Name()
		if _, ok := r.byName[name]; ok {
			return errors.ModularModuleDuplicate.Args(name)
		}

		dir, err := ResolvePath(module, r.mainModule, r.directory)
		if err != nil {
			return err
		}

		r.byName[name] = module
		r.paths[name] = dir
		r.modules = append(r.modules, module)
	}

	return nil
}

// All sorts non-negative orders, then negative ones, ascending; ties keep registration order.
func (r *Registry) All() []modular.Module {
	sorted := slices.Clone(r.modules)
	slices.SortStableFunc(sorted, func(a, b modular.Module) int {
		return compareOrder(order(a), order(b))
	})

	return sorted
}

func (r *Registry) Find(name string) (modular.Module, bool) {
	module, ok := r.byName[name]

	return module, ok
}

func (r *Registry) Has(name string) bool {
	_, ok := r.byName[name]

	return ok
}

func (r *Registry) Path(name string) (string, bool) {
	dir, ok := r.paths[name]

	return dir, ok
}

func (r *Registry) Capabilities(module modular.Module) map[string]bool {
	return map[string]bool{
		"services":     implements[foundation.ProvidesServices](module),
		"config":       implements[modular.ProvidesConfig](module),
		"middleware":   implements[modular.ProvidesGlobalMiddleware](module),
		"routes":       implements[modular.ProvidesRoutes](module),
		"api":          implements[modular.ProvidesApiRoutes](module),
		"migrations":   implements[modular.ProvidesMigrations](module),
		"seeders":      implements[modular.ProvidesSeeders](module),
		"commands":     implements[modular.ProvidesCommands](module),
		"jobs":         implements[modular.ProvidesJobs](module),
		"events":       implements[modular.ProvidesEvents](module),
		"schedule":     implements[modular.ProvidesSchedule](module),
		"views":        implements[modular.ProvidesViews](module),
		"translations": implements[modular.ProvidesTranslations](module),
		"vite-entries": implements[modular.ProvidesViteEntries](module),
		"vite-aliases": implements[modular.ProvidesViteAlias](module),
		"static":       implements[modular.ProvidesStatic](module),
		"sources":      implements[modular.ProvidesSources](module),
	}
}

// Providing returns the modules implementing T, as T.
func Providing[T any](modules []modular.Module) []T {
	var out []T
	for _, module := range modules {
		if p, ok := module.(T); ok {
			out = append(out, p)
		}
	}

	return out
}

func implements[T any](module modular.Module) bool {
	_, ok := module.(T)

	return ok
}

func order(module modular.Module) int {
	if o, ok := module.(modular.Ordered); ok {
		return o.Order()
	}

	return 0
}

// compareOrder sorts non-negative orders before negative ones, each ascending.
func compareOrder(a, b int) int {
	if (a < 0) != (b < 0) {
		if a < 0 {
			return 1
		}

		return -1
	}

	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
