package gorm

import (
	"context"

	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/database/gorm"
)

type Gorm interface {
	Make() (*gormio.DB, error)
}

type Initialize interface {
	InitializeGorm(config config.Config, connection string) *gorm.GormImpl
	InitializeQuery(ctx context.Context, config config.Config, connection string) (*gorm.QueryImpl, error)
}

type InitializeImpl struct{}

func NewInitializeImpl() *InitializeImpl {
	return &InitializeImpl{}
}

func (receive *InitializeImpl) InitializeGorm(config config.Config, connection string) *gorm.GormImpl {
	return gorm.InitializeGorm(config, connection)
}

func (receive *InitializeImpl) InitializeQuery(ctx context.Context, config config.Config, connection string) (*gorm.QueryImpl, error) {
	return gorm.InitializeQuery(ctx, config, connection)
}
