package orm

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/carbon"
)

const Associations = clause.Associations

var ErrRecordNotFound = errors.New("record not found")

var Observers = make([]Observer, 0)

type Observer struct {
	Model    any
	Observer contractsorm.Observer
}

type Model struct {
	ID uint `gorm:"primaryKey"`
	Timestamps
}

type SoftDeletes struct {
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

type Timestamps struct {
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at"`
}
