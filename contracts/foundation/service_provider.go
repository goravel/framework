package foundation

type ServiceProvider interface {
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
}

// UnimplementedServiceProvider is a default implementation of the Provider interface.
type UnimplementedServiceProvider struct{}

func (u *UnimplementedServiceProvider) Register(app Application) {}

func (u *UnimplementedServiceProvider) Boot(app Application) {}
