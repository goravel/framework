package foundation

type ProviderRepository interface {
	// Boot boots a list of registered providers, passing the app context.
	Boot(app Application, providers []ServiceProvider)

	// GetBooted returns a slice of all providers that have been booted.
	GetBooted() []ServiceProvider

	// LoadConfigured lazy-loads providers from the config file, using the app
	// context to access the Config facade.
	LoadConfigured(app Application) []ServiceProvider

	// Register sorts and registers a list of providers, passing the app context.
	Register(app Application, providers []ServiceProvider) []ServiceProvider

	// ResetConfiguredCache clears the in-memory cache of configured providers.
	ResetConfiguredCache()

	// SetConfigured manually sets the list of configured providers, bypassing config.
	SetConfigured(providers []ServiceProvider)
}
