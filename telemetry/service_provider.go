package telemetry

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/telemetry"
	"github.com/goravel/framework/errors"
)

var (
	Facade       telemetry.Telemetry
	ConfigFacade config.Config
)

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			binding.Telemetry,
		},
		Dependencies: binding.Bindings[binding.Telemetry].Dependencies,
		ProvideFor:   []string{},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(binding.Telemetry, func(app foundation.Application) (any, error) {
		cfg := app.MakeConfig()
		if cfg == nil {
			return nil, errors.ConfigFacadeNotSet.SetModule(errors.ModuleTelemetry)
		}

		var telemetryCfg Config
		if err := cfg.UnmarshalKey("telemetry", &telemetryCfg); err != nil {
			return nil, err
		}

		return NewApplication(telemetryCfg)
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	Facade = app.MakeTelemetry()
	ConfigFacade = app.MakeConfig()
}
