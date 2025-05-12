package queue

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

func NewDriver(connection string, config queue.Config) (queue.Driver, error) {
	driver := config.Driver(connection)

	switch driver {
	case queue.DriverSync:
		return NewSync(connection), nil
	case queue.DriverCustom:
		custom := config.Via(connection)
		if driver, ok := custom.(queue.Driver); ok {
			return driver, nil
		}
		if driver, ok := custom.(func() (queue.Driver, error)); ok {
			return driver()
		}
		return nil, errors.QueueDriverInvalid.Args(connection)
	default:
		return nil, errors.QueueDriverNotSupported.Args(driver)
	}
}
