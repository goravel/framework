package foundation

type ServiceProvider interface {
	//Register any application services.
	Register(app Application)
	//Boot any application services after register.
	Boot(app Application)
}
