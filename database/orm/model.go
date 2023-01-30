package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const Associations = clause.Associations

type Model struct {
	ID uint `gorm:"primaryKey"`
	Timestamps
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
