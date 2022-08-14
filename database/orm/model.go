package orm

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" sql:"index"`
}

type DateTimes struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
