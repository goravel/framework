package database

import (
	"context"

	"github.com/goravel/framework/contracts/config"
	contractsgorm "github.com/goravel/framework/contracts/database/gorm"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
)

type InitializeImpl struct{}

func NewInitializeImpl() *InitializeImpl {
	return &InitializeImpl{}
}

func (receive *InitializeImpl) InitializeGorm(config config.Config, connection string) contractsgorm.Gorm {
	return gorm.InitializeGorm(config, connection)
}

func (receive *InitializeImpl) InitializeQuery(ctx context.Context, config config.Config, connection string) (orm.Query, error) {
	return gorm.InitializeQuery(ctx, config, connection)
}
