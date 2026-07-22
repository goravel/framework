package broadcasting

type ConnectionConfig struct {
	Driver  string
	Key     string
	Secret  string
	AppID   string
	Options PusherOptions
}

type PusherOptions struct {
	Cluster string
	Host    string
	Port    int
	Scheme  string
}

type AuthConfig struct {
	Enabled bool
	Path    string
}
