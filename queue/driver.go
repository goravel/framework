package queue

import (
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
)

func NewDriver(connection string, config contractsqueue.Config, db contractsdb.DB, jobStorer contractsqueue.JobStorer, json contractsfoundation.Json) (contractsqueue.Driver, error) {
	driver := config.Driver(connection)

	switch driver {
	case contractsqueue.DriverSync:
		return NewSync(), nil
	case contractsqueue.DriverDatabase:
		return NewDatabase(config, db, jobStorer, json, connection)
	case contractsqueue.DriverCustom:
		custom := config.Via(connection)
		if driver, ok := custom.(contractsqueue.Driver); ok {
			return driver, nil
		}
		if driver, ok := custom.(func() (contractsqueue.Driver, error)); ok {
			return driver()
		}
		return nil, errors.QueueDriverInvalid.Args(connection)
	default:
		return nil, errors.QueueDriverNotSupported.Args(driver)
	}
}
