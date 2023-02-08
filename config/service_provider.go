package config

import (
	"flag"

	"github.com/goravel/framework/facades"
)

type ServiceProvider struct {
}

func (config *ServiceProvider) Register() {
	env := flag.String("env", ".env", "custom .env path")
	flag.Parse()
	facades.Config = NewApplication(*env)
}

func (config *ServiceProvider) Boot() {

}
