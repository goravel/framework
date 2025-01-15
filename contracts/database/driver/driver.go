package driver

import (
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/testing"
)

type Driver interface {
	Config() database.Config1
	Docker() (testing.DatabaseDriver, error)
	Gorm() (*gorm.DB, error)
	Grammar() schema.Grammar
	Processor() schema.Processor
	Schema() schema.DriverSchema
}
