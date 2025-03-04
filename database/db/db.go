package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/goravel/framework/contracts/config"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractslogger "github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/database/logger"
	"github.com/goravel/framework/errors"
)

type DB struct {
	config  config.Config
	ctx     context.Context
	db      *sqlx.DB
	driver  contractsdriver.Driver
	log     log.Log
	logger  contractslogger.Logger
	queries map[string]contractsdb.DB
	tx      *sqlx.Tx
	txLogs  *[]TxLog
}

func NewDB(ctx context.Context, config config.Config, driver contractsdriver.Driver, log log.Log, db *sqlx.DB, tx *sqlx.Tx, txLogs *[]TxLog) *DB {
	return &DB{ctx: ctx, config: config, driver: driver, log: log, logger: logger.NewLogger(config, log), db: db, queries: make(map[string]contractsdb.DB), tx: tx, txLogs: txLogs}
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

	return NewDB(ctx, config, driver, log, sqlx.NewDb(instance, driver.Config().Driver), nil, nil), nil
}

func (r *DB) BeginTransaction() (contractsdb.DB, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}

	return NewDB(r.ctx, r.config, r.driver, r.log, nil, tx, &[]TxLog{}), nil
}

func (r *DB) Commit() error {
	if r.tx == nil {
		return errors.DatabaseTransactionNotStarted
	}

	if err := r.tx.Commit(); err != nil {
		return err
	}

	for _, log := range *r.txLogs {
		r.logger.Trace(log.ctx, log.begin, log.sql, log.rowsAffected, log.err)
	}

	return nil
}

func (r *DB) Connection(name string) contractsdb.DB {
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

func (r *DB) Rollback() error {
	if r.tx == nil {
		return errors.DatabaseTransactionNotStarted
	}

	return r.tx.Rollback()
}

func (r *DB) Table(name string) contractsdb.Query {
	if r.tx != nil {
		return NewQuery(r.ctx, r.driver, r.tx, r.logger, name, r.txLogs)
	}

	return NewQuery(r.ctx, r.driver, r.db, r.logger, name, nil)
}

func (r *DB) Transaction(callback func(tx contractsdb.DB) error) error {
	tx, err := r.BeginTransaction()
	if err != nil {
		return err
	}

	err = callback(tx)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}

		return err
	}

	return tx.Commit()
}

func (r *DB) WithContext(ctx context.Context) contractsdb.DB {
	return NewDB(ctx, r.config, r.driver, r.log, r.db, r.tx, r.txLogs)
}
