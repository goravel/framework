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
		// TODO
		return nil
	case DriverASync:
		// TODO
		return nil
	case DriverRedis:
		// TODO
		return nil
	case DriverDatabase:
		return driver.NewDatabase(connection, config.Database(connection))
	default:
		return nil
	}
}
