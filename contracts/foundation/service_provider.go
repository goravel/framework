package foundation

type ServiceProvider interface {
	// Register any application services.
	Register(app Application)
	// Boot any application services after register.
	Boot(app Application)
}

// BaseServiceProvider is a default implementation of the Provider interface.
type BaseServiceProvider struct{}

func (s *BaseServiceProvider) Register(Application) {}

func (s *BaseServiceProvider) Boot(Application) {}
