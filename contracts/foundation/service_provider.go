package foundation

import "github.com/goravel/framework/contracts/binding"

type ServiceProvider interface {
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
}

type ServiceProviderWithRelations interface {
	// Relationship returns the service provider's relationship.
	Relationship() binding.Relationship
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
}

type ServiceProviderWithRunners interface {
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
	// Runners returns the service provider's runners.
	Runners(app Application) []Runner
}
