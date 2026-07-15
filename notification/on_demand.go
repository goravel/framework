package notification

import contractsnotification "github.com/goravel/framework/contracts/notification"

// onDemandNotifiable is a Notifiable built on the fly by Manager.Route.
// routes is map[string]any (not map[string]string) to match Route's
// contract — RouteNotificationFor still returns string, type-asserting
// on read, since the string-based Notifiable interface can't change
// without breaking every existing channel.
type onDemandNotifiable struct {
	manager *Manager
	routes  map[string]any
}

// RouteNotificationFor satisfies contracts/notification.Notifiable.
// Non-string routes (stored via Route(channel, someStruct)) return ""
// here — they're only reachable by a custom channel that knows to look
// for them some other way; no built-in channel needs anything but a
// string today.
func (o *onDemandNotifiable) RouteNotificationFor(channel string) string {
	if s, ok := o.routes[channel].(string); ok {
		return s
	}
	return ""
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
