package orm

import (
	"time"

	"gorm.io/gorm"

	"github.com/goravel/framework/facades"
)

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

type BaseModel struct {
}

func (r *BaseModel) HasOne(dest, id interface{}, foreignKey string) error {
	return facades.Orm.Query().Where(foreignKey+" = ?", id).Find(dest)
}

func (r *BaseModel) HasMany(dest, id interface{}, foreignKey string) error {
	return facades.Orm.Query().Where(foreignKey+" in ?", id).Find(dest)
}

func (r *BaseModel) BelongsTo(dest, id interface{}) error {
	return facades.Orm.Query().Find(dest, id)
}
