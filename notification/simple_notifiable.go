package notification

type MapNotifiable struct {
    Routes map[string]any
}

func (n MapNotifiable) RouteNotificationFor(channel string) any {
    if n.Routes == nil {
        return nil
    }
    return n.Routes[channel]
}