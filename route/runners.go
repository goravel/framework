pakcage route

type RouteRunner struct {
	route route.Route
}

func NewRouteRunner(route route.Route) *RouteRunner {
	return &RouteRunner{
		route: route,
	}
}

func (r *RouteRunner) ShouldRun() bool {
	return "Route"
}

func (r *RouteRunner) Run() error {
	return r.route.Run()
}

func (r *RouteRunner) Shutdown() error {
	return r.route.Shutdown()
}
