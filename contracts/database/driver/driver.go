package driver

import (
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Driver interface {
	// Config returns the database configuration.
	Config() database.Config
	// DB returns the database connection.
	DB() (*sql.DB, error)
	// Docker returns the database driver for Docker.
	Docker() (docker.DatabaseDriver, error)
	// Explain generates an SQL string with given parameters.
	Explain(sql string, args ...any) string
	// Gorm returns the Gorm database connection.
	Gorm() (*gorm.DB, GormQuery, error)
	// Grammar returns the database grammar.
	Grammar() Grammar
	// Processor returns the database processor.
	Processor() Processor
}

// TODO: Remove this, use Compile instead
type GormQuery interface {
	LockForUpdate() clause.Expression
	RandomOrder() string
	SharedLock() clause.Expression
}
