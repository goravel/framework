package http

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/route"
)

type HTTPRunner struct {
	config config.Config
	route  route.Route
}

func NewHTTPRunner(config config.Config, route route.Route) *HTTPRunner {
	return &HTTPRunner{
		config: config,
		route:  route,
	}
}

func (r *HTTPRunner) Signature() string {
	return "goravel:http"
}

func (r *HTTPRunner) ShouldRun() bool {
	return r.route != nil && r.config.GetString("http.default") != ""
}

func (r *HTTPRunner) Run() error {
	tlsHost := r.config.GetString("http.tls.host")
	tlsPort := r.config.GetString("http.tls.port")
	certFile := r.config.GetString("http.tls.ssl.cert")
	keyFile := r.config.GetString("http.tls.ssl.key")

	tlsShouldRun := tlsHost != "" && tlsPort != "" && certFile != "" && keyFile != ""
	if tlsShouldRun {
		if err := r.route.RunTLS(); err != nil {
			return err
		}
	}

	host := r.config.GetString("http.host")
	port := r.config.GetString("http.port")

	if host != "" && port != "" && (!tlsShouldRun || port != tlsPort) {
		if err := r.route.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (r *HTTPRunner) Shutdown() error {
	return r.route.Shutdown()
}
