package broadcasting

import (
	contractshttp "github.com/goravel/framework/contracts/http"
)

type ConnectionConfig struct {
	Driver  string
	Key     string
	Secret  string
	AppID   string `json:"app_id"`
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
	Middleware []contractshttp.Middleware
}
