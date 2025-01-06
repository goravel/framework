package orm

import "gorm.io/gorm"

type Driver interface {
	Dialector() gorm.Dialector
}
