package foundation

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation/configuration"
)

type ApplicationBuilder interface {
	// Create a new application instance after configuring.
	Create() Application
	// Run the application.
	Run()
	// WithCommands sets the application's commands.
	WithCommands(commands []console.Command) ApplicationBuilder
	// WithConfig sets a callback function to configure the application.
	WithConfig(config func()) ApplicationBuilder
	// WithEvents sets event listeners for the application.
	WithEvents(eventToListeners map[event.Event][]event.Listener) ApplicationBuilder
	// WithMiddleware registers the http's middleware.
	WithMiddleware(fn func(handler configuration.Middleware)) ApplicationBuilder
	// WithProviders registers and boots custom service providers.
	WithProviders(providers []ServiceProvider) ApplicationBuilder
	// WithRouting registers the application's routes.
	WithRouting(routes []func()) ApplicationBuilder
}
