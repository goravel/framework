package config

import (
	"flag"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support"
)

const Binding = "goravel.config"

type ServiceProvider struct {
}

func (config *ServiceProvider) Register(app foundation.Application) {
	var env *string
	if support.Env == support.EnvTest {
		testEnv := ".env"
		env = &testEnv
	} else {
		env = flag.String("env", ".env", "custom .env path")
		flag.Parse()
	}

	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewApplication(*env), nil
	})
}

func (config *ServiceProvider) Boot(app foundation.Application) {

}
