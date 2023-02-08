package cache

import (
	"github.com/goravel/framework/facades"
)

func prefix() string {
	return facades.Config.GetString("cache.prefix") + ":"
}
