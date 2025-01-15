package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/support/carbon"
)

const Associations = clause.Associations

type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Timestamps
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

type Timestamps struct {
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}
