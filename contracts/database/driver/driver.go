package driver

import (
	"database/sql"

	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Driver interface {
	// Config returns the database configuration.
	// DEPRECATED
	Config() database.Config
	// DB returns the database connection.
	// DEPRECATED: 123
	DB() (*sql.DB, error)
	// Docker returns the database driver for Docker.
	Docker() (docker.DatabaseDriver, error)
	// Explain generates an SQL string with given parameters.
	// DEPRECATED
	Explain(sql string, args ...any) string
	// Gorm returns the Gorm database connection.
	// DEPRECATED
	Gorm() (*gorm.DB, error)
	// Grammar returns the database grammar.
	Grammar() Grammar
	// Pool returns the database pool.
	Pool() database.Pool
	// Processor returns the database processor.
	Processor() Processor
}
