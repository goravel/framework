package cache

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/cache"
	"github.com/goravel/framework/contracts/config"
)

//go:generate mockery --name=Driver
type Driver interface {
	New(store string) cache.Driver
}

type DriverImpl struct {
	config config.Config
}

func NewDriverImpl(config config.Config) *DriverImpl {
	return &DriverImpl{
		config: config,
	}
}

func (d *DriverImpl) New(store string) cache.Driver {
	driver := d.config.GetString(fmt.Sprintf("cache.stores.%s.driver", store))
	switch driver {
	case "redis":
		return d.redis(store)
	case "memory":
		return d.memory()
	case "custom":
		return d.custom(store)
	default:
		color.Redf("[Cache] Not supported cache store: %s\n", store)
		return nil
	}
}

func (d *DriverImpl) redis(store string) cache.Driver {
	redis, err := NewRedis(context.Background(), d.config, store)
	if err != nil {
		color.Redf("[Cache] Init redis driver error: %v\n", err)
		return nil
	}
	if redis == nil {
		return nil
	}

	return redis
}

func (d *DriverImpl) memory() cache.Driver {
	memory, err := NewMemory(d.config)
	if err != nil {
		color.Redf("[Cache] Init memory driver error: %v\n", err)
		return nil
	}

	return memory
}

func (d *DriverImpl) custom(store string) cache.Driver {
	if custom, ok := d.config.Get(fmt.Sprintf("cache.stores.%s.via", store)).(cache.Driver); ok {
		return custom
	}
	color.Redf("[Cache] %s doesn't implement contracts/cache/store\n", store)

	return nil
}
