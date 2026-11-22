package filesystem

import (
	"fmt"
	"sync"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

type Driver string

const (
	DriverLocal  Driver = "local"
	DriverCustom Driver = "custom"
)

type Storage struct {
	filesystem.Driver
	config      config.Config
	defaultDisk string
	drivers     map[string]filesystem.Driver
	mu          sync.RWMutex
}

func NewStorage(config config.Config) (*Storage, error) {
	defaultDisk := config.GetString("filesystems.default")
	if defaultDisk == "" {
		return nil, errors.FilesystemDefaultDiskNotSet.SetModule(errors.ModuleFilesystem)
	}

	driver, err := NewDriver(config, defaultDisk)
	if err != nil {
		return nil, err
	}

	drivers := make(map[string]filesystem.Driver)
	drivers[defaultDisk] = driver
	return &Storage{
		Driver:      driver,
		config:      config,
		defaultDisk: defaultDisk,
		drivers:     drivers,
	}, nil
}

func NewDriver(config config.Config, disk string) (filesystem.Driver, error) {
	driver := Driver(config.GetString(fmt.Sprintf("filesystems.disks.%s.driver", disk)))
	switch driver {
	case DriverLocal:
		return NewLocal(config, disk)
	case DriverCustom:
		via := config.Get(fmt.Sprintf("filesystems.disks.%s.via", disk))
		driver, ok := via.(filesystem.Driver)
		if ok {
			return driver, nil
		}

		driverCallback, ok := via.(func() (filesystem.Driver, error))
		if ok {
			return driverCallback()
		}

		return nil, errors.FilesystemInvalidCustomDriver.Args(disk)
	}

	return nil, errors.FilesystemDriverNotSupported.Args(driver)
}

func (r *Storage) Disk(disk string) filesystem.Driver {
	r.mu.RLock()
	driver, exist := r.drivers[disk]
	r.mu.RUnlock()
	if exist {
		return driver
	}

	driver, err := NewDriver(r.config, disk)
	if err != nil {
		panic(err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Another call may have loaded or faked the disk meanwhile.
	if existing, exist := r.drivers[disk]; exist {
		return existing
	}
	r.drivers[disk] = driver

	return driver
}
