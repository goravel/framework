package testing

import (
	"github.com/ory/dockertest/v3"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/seeder"
)

type Testing interface {
	Docker() Docker
}

type Docker interface {
	Database(connection ...string) (Database, error)
}

type Database interface {
	Build() error
	Config() Config
	Clear() error
	Image(Image)
	Seed(seeds ...seeder.Seeder)
}

type DatabaseDriver interface {
	Config(resource *dockertest.Resource) Config
	Clear(pool *dockertest.Pool, resource *dockertest.Resource) error
	Name() orm.Driver
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
