package testing

import (
	"github.com/ory/dockertest/v3"

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
	Config() Config
	// Clear clears the database.
	Clear() error
	// Image gets the database image.
	Image(Image)
	// Seed runs the database seeds.
	Seed(seeds ...seeder.Seeder)
}

type DatabaseDriver interface {
	// Config gets the database configuration.
	Config(resource *dockertest.Resource) Config
	// Clear clears the database.
	Clear(pool *dockertest.Pool, resource *dockertest.Resource) error
	// Name gets the database driver name.
	Name() orm.Driver
	// Image gets the database image.
	Image() *dockertest.RunOptions
}

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type Image struct {
	Env        []string
	Repository string
	Tag        string
	// unit: second
	Timeout uint
}
