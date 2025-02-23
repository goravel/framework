package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/logger"
	"github.com/goravel/framework/errors"
)

type DB struct {
	builder db.Builder
	config  config.Config
	ctx     context.Context
	driver  contractsdriver.Driver
	log     log.Log
	queries map[string]db.DB
}

func NewDB(ctx context.Context, config config.Config, driver contractsdriver.Driver, log log.Log, builder db.Builder) *DB {
	return &DB{ctx: ctx, config: config, driver: driver, log: log, builder: builder, queries: make(map[string]db.DB)}
}

func BuildDB(ctx context.Context, config config.Config, log log.Log, connection string) (*DB, error) {
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

	return NewDB(ctx, config, driver, log, sqlx.NewDb(instance, driver.Config().Driver)), nil
}

func (r *DB) Connection(name string) db.DB {
	if name == "" {
		name = r.config.GetString("database.default")
	}

	if _, ok := r.queries[name]; !ok {
		db, err := BuildDB(r.ctx, r.config, r.log, name)
		if err != nil {
			r.log.Panic(err.Error())
			return nil
		}
		r.queries[name] = db
		db.queries = r.queries
	}

	return r.queries[name]
}

func (r *DB) Table(name string) db.Query {
	return NewQuery(r.ctx, r.driver, r.builder, logger.NewLogger(r.config, r.log), name)
}

func (r *DB) WithContext(ctx context.Context) db.DB {
	return NewDB(ctx, r.config, r.driver, r.log, r.builder)
}
