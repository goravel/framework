package db

import (
	"context"
	databasesql "database/sql"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"

	"github.com/goravel/framework/contracts/config"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractslogger "github.com/goravel/framework/contracts/database/logger"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/carbon"
)

type DB struct {
	contractsdb.Tx
	config  config.Config
	ctx     context.Context
	db      contractsdb.Builder
	driver  contractsdriver.Driver
	logger  contractslogger.Logger
	queries map[string]contractsdb.DB
}

func NewDB(ctx context.Context, config config.Config, driver contractsdriver.Driver, logger contractslogger.Logger, db contractsdb.Builder) *DB {
	return &DB{
		Tx:      NewTx(ctx, driver, logger, db, nil, &[]TxLog{}),
		ctx:     ctx,
		config:  config,
		driver:  driver,
		logger:  logger,
		db:      db,
		queries: make(map[string]contractsdb.DB),
	}
}

func BuildDB(ctx context.Context, config config.Config, logger contractslogger.Logger, connection string) (*DB, error) {
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

	return NewDB(ctx, config, driver, logger, sqlx.NewDb(instance, driver.Config().Driver)), nil
}

func (r *DB) BeginTransaction() (contractsdb.Tx, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}

	return NewTx(r.ctx, r.driver, r.logger, nil, tx, &[]TxLog{}), nil
}

func (r *DB) Connection(name string) contractsdb.DB {
	if name == "" {
		name = r.config.GetString("database.default")
	}

	if _, ok := r.queries[name]; !ok {
		db, err := BuildDB(r.ctx, r.config, r.logger, name)
		if err != nil {
			r.logger.Panicf(r.ctx, err.Error())
			return nil
		}
		r.queries[name] = db
		db.queries = r.queries
	}

	return r.queries[name]
}

func (r *DB) Transaction(callback func(tx contractsdb.Tx) error) error {
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
	return NewDB(ctx, r.config, r.driver, r.logger, r.db)
}

type Tx struct {
	ctx    context.Context
	db     contractsdb.Builder
	driver contractsdriver.Driver
	logger contractslogger.Logger
	tx     contractsdb.TxBuilder
	txLogs *[]TxLog
}

func NewTx(ctx context.Context, driver contractsdriver.Driver, logger contractslogger.Logger, db contractsdb.Builder, tx contractsdb.TxBuilder, txLogs *[]TxLog) *Tx {
	return &Tx{
		ctx: ctx, driver: driver, logger: logger, db: db, tx: tx, txLogs: txLogs,
	}
}

func (r *Tx) Commit() error {
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

func (r *Tx) Delete(sql string, args ...any) (*contractsdb.Result, error) {
	return r.exec(sql, args...)
}

func (r *Tx) Insert(sql string, args ...any) (*contractsdb.Result, error) {
	return r.exec(sql, args...)
}

func (r *Tx) Rollback() error {
	if r.tx == nil {
		return errors.DatabaseTransactionNotStarted
	}

	return r.tx.Rollback()
}

func (r *Tx) Select(dest any, sql string, args ...any) error {
	var (
		realSql string
		err     error
	)

	realSql = r.driver.Explain(sql, args...)

	if r.tx != nil {
		err = r.tx.SelectContext(r.ctx, dest, realSql, args...)
	} else {
		err = r.db.SelectContext(r.ctx, dest, realSql, args...)
	}

	if err != nil {
		r.logger.Trace(r.ctx, carbon.Now(), realSql, -1, err)

		return err
	}

	destValue := reflect.Indirect(reflect.ValueOf(dest))
	rowsAffected := int64(-1)
	if destValue.Kind() == reflect.Slice {
		rowsAffected = int64(destValue.Len())
	}

	r.logger.Trace(r.ctx, carbon.Now(), realSql, rowsAffected, nil)

	return nil
}

func (r *Tx) Table(name string) contractsdb.Query {
	if r.tx != nil {
		return NewQuery(r.ctx, r.driver, r.tx, r.logger, name, r.txLogs)
	}

	return NewQuery(r.ctx, r.driver, r.db, r.logger, name, nil)
}

func (r *Tx) Update(sql string, args ...any) (*contractsdb.Result, error) {
	return r.exec(sql, args...)
}

func (r *Tx) exec(sql string, args ...any) (*contractsdb.Result, error) {
	var (
		result databasesql.Result
		err    error
	)

	realSql := r.driver.Explain(sql, args...)

	if r.tx != nil {
		result, err = r.tx.ExecContext(r.ctx, sql, args...)
	} else {
		result, err = r.db.ExecContext(r.ctx, sql, args...)
	}

	if err != nil {
		r.logger.Trace(r.ctx, carbon.Now(), realSql, -1, err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Trace(r.ctx, carbon.Now(), realSql, -1, err)
		return nil, err
	}

	r.logger.Trace(r.ctx, carbon.Now(), realSql, rowsAffected, nil)

	return &contractsdb.Result{RowsAffected: rowsAffected}, nil
}
