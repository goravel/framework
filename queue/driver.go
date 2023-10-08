package queue

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/driver"
)

const DriverSync string = "sync"
const DriverASync string = "async"
const DriverRedis string = "redis"
const DriverDatabase string = "database"

func NewDriver(connection string, config *Config) queue.Driver {
	switch config.Driver(config.DefaultConnection()) {
	case DriverSync:
		return driver.NewSync(connection)
	case DriverASync:
		return driver.NewASync(connection)
	case DriverRedis:
		return driver.NewRedis(connection, config.Redis(connection))
	case DriverDatabase:
		return driver.NewDatabase(connection, config.Database(connection))
	default:
		return nil
	}
}
