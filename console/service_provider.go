package console

type ServiceProvider struct {
}

//Boot Bootstrap any application services after register.
func (console *ServiceProvider) Boot() {
	app := &Application{}
	app.Init()
}

//Register Register any application services.
func (console *ServiceProvider) Register() {
}
