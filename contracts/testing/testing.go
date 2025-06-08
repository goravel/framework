package testing

import (
	"github.com/goravel/framework/contracts/testing/docker"
)

type Testing interface {
	// Docker get the Docker instance.
	Docker() Docker
}

type Docker interface {
	// Cache gets a cache connection instance.
	Cache(connection string) (docker.CacheDriver, error)
	// Database gets a database connection instance.
	Database(connection ...string) (docker.Database, error)
}
