package facades

import (
	configcontract "github.com/goravel/framework/contracts/config"
)

var Config = NewConfig()

func NewConfig() configcontract.Config {
	return App().MakeConfig()
}
