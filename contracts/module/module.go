package module

import (
	"io/fs"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"
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

// Dependent names the modules that come before this one. Other modules keep
// the order they are registered in.
type Dependent interface {
	Dependencies() []string
}

// ProvidesConfig contributes the config stored under "modules.<name>".
type ProvidesConfig interface {
	Config() map[string]any
}

// ProvidesGlobalMiddleware contributes global HTTP middleware, added before WithMiddleware runs.
type ProvidesGlobalMiddleware interface {
	GlobalMiddleware() []http.Middleware
}

// ProvidesRoutes registers routes with the Route facade, after WithRouting runs.
type ProvidesRoutes interface {
	Routes()
}

// ProvidesLateRoutes registers routes after every module's Routes(), in reverse
// dependency order, so they follow the routes of the modules that depend on it.
type ProvidesLateRoutes interface {
	LateRoutes()
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
// The default directory is "resources/views".
type ProvidesViews interface {
	Views() Files
}

// ProvidesTranslations contributes translations: "en/messages.json" is the group
// "<name>/messages". The "*" group (en.json) is not supported. The default
// directory is "lang".
type ProvidesTranslations interface {
	Translations() Files
}

// ProvidesFrontend contributes the files Vite and Tailwind read.
type ProvidesFrontend interface {
	Frontend() Frontend
}

// Frontend is a module's Vite input. Entries and Aliases are relative to Files.
type Frontend struct {
	// Files is the module directory when zero. A third-party module sets Files.FS, which
	// module:publish copies into its directory; Go files (*.go, go.mod, go.sum) are skipped.
	Files Files
	// Entries are built even when no view references them, e.g. "resources/js/app.js".
	Entries []string
	// Aliases map e.g. "@blog-assets" to "resources/assets". "@<name>" replaces the default.
	Aliases map[string]string
}

// ProvidesStatic serves files as-is: URL prefix -> files. There is no default directory.
type ProvidesStatic interface {
	Static() map[string]Files
}

// Files is a directory of files: Dir inside FS, or on disk relative to the module
// directory when FS is nil. Dir must stay inside the module; the zero value is no files.
type Files struct {
	FS fs.FS
	// Dir "" is the capability's default directory inside FS, "." the root of FS.
	Dir string
}

// Registry holds the application's modules. The framework registers them at boot.
type Registry interface {
	// All returns the modules in registration order, each after its dependencies.
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
