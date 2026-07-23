package broadcasting

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
