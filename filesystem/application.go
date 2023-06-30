package filesystem

import (
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
)

type Driver string

const (
	DriverLocal  Driver = "local"
	DriverCustom Driver = "custom"
)

type Storage struct {
	filesystem.Driver
	config  config.Config
	drivers map[string]filesystem.Driver
}

func NewStorage(config config.Config) *Storage {
	defaultDisk := config.GetString("filesystems.default")
	if defaultDisk == "" {
		color.Redln("[filesystem] please set default disk")

		return nil
	}

	driver, err := NewDriver(config, defaultDisk)
	if err != nil {
		color.Redf("[filesystem] %s\n", err)

		return nil
	}

	drivers := make(map[string]filesystem.Driver)
	drivers[defaultDisk] = driver
	return &Storage{
		Driver:  driver,
		config:  config,
		drivers: drivers,
	}
}

func NewDriver(config config.Config, disk string) (filesystem.Driver, error) {
	driver := Driver(config.GetString(fmt.Sprintf("filesystems.disks.%s.driver", disk)))
	switch driver {
	case DriverLocal:
		return NewLocal(config, disk)
	case DriverCustom:
		driver, ok := config.Get(fmt.Sprintf("filesystems.disks.%s.via", disk)).(filesystem.Driver)
		if ok {
			return driver, nil
		}

		driverCallback, ok := config.Get(fmt.Sprintf("filesystems.disks.%s.via", disk)).(func() (filesystem.Driver, error))
		if ok {
			return driverCallback()
		}

		return nil, fmt.Errorf("init %s disk fail: via must be implement filesystem.Driver or func() (filesystem.Driver, error)", disk)
	}

	return nil, fmt.Errorf("invalid driver: %s, only support local, custom", driver)
}

func (r *Storage) Disk(disk string) filesystem.Driver {
	if driver, exist := r.drivers[disk]; exist {
		return driver
	}

	driver, err := NewDriver(r.config, disk)
	if err != nil {
		panic(err)
	}

	r.drivers[disk] = driver

	return driver
}
