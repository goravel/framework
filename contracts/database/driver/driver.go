package driver

import (
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Driver interface {
	Config() database.Config
	DB() (*sql.DB, error)
	Docker() (docker.DatabaseDriver, error)
	// Explain generate SQL string with given parameters
	Explain(sql string, args ...any) string
	Gorm() (*gorm.DB, GormQuery, error)
	Grammar() Grammar
	Processor() Processor
}

// TODO: Remove this, use Compile instead
type GormQuery interface {
	LockForUpdate() clause.Expression
	RandomOrder() string
	SharedLock() clause.Expression
}
