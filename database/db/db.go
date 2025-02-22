package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractslogger "github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/logger"
	"github.com/goravel/framework/errors"
)

type DB struct {
	builder db.Builder
	ctx     context.Context
	driver  contractsdriver.Driver
	logger  contractslogger.Logger
}

func NewDB(ctx context.Context, driver contractsdriver.Driver, logger contractslogger.Logger, builder db.Builder) db.DB {
	return &DB{ctx: ctx, driver: driver, logger: logger, builder: builder}
}

func BuildDB(config config.Config, log log.Log, connection string) (db.DB, error) {
	driverCallback, exist := config.Get(fmt.Sprintf("database.connections.%s.via", connection)).(func() (contractsdriver.Driver, error))
	if !exist {
		return nil, errors.DatabaseConfigNotFound
	}

	driver, err := driverCallback()
	if err != nil {
		return nil, err
	}

	instance, err := driver.DB()
	if err != nil {
		return nil, err
	}

	return NewDB(context.Background(), driver, logger.NewLogger(config, log), sqlx.NewDb(instance, driver.Config().Driver)), nil
}

func (r *DB) Table(name string) db.Query {
	return NewQuery(r.ctx, r.driver, r.builder, r.logger, name)
}

func (r *DB) WithContext(ctx context.Context) db.DB {
	return NewDB(ctx, r.driver, r.logger, r.builder)
}
