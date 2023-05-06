package foundation

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
)

type Container interface {
	Bind(key any, callback func() (any, error))
	BindWith(key any, callback func(parameters map[string]any) (any, error))
	Make(key any) (any, error)
	MakeConfig() config.Config
	MakeArtisan() console.Artisan
	MakeWith(key any, parameters map[string]any) (any, error)
	Singleton(key any, callback func() (any, error))
}
