package broadcasting

type Driver interface {
	Broadcast(channels []Channel, event string, payload map[string]any) error
}

type ConnectionConfig struct {
	Driver  string
	Key     string
	Secret  string
	AppID   string
	Options map[string]any
}

type PusherOptions struct {
	Cluster string
	Host    string
	Port    int
	Scheme  string
}

type AuthConfig struct {
	Enabled    bool
	Path       string
	Middleware []string
}
