package notification

type MapNotifiable struct {
	Routes map[string]any
}

func (n MapNotifiable) NotificationParams() map[string]interface{} {
	if n.Routes == nil {
		return nil
	}
	return n.Routes
}
