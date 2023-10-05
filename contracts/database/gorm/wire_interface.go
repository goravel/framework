package gorm

import (
	"context"

	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
)

type Gorm interface {
	Make() (*gormio.DB, error)
}

type Initialize interface {
	InitializeGorm(config config.Config, connection string) Gorm
	InitializeQuery(ctx context.Context, config config.Config, connection string) (orm.Query, error)
}
