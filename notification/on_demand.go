package notification

import contractsnotification "github.com/goravel/framework/contracts/notification"

// onDemandNotifiable is a Notifiable built on the fly by Manager.Route.
type onDemandNotifiable struct {
	manager *Manager
	routes  map[string]any
}

func (o *onDemandNotifiable) RouteNotificationFor(channel string) any {
	return o.routes[channel]
}

func (o *onDemandNotifiable) Route(channel string, route any) contractsnotification.OnDemandNotifiable {
	o.routes[channel] = route
	return o
}

func (o *onDemandNotifiable) Notify(n contractsnotification.Notification) error {
	return o.manager.Send(o, n)
}

func (o *onDemandNotifiable) NotifyNow(n contractsnotification.Notification) error {
	return o.manager.SendNow(o, n)
}

var _ contractsnotification.OnDemandNotifiable = (*onDemandNotifiable)(nil)
