package config

import (
	"flag"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support"
)

type ServiceProvider struct {
}

func (config *ServiceProvider) Register() {
	var env *string
	if support.Env == support.EnvTest {
		testEnv := ".env"
		env = &testEnv
	} else {
		env = flag.String("env", ".env", "custom .env path")
		flag.Parse()
	}
	facades.Config = NewApplication(*env)
}

func (config *ServiceProvider) Boot() {

}
