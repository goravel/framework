package gorm

import (
	"context"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/database/gorm"
)

//go:generate mockery --name=Initialize
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
