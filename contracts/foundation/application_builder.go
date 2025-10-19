package foundation

type ApplicationBuilder interface {
	// Create a new application instance after configuring.
	Create() Application
	// Run the application.
	Run()
	// WithConfig sets a callback function to configure the application.
	WithConfig(func()) ApplicationBuilder
	// WithProviders registers and boots custom service providers.
	WithProviders(providers []ServiceProvider) ApplicationBuilder
}
