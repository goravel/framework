package gorm

import (
	"context"

	"github.com/goravel/framework/contracts/config"
)

//go:generate mockery --name=Initialize
type Initialize interface {
	InitializeGorm(config config.Config, connection string) *GormImpl
	InitializeQuery(ctx context.Context, config config.Config, connection string) (*QueryImpl, error)
}

type InitializeImpl struct{}

func NewInitializeImpl() *InitializeImpl {
	return &InitializeImpl{}
}

func (receive *InitializeImpl) InitializeGorm(config config.Config, connection string) *GormImpl {
	return InitializeGorm(config, connection)
}

func (receive *InitializeImpl) InitializeQuery(ctx context.Context, config config.Config, connection string) (*QueryImpl, error) {
	return InitializeQuery(ctx, config, connection)
}
