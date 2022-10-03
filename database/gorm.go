package database

import (
	"context"
	"errors"
	"fmt"

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

type Gorm struct {
	instance *gorm.DB
}

func NewGorm(ctx context.Context, connection string) (contractsorm.DB, error) {
	db, err := NewGormInstance(connection)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("gorm open ddatabase error: %v", err))
	}

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return &Gorm{
		instance: db,
	}, nil
}

func NewGormInstance(connection string) (*gorm.DB, error) {
	gormConfig, err := getGormConfig(connection)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init gorm config error: %v", err))
	}

	var logLevel gormLogger.LogLevel
	if facades.Config.GetBool("app.debug") {
		logLevel = gormLogger.Info
	} else {
		logLevel = gormLogger.Error
	}

	return gorm.Open(gormConfig, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormLogger.Default.LogMode(logLevel),
	})
}

func (r *Gorm) Begin() (contractsorm.Transaction, error) {
	r.instance = r.instance.Begin(nil)

	return r, r.instance.Error
}

func (r *Gorm) Commit() error {
	r.instance = r.instance.Commit()

	return r.instance.Error
}

func (r *Gorm) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *Gorm) Create(value interface{}) error {
	return r.instance.Create(value).Error
}

func (r *Gorm) Delete(value interface{}, conds ...interface{}) error {
	return r.instance.Delete(value, conds).Error
}

func (r *Gorm) Exec(sql string, values ...interface{}) error {
	return r.instance.Exec(sql, values).Error
}

func (r *Gorm) Find(dest interface{}, conds ...interface{}) error {
	return r.instance.Find(dest, conds...).Error
}

func (r *Gorm) First(dest interface{}) error {
	return r.instance.First(dest).Error
}

func (r *Gorm) FirstOrCreate(dest interface{}, conds ...interface{}) error {
	return r.instance.FirstOrCreate(dest, conds...).Error
}

func (r *Gorm) ForceDelete(value interface{}, conds ...interface{}) error {
	return r.instance.Unscoped().Delete(value, conds).Error
}

func (r *Gorm) Get(dest interface{}) error {
	return r.instance.Find(dest).Error
}

func (r *Gorm) Group(name string) contractsorm.Query {
	r.instance = r.instance.Group(name)

	return r
}

func (r *Gorm) Having(query interface{}, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Having(query, args...)

	return r
}

func (r *Gorm) Join(query string, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Joins(query, args...)

	return r
}

func (r *Gorm) Limit(limit int) contractsorm.Query {
	r.instance = r.instance.Limit(limit)

	return r
}

func (r *Gorm) Model(value interface{}) contractsorm.Query {
	r.instance = r.instance.Model(value)

	return r
}

func (r *Gorm) Offset(offset int) contractsorm.Query {
	r.instance = r.instance.Offset(offset)

	return r
}

func (r *Gorm) Order(value interface{}) contractsorm.Query {
	r.instance = r.instance.Order(value)

	return r
}

func (r *Gorm) OrWhere(query interface{}, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Or(query, args...)

	return r
}

func (r *Gorm) Pluck(column string, dest interface{}) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *Gorm) Raw(sql string, values ...interface{}) contractsorm.Query {
	r.instance = r.instance.Raw(sql, values)

	return r
}

func (r *Gorm) Rollback() error {
	r.instance = r.instance.Rollback()

	return r.instance.Error
}

func (r *Gorm) Save(value interface{}) error {
	return r.instance.Save(value).Error
}

func (r *Gorm) Scan(dest interface{}) error {
	return r.instance.Scan(dest).Error
}

func (r *Gorm) Select(query interface{}, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Select(query, args...)

	return r
}

func (r *Gorm) Table(name string, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Table(name, args...)

	return r
}

func (r *Gorm) Update(column string, value interface{}) error {
	return r.instance.Update(column, value).Error
}

func (r *Gorm) Updates(values interface{}) error {
	return r.instance.Updates(values).Error
}

func (r *Gorm) Where(query interface{}, args ...interface{}) contractsorm.Query {
	r.instance = r.instance.Where(query, args...)

	return r
}

func (r *Gorm) WithTrashed() contractsorm.Query {
	r.instance = r.instance.Unscoped()

	return r
}

func (r *Gorm) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	var gormFuncs []func(*gorm.DB) *gorm.DB
	for _, item := range funcs {
		gormFuncs = append(gormFuncs, func(db *gorm.DB) *gorm.DB {
			r.instance = db
			item(r)

			return r.instance
		})
	}

	r.instance = r.instance.Scopes(gormFuncs...)

	return r
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
		return nil, errors.New("database driver only support mysql, postgresql, sqlite and sqlserver")
	}
}

func getMysqlGormConfig(connection string) gorm.Dialector {
	return mysql.New(mysql.Config{
		DSN: support.GetMysqlDsn(connection),
	})
}

func getPostgresqlGormConfig(connection string) gorm.Dialector {
	return postgres.New(postgres.Config{
		DSN: support.GetPostgresqlDsn(connection),
	})
}

func getSqliteGormConfig(connection string) gorm.Dialector {
	return sqlite.Open(support.GetSqliteDsn(connection))
}

func getSqlserverGormConfig(connection string) gorm.Dialector {
	return sqlserver.New(sqlserver.Config{
		DSN: support.GetSqlserverDsn(connection),
	})
}
