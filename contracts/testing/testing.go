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
	Seed(seeders ...seeder.Seeder) error
}

type DatabaseDriver interface {
	// Build a database container, it doesn't wait for the database to be ready, the Ready method needs to be called if
	// you want to check the container status.
	Build() error
	// Config get database configuration.
	Config() DatabaseConfig
	// Database returns a new instance with a new database, the Build method needs to be called first.
	Database(name string) (DatabaseDriver, error)
	// Driver gets the database driver name.
	Driver() database.Driver
	// Fresh the database.
	Fresh() error
	// Image gets the database image.
	Image(image Image)
	// Ready checks if the database is ready, the Build method needs to be called first.
	Ready() error
	// Stop the database.
	Stop() error
}

type DatabaseConfig struct {
	Host        string
	Port        int
	Database    string
	Username    string
	Password    string
	ContainerID string
}

type Image struct {
	Env          []string
	ExposedPorts []string
	Repository   string
	Tag          string
}
