package testing

import (
	"github.com/goravel/framework/contracts/database/orm"
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
	// Build the database.
	Build() error
	// Config gets the database configuration.
	Config() DatabaseConfig
	// DEPRECATED use Stop instead.
	Clear() error
	// Image gets the database image.
	Image(Image)
	// Seed runs the database seeds.
	Seed(seeds ...seeder.Seeder)
	// Stop stops the database.
	Stop() error
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
	// Name gets the database driver name.
	Name() orm.Driver
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
