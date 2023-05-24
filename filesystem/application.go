package filesystem

import (
	"context"
	"fmt"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/filesystem"
)

type Driver string

const (
	DriverLocal  Driver = "local"
	DriverS3     Driver = "s3"
	DriverOss    Driver = "oss"
	DriverCos    Driver = "cos"
	DriverMinio  Driver = "minio"
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
	ctx := context.Background()
	driver := Driver(config.GetString(fmt.Sprintf("filesystems.disks.%s.driver", disk)))
	switch driver {
	case DriverLocal:
		return NewLocal(config, disk)
	case DriverOss:
		return NewOss(ctx, config, disk)
	case DriverCos:
		return NewCos(ctx, config, disk)
	case DriverS3:
		return NewS3(ctx, config, disk)
	case DriverMinio:
		return NewMinio(ctx, config, disk)
	case DriverCustom:
		driver, ok := config.Get(fmt.Sprintf("filesystems.disks.%s.via", disk)).(filesystem.Driver)
		if !ok {
			return nil, fmt.Errorf("init %s disk fail: via must be implement filesystem.Driver", disk)
		}

		return driver, nil
	}

	return nil, fmt.Errorf("invalid driver: %s, only support local, s3, oss, cos, minio, custom", driver)
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
