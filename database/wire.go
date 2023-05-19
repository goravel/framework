//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package database

import (
	"context"

	"github.com/google/wire"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
)

//go:generate wire
func InitializeOrm(ctx context.Context, config config.Config, connection string) (*OrmImpl, error) {
	wire.Build(NewOrmImpl, gorm.QuerySet, gorm.GormSet, db.ConfigSet, gorm.DialectorSet)

	return nil, nil
}
