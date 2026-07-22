package broadcasting

import (
	"fmt"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
)

type Config struct {
	app config.Config
}

func NewConfig(app config.Config) *Config {
	return &Config{app: app}
}

func (c *Config) DefaultConnection() string {
	return c.app.GetString("broadcasting.default", "log")
}

func (c *Config) Connection(name string) (broadcasting.ConnectionConfig, error) {
	conn := c.app.Get(fmt.Sprintf("broadcasting.connections.%s", name))
	if conn == nil {
		return broadcasting.ConnectionConfig{}, errors.BroadcastConnectionNotFound.Args(name)
	}
	raw, ok := conn.(map[string]any)
	if !ok {
		return broadcasting.ConnectionConfig{}, errors.BroadcastConnectionInvalidFormat.Args(name)
	}

	options, _ := raw["options"].(map[string]any)

	return broadcasting.ConnectionConfig{
		Driver: getString(raw, "driver"),
		Key:    getString(raw, "key"),
		Secret: getString(raw, "secret"),
		AppID:  getString(raw, "app_id"),
		Options: broadcasting.PusherOptions{
			Cluster: getString(options, "cluster"),
			Host:    getString(options, "host"),
			Port:    getInt(options, "port"),
			Scheme:  getString(options, "scheme"),
		},
	}, nil
}

func (c *Config) Auth() broadcasting.AuthConfig {
	return broadcasting.AuthConfig{
		Enabled: c.app.GetBool("broadcasting.auth.enabled", true),
		Path:    c.app.GetString("broadcasting.auth.path", "/broadcasting/auth"),
	}
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}
