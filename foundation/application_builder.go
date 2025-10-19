package foundation

import "github.com/goravel/framework/contracts/foundation"

func Configure() foundation.ApplicationBuilder {
	return NewApplicationBuilder(App)
}

type ApplicationBuilder struct {
	app    foundation.Application
	config func()
}

func NewApplicationBuilder(app foundation.Application) *ApplicationBuilder {
	return &ApplicationBuilder{
		app: app,
	}
}

func (r *ApplicationBuilder) Create() foundation.Application {
	r.app.Boot()

	if r.config != nil {
		r.config()
	}

	return r.app
}

func (r *ApplicationBuilder) Run() {
	r.Create().Run()
}

func (r *ApplicationBuilder) WithConfig(fn func()) foundation.ApplicationBuilder {
	r.config = fn

	return r
}

func (r *ApplicationBuilder) WithProviders(providers []foundation.ServiceProvider) foundation.ApplicationBuilder {
	for _, provider := range providers {
		_ = provider
	}

	return r
}
