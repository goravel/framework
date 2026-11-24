package foundation

// ProvidesServiceProviders is a module capability, declared here to keep contracts/module
// free of this package.
type ProvidesServiceProviders interface {
	ServiceProviders() []ServiceProvider
}
