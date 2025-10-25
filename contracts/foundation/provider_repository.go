package foundation

type ProviderRepository interface {
	// Add appends new providers to the repository.
	// It skips any providers that have already been added.
	Add(providers []ServiceProvider)

	// Boot boots all registered service providers in dependency order.
	Boot(app Application)

	// GetBooted returns a slice of all providers that have been booted.
	GetBooted() []ServiceProvider

	// LoadFromConfig lazy-loads providers from the "app.providers" config.
	LoadFromConfig(app Application) []ServiceProvider

	// Register sorts and registers all configured providers in dependency order.
	Register(app Application) []ServiceProvider

	// Reset clears all configured providers and cached state.
	Reset()
}
