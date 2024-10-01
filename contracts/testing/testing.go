package testing

import (
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/seeder"
)

type Testing interface {
	// Docker get the Docker instance.
	Docker() Docker
}

type Docker interface {
	// Database get a database connection instance.
	Database(connection ...string) (Database, error)
}

type Database interface {
	DatabaseDriver
	// Seed runs the database seeds.
	Seed(seeds ...seeder.Seeder)
}

type DatabaseDriver interface {
	// Build the database.
	Build() error
	// Config get database configuration.
	Config() DatabaseConfig
	// Fresh the database.
	Fresh() error
	// Image gets the database image.
	Image(image Image)
	// Driver gets the database driver name.
	Driver() database.Driver
	// Stop the database.
	Stop() error
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type Image struct {
	Env          []string
	ExposedPorts []string
	Repository   string
	Tag          string
}
