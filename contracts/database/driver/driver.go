package driver

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/testing"
)

type Driver interface {
	Config() database.Config
	Docker() (testing.DatabaseDriver, error)
	Gorm() (*gorm.DB, GormQuery, error)
	Grammar() schema.Grammar
	Processor() schema.Processor
	Schema(orm.Orm) schema.DriverSchema
}

type GormQuery interface {
	LockForUpdate() clause.Expression
	RandomOrder() string
	SharedLock() clause.Expression
}
