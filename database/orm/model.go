package orm

import (
	"time"

	"gorm.io/gorm"

	"github.com/goravel/framework/facades"
)

type Model struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

type DateTimes struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Relationship struct {
}

func (r *Relationship) HasOne(dest, id interface{}, foreignKey string) error {
	return facades.Orm.Query().Where(foreignKey+" = ?", id).Find(dest)
}

func (r *Relationship) HasMany(dest, id interface{}, foreignKey string) error {
	return facades.Orm.Query().Where(foreignKey+" in ?", id).Find(dest)
}

func (r *Relationship) belongsTo(dest, id interface{}) error {
	return facades.Orm.Query().Find(dest, id)
}
