package queue

import (
	"github.com/goravel/framework/contracts/queue"
)

const DriverSync string = "sync"
const DriverASync string = "async"
const DriverRedis string = "redis"
const DriverDatabase string = "database"

func NewDriver(connection string, config *Config) queue.Driver {
	switch config.Driver(config.DefaultConnection()) {
	case DriverSync:
		return NewSync(connection)
	case DriverASync:
		return NewASync(connection)
	case DriverRedis:
		return NewRedis(connection, config.Redis(connection))
	case DriverDatabase:
		return NewDatabase(connection, config.Database(connection))
	default:
		panic("unknown queue driver")
	}
}
