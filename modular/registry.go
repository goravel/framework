// Package modular implements the module registry and module filesystems. It
// imports no framework implementation, so foundation can import it.
package modular

import (
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
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
	// sorted caches the dependency order, nil until needed.
	sorted []contractsmodule.Module

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

// Register adds modules and resolves their directories (Located, app package, or module.path).
// It stops at the first error; earlier modules stay registered.
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
		r.sorted = nil
	}

	return nil
}

// All returns the modules in registration order, each after its dependencies.
// It is plain registration order when the dependencies cannot be resolved.
func (r *Registry) All() []contractsmodule.Module {
	if r.sorted == nil {
		r.sorted, _ = r.sort()
	}

	return slices.Clone(r.sorted)
}

// Resolve reports a dependency that is not registered, or a cycle. Call it once
// every module is registered.
func (r *Registry) Resolve() error {
	sorted, err := r.sort()
	r.sorted = sorted

	return err
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
		"late-routes":  implements[contractsmodule.ProvidesLateRoutes](module),
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

// sort places each module after its dependencies, depth first in registration
// order. On an error it returns the registration order.
func (r *Registry) sort() ([]contractsmodule.Module, error) {
	const (
		visiting = iota + 1
		done
	)

	sorted := make([]contractsmodule.Module, 0, len(r.modules))
	state := make(map[string]int, len(r.modules))
	var stack []string

	var visit func(module contractsmodule.Module) error
	visit = func(module contractsmodule.Module) error {
		name := module.Name()
		if state[name] == done {
			return nil
		}

		stack = append(stack, name)
		if state[name] == visiting {
			cycle := stack[slices.Index(stack, name):]

			return errors.ModularDependencyCycle.Args(strings.Join(cycle, " -> "))
		}
		state[name] = visiting

		if dependent, ok := module.(contractsmodule.Dependent); ok {
			for _, dependency := range dependent.Dependencies() {
				required, ok := r.byName[dependency]
				if !ok {
					return errors.ModularDependencyMissing.Args(name, dependency)
				}
				if err := visit(required); err != nil {
					return err
				}
			}
		}

		stack = stack[:len(stack)-1]
		state[name] = done
		sorted = append(sorted, module)

		return nil
	}

	for _, module := range r.modules {
		if err := visit(module); err != nil {
			return slices.Clone(r.modules), err
		}
	}

	return sorted, nil
}
