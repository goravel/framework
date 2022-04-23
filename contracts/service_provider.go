package contracts

type ServiceProvider interface {
	//Boot Bootstrap any application services after register.
	Boot()
	//Register Register any application services.
	Register()
}
