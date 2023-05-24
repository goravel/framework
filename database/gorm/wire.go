//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package gorm

import (
	"context"

	"github.com/google/wire"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/database/db"
)

//go:generate wire
func InitializeGorm(config config.Config, connection string) *GormImpl {
	wire.Build(NewGormImpl, db.ConfigSet, DialectorSet)

	return nil
}

//go:generate wire
func InitializeQuery(ctx context.Context, config config.Config, connection string) (*QueryImpl, error) {
	wire.Build(NewQueryImpl, GormSet, db.ConfigSet, DialectorSet)

	return nil, nil
}
