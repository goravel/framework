package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const Relationships string = clause.Associations

type Model struct {
	ID uint `gorm:"primaryKey"`
	Timestamps
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"type:datetime"`
	UpdatedAt time.Time `gorm:"type:datetime"`
}
