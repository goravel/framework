package testing

import (
	"github.com/goravel/framework/contracts/testing/docker"
)

type Testing interface {
	// Docker get the Docker instance.
	Docker() Docker
}

type Docker interface {
	// Database get a database connection instance.
	Database(connection ...string) (docker.Database, error)
}
