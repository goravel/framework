package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/support/carbon"
)

const Associations = clause.Associations

// Model is the base model for all models in the application.
// @Deprecated use BaseModel instead.
type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Timestamps
}

// SoftDeletes is used to add soft delete functionality to a model.
// @Deprecated use NullableSoftDeletes instead.
type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// Timestamps is used to add created_at and updated_at timestamps to a model.
// @Deprecated use NullableTimestamps instead.
type Timestamps struct {
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type BaseModel struct {
	ID uint `gorm:"primaryKey" json:"id"`
	NullableTimestamps
}

type NullableSoftDeletes struct {
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

type NullableTimestamps struct {
	CreatedAt *carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt *carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}
