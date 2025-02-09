package driver

import (
	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Driver interface {
	Config() database.Config
	DB() (*sqlx.DB, error)
	Docker() (docker.DatabaseDriver, error)
	Gorm() (*gorm.DB, GormQuery, error)
	Grammar() schema.Grammar
	Processor() schema.Processor
}

type GormQuery interface {
	LockForUpdate() clause.Expression
	RandomOrder() string
	SharedLock() clause.Expression
}
