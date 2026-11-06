package foundation

// ProvidesServices is a module capability, declared here to keep contracts/modular
// free of this package.
type ProvidesServices interface {
	ServiceProviders() []ServiceProvider
}
