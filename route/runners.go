package route

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
)

type RouteRunner struct {
	config config.Config
	route  route.Route
}

func NewRouteRunner(config config.Config, route route.Route) *RouteRunner {
	return &RouteRunner{
		config: config,
		route:  route,
	}
}

func (r *RouteRunner) ShouldRun() bool {
	return r.route != nil && r.config.GetString("http.default") != ""
}

func (r *RouteRunner) Run() error {
	host := r.config.GetString("http.host")
	port := r.config.GetString("http.port")

	if host != "" && port != "" {
		if err := r.route.Run(); err != nil {
			return err
		}
	}

	tlsHost := r.config.GetString("http.tls.host")
	tlsPort := r.config.GetString("http.tls.port")

	if tlsHost != "" && tlsPort != "" && port != tlsPort {
		if err := r.route.RunTLS(); err != nil {
			return err
		}
	}

	return nil
}

func (r *RouteRunner) Shutdown() error {
	return r.route.Shutdown()
}
