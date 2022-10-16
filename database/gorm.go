package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/support"
	"github.com/goravel/framework/facades"
)

type GormDB struct {
	contractsorm.Query
	instance *gorm.DB
}

func NewGormDB(ctx context.Context, connection string) (contractsorm.DB, error) {
	db, err := NewGormInstance(connection)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return &GormDB{
		Query:    NewGormQuery(db),
		instance: db,
	}, nil
}

func NewGormInstance(connection string) (*gorm.DB, error) {
	gormConfig, err := getGormConfig(connection)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init gorm config error: %v", err))
	}
	if gormConfig == nil {
		return nil, nil
	}

	var logLevel gormLogger.LogLevel
	if facades.Config.GetBool("app.debug") {
		logLevel = gormLogger.Info
	} else {
		logLevel = gormLogger.Error
	}

	logger := New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})

	return gorm.Open(gormConfig, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.LogMode(logLevel),
	})
}

func (r *GormDB) Begin() (contractsorm.Transaction, error) {
	tx := r.instance.Begin()

	return NewGormTransaction(tx), tx.Error
}

type GormTransaction struct {
	contractsorm.Query
	instance *gorm.DB
}

func NewGormTransaction(instance *gorm.DB) contractsorm.Transaction {
	return &GormTransaction{Query: NewGormQuery(instance), instance: instance}
}

func (r *GormTransaction) Commit() error {
	return r.instance.Commit().Error
}

func (r *GormTransaction) Rollback() error {
	return r.instance.Rollback().Error
}

type GormQuery struct {
	instance *gorm.DB
}

func NewGormQuery(instance *gorm.DB) contractsorm.Query {
	return &GormQuery{instance}
}

func (r *GormQuery) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *GormQuery) Create(value interface{}) error {
	return r.instance.Create(value).Error
}

func (r *GormQuery) Delete(value interface{}, conds ...interface{}) error {
	return r.instance.Delete(value, conds...).Error
}

func (r *GormQuery) Distinct(args ...interface{}) contractsorm.Query {
	tx := r.instance.Distinct(args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Exec(sql string, values ...interface{}) error {
	return r.instance.Exec(sql, values...).Error
}

func (r *GormQuery) Find(dest interface{}, conds ...interface{}) error {
	return r.instance.Find(dest, conds...).Error
}

func (r *GormQuery) First(dest interface{}) error {
	return r.instance.First(dest).Error
}

func (r *GormQuery) FirstOrCreate(dest interface{}, conds ...interface{}) error {
	var err error
	if len(conds) > 1 {
		err = r.instance.Attrs([]interface{}{conds[1]}...).FirstOrCreate(dest, []interface{}{conds[0]}...).Error
	} else {
		err = r.instance.FirstOrCreate(dest, conds...).Error
	}

	return err
}

func (r *GormQuery) ForceDelete(value interface{}, conds ...interface{}) error {
	return r.instance.Unscoped().Delete(value, conds...).Error
}

func (r *GormQuery) Get(dest interface{}) error {
	return r.instance.Find(dest).Error
}

func (r *GormQuery) Group(name string) contractsorm.Query {
	tx := r.instance.Group(name)

	return NewGormQuery(tx)
}

func (r *GormQuery) Having(query interface{}, args ...interface{}) contractsorm.Query {
	tx := r.instance.Having(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Join(query string, args ...interface{}) contractsorm.Query {
	tx := r.instance.Joins(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Limit(limit int) contractsorm.Query {
	tx := r.instance.Limit(limit)

	return NewGormQuery(tx)
}

func (r *GormQuery) Model(value interface{}) contractsorm.Query {
	tx := r.instance.Model(value)

	return NewGormQuery(tx)
}

func (r *GormQuery) Offset(offset int) contractsorm.Query {
	tx := r.instance.Offset(offset)

	return NewGormQuery(tx)
}

func (r *GormQuery) Order(value interface{}) contractsorm.Query {
	tx := r.instance.Order(value)

	return NewGormQuery(tx)
}

func (r *GormQuery) OrWhere(query interface{}, args ...interface{}) contractsorm.Query {
	tx := r.instance.Or(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Pluck(column string, dest interface{}) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *GormQuery) Raw(sql string, values ...interface{}) contractsorm.Query {
	tx := r.instance.Raw(sql, values...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Save(value interface{}) error {
	return r.instance.Save(value).Error
}

func (r *GormQuery) Scan(dest interface{}) error {
	return r.instance.Scan(dest).Error
}

func (r *GormQuery) Select(query interface{}, args ...interface{}) contractsorm.Query {
	tx := r.instance.Select(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Table(name string, args ...interface{}) contractsorm.Query {
	tx := r.instance.Table(name, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Update(column string, value interface{}) error {
	return r.instance.Update(column, value).Error
}

func (r *GormQuery) Updates(values interface{}) error {
	return r.instance.Updates(values).Error
}

func (r *GormQuery) Where(query interface{}, args ...interface{}) contractsorm.Query {
	tx := r.instance.Where(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) WithTrashed() contractsorm.Query {
	tx := r.instance.Unscoped()

	return NewGormQuery(tx)
}

func (r *GormQuery) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	var gormFuncs []func(*gorm.DB) *gorm.DB
	for _, item := range funcs {
		gormFuncs = append(gormFuncs, func(db *gorm.DB) *gorm.DB {
			item(&GormQuery{db})

			return db
		})
	}

	tx := r.instance.Scopes(gormFuncs...)

	return NewGormQuery(tx)
}

func getGormConfig(connection string) (gorm.Dialector, error) {
	defaultDatabase := facades.Config.GetString("database.default")
	driver := facades.Config.GetString("database.connections." + defaultDatabase + ".driver")

	switch driver {
	case support.Mysql:
		return getMysqlGormConfig(connection), nil
	case support.Postgresql:
		return getPostgresqlGormConfig(connection), nil
	case support.Sqlite:
		return getSqliteGormConfig(connection), nil
	case support.Sqlserver:
		return getSqlserverGormConfig(connection), nil
	default:
		return nil, errors.New(fmt.Sprintf("err database driver: %s, only support mysql, postgresql, sqlite and sqlserver", driver))
	}
}

func getMysqlGormConfig(connection string) gorm.Dialector {
	dsn := support.GetMysqlDsn(connection)
	if dsn == "" {
		return nil
	}

	return mysql.New(mysql.Config{
		DSN: dsn,
	})
}

func getPostgresqlGormConfig(connection string) gorm.Dialector {
	dsn := support.GetPostgresqlDsn(connection)
	if dsn == "" {
		return nil
	}

	return postgres.New(postgres.Config{
		DSN: dsn,
	})
}

func getSqliteGormConfig(connection string) gorm.Dialector {
	dsn := support.GetSqlserverDsn(connection)
	if dsn == "" {
		return nil
	}

	return sqlite.Open(dsn)
}

func getSqlserverGormConfig(connection string) gorm.Dialector {
	dsn := support.GetSqlserverDsn(connection)
	if dsn == "" {
		return nil
	}

	return sqlserver.New(sqlserver.Config{
		DSN: dsn,
	})
}
