package notification

import contractsnotification "github.com/goravel/framework/contracts/notification"

// Route begins an on-demand notification targeting a raw address, with no
// backing Notifiable model.
func (m *Manager) Route(channel, route string) contractsnotification.OnDemandNotifiable {
	return &onDemandNotifiable{
		manager: m,
		routes:  map[string]string{channel: route},
	}
}

// onDemandNotifiable is a Notifiable built on the fly by Manager.Route.
type onDemandNotifiable struct {
	manager *Manager
	routes  map[string]string
}

func (o *onDemandNotifiable) RouteNotificationFor(channel string) string {
	return o.routes[channel]
}

func (o *onDemandNotifiable) Route(channel, route string) contractsnotification.OnDemandNotifiable {
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
