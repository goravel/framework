package notification

type Notification interface {
	Send(notifiable Notifiable) error
}

type Notif interface {
	// Via Return to the list of channel names
	Via(notifiable Notifiable) []string
}

type Channel interface {
	// Send sends the given notification to the given notifiable.
	Send(notifiable Notifiable, notif interface{}) error
}

type Notifiable interface {
    // RouteNotificationFor returns the route notification for the given channel.
    RouteNotificationFor(channel string) any
}

type PayloadProvider interface {
    // PayloadFor returns prepared payload data for specific channel.
    PayloadFor(channel string, notifiable Notifiable) (map[string]interface{}, error)
}
