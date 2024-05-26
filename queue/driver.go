package queue

import (
	"github.com/goravel/framework/contracts/queue"
)

const DriverSync string = "sync"
const DriverASync string = "async"
const DriverCustom string = "custom"

func NewDriver(connection string, config *Config) queue.Driver {
	switch config.Driver(config.DefaultConnection()) {
	case DriverSync:
		return NewSync(connection)
	case DriverASync:
		return NewASync(connection)
	case DriverCustom:
		return NewCustom(connection)
	default:
		panic("unknown queue driver")
	}
}
