package broadcasting

import (
	"github.com/goravel/framework/contracts/broadcasting"
	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/errors"
)

type Config struct {
	Default     string                              `json:"default"`
	Connections map[string]broadcasting.ConnectionConfig `json:"connections"`
	Auth        broadcasting.AuthConfig             `json:"auth"`
}

func NewConfig(cfg contractsconfig.Config) (*Config, error) {
	var c Config
	if err := cfg.UnmarshalKey("broadcasting", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) DefaultConnection() string {
	if c.Default == "" {
		return "log"
	}
	return c.Default
}

func (c *Config) Connection(name string) (broadcasting.ConnectionConfig, error) {
	conn, ok := c.Connections[name]
	if !ok {
		return broadcasting.ConnectionConfig{}, errors.BroadcastConnectionNotFound.Args(name)
	}
	return conn, nil
}
