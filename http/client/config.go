package client

import "github.com/goravel/framework/contracts/http/client"

type FactoryConfig struct {
	Default string                   `mapstructure:"default_client"`
	Clients map[string]client.Config `mapstructure:"clients"`
}
