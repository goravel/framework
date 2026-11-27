package foundation

import "github.com/goravel/framework/contracts/binding"

type ServiceProvider interface {
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
}

type ServiceProviderWithRelations interface {
	ServiceProvider

	// Relationship returns the service provider's relationship.
	Relationship() binding.Relationship
}

// ServiceProviderWithFlushState is an optional interface for providers that own
// package-level state, such as caches or variables that hold facades.
//
// FlushState resets the package-level state the provider owns. Refresh calls it
// only for providers whose Boot completed, so a provider that registered but did
// not boot is not flushed. Restart reaches it through Refresh. Fresh never calls it.
//
// Refresh calls FlushState after it drops the container instances and before it
// registers the providers again. FlushState may resolve the configuration facade,
// which Fresh keeps. It must not resolve any other facade, because the previous
// bindings are still registered and a resolution would cache an instance built
// from them.
//
// FlushState must not depend on the order in which providers are flushed. The
// framework calls the hooks in the reverse of the boot order, but that order is
// an internal detail that may change. FlushState must be safe when the state is
// already empty. A panic in FlushState is not recovered and aborts the refresh,
// as a panic in Register or Boot does.
//
// Refresh must run while no goroutine uses the state that providers flush,
// because flushing can close resources such as connections. This includes
// runners, request handlers, and goroutines that the application or the tests
// start. A caller that keeps a facade or client resolved before Refresh must not
// use it after Refresh.
type ServiceProviderWithFlushState interface {
	ServiceProvider

	// FlushState resets the package-level state the provider owns.
	FlushState(app Application)
}

type ServiceProviderWithRunners interface {
	ServiceProvider

	// Runners returns the service provider's runners.
	Runners(app Application) []Runner
}
