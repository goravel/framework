package orm

import (
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
)

type Driver interface {
	Config() database.Config1
	Docker() (testing.DatabaseDriver, error)
	Gorm() (*gorm.DB, error)
}
