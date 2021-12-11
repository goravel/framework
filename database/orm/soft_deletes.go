package orm

import (
	"gorm.io/gorm"
)

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" sql:"index"`
}
