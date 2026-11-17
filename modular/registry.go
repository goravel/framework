// Package modular implements the module registry and module filesystems. It
// imports no framework implementation, so foundation can import it.
package modular

import (
	"regexp"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/goravel/framework/contracts/foundation"
	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/env"
)

var _ contractsmodule.Registry = (*Registry)(nil)

// namePattern is a kebab-case module name, e.g. "blog" or "blog-posts".
var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// ValidName reports whether name is a kebab-case module name.
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

var (
	defaultMu       sync.RWMutex
	defaultRegistry contractsmodule.Registry
)

// SetDefault records the application's registry for code that cannot reach the
// container (support/console). Build() calls it.
func SetDefault(registry contractsmodule.Registry) {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	defaultRegistry = registry
}

// Default returns the application's registry, nil before Build().
func Default() contractsmodule.Registry {
	defaultMu.RLock()
	defer defaultMu.RUnlock()

	return defaultRegistry
}

// Local reports whether module's sources are in the application, not in a dependency.
func Local(module contractsmodule.Module) bool {
	info, _ := debug.ReadBuildInfo()
	_, thirdParty := DependencyOf(module, info)

	return !thirdParty
}

// Registry keeps modules in registration order.
type Registry struct {
	modules []contractsmodule.Module
	byName  map[string]contractsmodule.Module
	paths   map[string]string

	// mainModule is the app's Go module path; directory holds third-party modules.
	mainModule string
	directory  string
}

// NewRegistry places third-party modules under directory (module.path), "" for
// DefaultDirectory.
func NewRegistry(directory string) *Registry {
	if directory == "" {
		directory = DefaultDirectory
	}

	return &Registry{
		byName:     make(map[string]contractsmodule.Module),
		paths:      make(map[string]string),
		mainModule: env.MainPath(),
		directory:  directory,
	}
}

// Register stops at the first error; earlier modules stay registered.
func (r *Registry) Register(modules ...contractsmodule.Module) error {
	for _, module := range modules {
		if module == nil {
			continue
		}

		name := module.Name()
		if !ValidName(name) {
			return errors.ModularModuleNameNotKebab.Args(module, name)
		}
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
func (r *Registry) All() []contractsmodule.Module {
	sorted := slices.Clone(r.modules)
	slices.SortStableFunc(sorted, func(a, b contractsmodule.Module) int {
		return compareOrder(order(a), order(b))
	})

	return sorted
}

func (r *Registry) Find(name string) (contractsmodule.Module, bool) {
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

func (r *Registry) Capabilities(module contractsmodule.Module) map[string]bool {
	return map[string]bool{
		"services":     implements[foundation.ProvidesServices](module),
		"config":       implements[contractsmodule.ProvidesConfig](module),
		"middleware":   implements[contractsmodule.ProvidesGlobalMiddleware](module),
		"routes":       implements[contractsmodule.ProvidesRoutes](module),
		"migrations":   implements[contractsmodule.ProvidesMigrations](module),
		"seeders":      implements[contractsmodule.ProvidesSeeders](module),
		"commands":     implements[contractsmodule.ProvidesCommands](module),
		"jobs":         implements[contractsmodule.ProvidesJobs](module),
		"events":       implements[contractsmodule.ProvidesEvents](module),
		"schedule":     implements[contractsmodule.ProvidesSchedule](module),
		"views":        implements[contractsmodule.ProvidesViews](module),
		"translations": implements[contractsmodule.ProvidesTranslations](module),
		"vite-entries": implements[contractsmodule.ProvidesViteEntries](module),
		"vite-aliases": implements[contractsmodule.ProvidesViteAlias](module),
		"static":       implements[contractsmodule.ProvidesStatic](module),
		"sources":      implements[contractsmodule.ProvidesSources](module),
	}
}

// Providing returns the modules implementing T, as T.
func Providing[T any](modules []contractsmodule.Module) []T {
	var out []T
	for _, module := range modules {
		if p, ok := module.(T); ok {
			out = append(out, p)
		}
	}

	return out
}

func implements[T any](module contractsmodule.Module) bool {
	_, ok := module.(T)

	return ok
}

func order(module contractsmodule.Module) int {
	if o, ok := module.(contractsmodule.Ordered); ok {
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
