package modular

import (
	"io/fs"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
)

// Module is an application feature in one Go package. Its capabilities are the
// optional interfaces below it implements; providers come from
// foundation.ProvidesServices.
type Module interface {
	// Name is unique and kebab-case, e.g. "blog": views "blog/...", translation
	// groups "blog/...", config "modules.blog", Vite alias "@blog".
	Name() string
}

// Located sets the module directory, relative to the project root. Without it,
// the directory is derived from the module's Go package (see Registry.Register).
type Located interface {
	Path() string
}

// Ordered sets the module order: lower first, negative last, ties by registration.
// The default is 0.
type Ordered interface {
	Order() int
}

// ProvidesConfig contributes the config stored under "modules.<name>".
type ProvidesConfig interface {
	Config() map[string]any
}

// ProvidesGlobalMiddleware contributes global HTTP middleware, added before WithMiddleware runs.
type ProvidesGlobalMiddleware interface {
	GlobalMiddleware() []http.Middleware
}

// ProvidesRoutes registers web routes in the modular.routes.web group.
type ProvidesRoutes interface {
	Routes(router route.Router)
}

// ProvidesApiRoutes registers API routes in the modular.routes.api group.
type ProvidesApiRoutes interface {
	ApiRoutes(router route.Router)
}

// ProvidesMigrations contributes database migrations.
type ProvidesMigrations interface {
	Migrations() []schema.Migration
}

// ProvidesSeeders contributes database seeders.
type ProvidesSeeders interface {
	Seeders() []seeder.Seeder
}

// ProvidesCommands contributes Artisan commands.
type ProvidesCommands interface {
	Commands() []console.Command
}

// ProvidesJobs contributes queue jobs.
type ProvidesJobs interface {
	Jobs() []queue.Job
}

// ProvidesEvents contributes listeners, merged per event with WithEvents.
type ProvidesEvents interface {
	Events() map[event.Event][]event.Listener
}

// ProvidesSchedule contributes scheduled events.
type ProvidesSchedule interface {
	Schedule() []schedule.Event
}

// ProvidesViews contributes views: "index.tmpl" becomes "<name>/index.tmpl".
type ProvidesViews interface {
	Views() fs.FS
}

// ProvidesTranslations contributes translations: "en/messages.json" is the group
// "<name>/messages". The "*" group (en.json) is not supported.
type ProvidesTranslations interface {
	Translations() fs.FS
}

// ProvidesViteEntries lists Vite entries relative to the module directory, built
// even when no view references them.
type ProvidesViteEntries interface {
	ViteEntries() []string
}

// ProvidesViteAlias contributes Vite aliases, e.g. "@blog-assets" -> "resources/assets"
// (relative to the module directory). Its own "@<name>" key replaces the default.
type ProvidesViteAlias interface {
	ViteAliases() map[string]string
}

// ProvidesSources ships files of a third-party module that modular:publish copies
// into its directory for Vite and Tailwind, e.g. "resources/js/app.js". Go files
// (*.go, go.mod, go.sum) are skipped.
type ProvidesSources interface {
	Sources() fs.FS
}

// ProvidesStatic serves files as-is: URL prefix -> Path.
type ProvidesStatic interface {
	Static() map[string]Path
}

// Path is a static directory: Dir inside FS, or on disk relative to the module
// directory when FS is nil. Dir must stay inside the module.
type Path struct {
	FS  fs.FS
	Dir string
}

// Registry holds the application's modules.
type Registry interface {
	// Register adds modules and resolves their directories: Located, else the
	// package path inside the app, else "<modular.path>/<name>" for another Go
	// module. Duplicate names and unresolvable directories are errors.
	Register(modules ...Module) error
	// All returns the modules sorted by Order.
	All() []Module
	// Find returns the module named name.
	Find(name string) (Module, bool)
	// Has reports whether name is registered.
	Has(name string) bool
	// Path returns a module's directory relative to the project root.
	Path(name string) (string, bool)
	// Capabilities reports which optional interfaces the module implements, by label.
	Capabilities(module Module) map[string]bool
}
