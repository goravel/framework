package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/support/carbon"
)

const Associations = clause.Associations

// Deprecated: use BaseModel instead
type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Timestamps
}

// Deprecated: use NullableSoftDeletes instead
type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// Deprecated: use NullableTimestamps instead
type Timestamps struct {
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type BaseModel struct {
	ID uint `gorm:"primaryKey" json:"id" db:"id"`
	NullableTimestamps
}

type NullableSoftDeletes struct {
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at" db:"deleted_at"`
}

type NullableTimestamps struct {
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at" db:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at" db:"updated_at"`
}
