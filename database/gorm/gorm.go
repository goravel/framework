package gorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/database"
	supporttime "github.com/goravel/framework/support/time"
)

func New(connection string) (*gorm.DB, error) {
	readConfigs, writeConfigs, err := Configs(connection)
	if err != nil {
		return nil, err
	}

	dial, err := dialector(connection, writeConfigs[0])
	if err != nil {
		return nil, fmt.Errorf("init gorm dialector error: %v", err)
	}
	if dial == nil {
		return nil, nil
	}

	instance, err := instance(connection, dial)
	if err != nil {
		return nil, err
	}

	if err := configurePool(instance); err != nil {
		return nil, err
	}

	if err := readWriteSeparate(connection, instance, readConfigs, writeConfigs); err != nil {
		return nil, err
	}

	return instance, err
}

func configurePool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(facades.Config.GetInt("database.pool.max_idle_conns", 10))
	sqlDB.SetMaxOpenConns(facades.Config.GetInt("database.pool.max_open_conns", 100))
	sqlDB.SetConnMaxIdleTime(time.Duration(facades.Config.GetInt("database.pool.conn_max_idletime", 3600)) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Duration(facades.Config.GetInt("database.pool.conn_max_lifetime", 3600)) * time.Second)

	return nil
}

func instance(connection string, dialector gorm.Dialector) (*gorm.DB, error) {
	var logLevel gormLogger.LogLevel
	if facades.Config.GetBool("app.debug") {
		logLevel = gormLogger.Info
	} else {
		logLevel = gormLogger.Error
	}

	logger := NewLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	return gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.LogMode(logLevel),
		NowFunc: func() time.Time {
			return supporttime.Now()
		},
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   facades.Config.GetString(fmt.Sprintf("database.connections.%s.prefix", connection)),
			SingularTable: facades.Config.GetBool(fmt.Sprintf("database.connections.%s.singular", connection)),
		},
	})
}

func readWriteSeparate(connection string, instance *gorm.DB, readConfigs, writeConfigs []contractsdatabase.Config) error {
	if len(readConfigs) == 0 || len(writeConfigs) == 0 {
		return nil
	}

	readDialectors, err := dialectors(connection, readConfigs)
	if err != nil {
		return err
	}

	writeDialectors, err := dialectors(connection, writeConfigs)
	if err != nil {
		return err
	}

	return instance.Use(dbresolver.Register(dbresolver.Config{
		Sources:           writeDialectors,
		Replicas:          readDialectors,
		Policy:            dbresolver.RandomPolicy{},
		TraceResolverMode: true,
	}))
}

type DB struct {
	contractsorm.Query
	instance *gorm.DB
}

func NewDB(ctx context.Context, connection string) (*DB, error) {
	db, err := New(connection)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return &DB{
		Query:    NewQuery(db),
		instance: db,
	}, nil
}

func (r *DB) Begin() (contractsorm.Transaction, error) {
	tx := r.instance.Begin()

	return NewTransaction(tx), tx.Error
}

func (r *DB) Instance() *gorm.DB {
	return r.instance
}

type Transaction struct {
	contractsorm.Query
	instance *gorm.DB
}

func NewTransaction(instance *gorm.DB) *Transaction {
	return &Transaction{Query: NewQuery(instance), instance: instance}
}

func (r *Transaction) Commit() error {
	return r.instance.Commit().Error
}

func (r *Transaction) Rollback() error {
	return r.instance.Rollback().Error
}

type Query struct {
	instance *gorm.DB
}

func NewQuery(instance *gorm.DB) *Query {
	return &Query{instance}
}

func (r *Query) Association(association string) contractsorm.Association {
	return r.instance.Association(association)
}

func (r *Query) Driver() contractsorm.Driver {
	return contractsorm.Driver(r.instance.Dialector.Name())
}

func (r *Query) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *Query) Create(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		if len(r.instance.Statement.Selects) == 1 && r.instance.Statement.Selects[0] == orm.Associations {
			r.instance.Statement.Selects = []string{}
			return r.instance.Create(value).Error
		}

		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				return errors.New("cannot set orm.Associations and other fields at the same time")
			}
		}

		return r.instance.Create(value).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		if len(r.instance.Statement.Omits) == 1 && r.instance.Statement.Omits[0] == orm.Associations {
			r.instance.Statement.Selects = []string{}
			return r.instance.Omit(orm.Associations).Create(value).Error
		}

		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				return errors.New("cannot set orm.Associations and other fields at the same time")
			}
		}

		return r.instance.Create(value).Error
	}

	return r.instance.Omit(orm.Associations).Create(value).Error
}

func (r *Query) Delete(value any, conds ...any) error {
	return r.instance.Delete(value, conds...).Error
}

func (r *Query) Distinct(args ...any) contractsorm.Query {
	tx := r.instance.Distinct(args...)

	return NewQuery(tx)
}

func (r *Query) Exec(sql string, values ...any) error {
	return r.instance.Exec(sql, values...).Error
}

func (r *Query) Find(dest any, conds ...any) error {
	if len(conds) == 1 {
		switch cond := conds[0].(type) {
		case string:
			if cond == "" {
				return ErrorMissingWhereClause
			}
		default:
			reflectValue := reflect.Indirect(reflect.ValueOf(cond))
			switch reflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				if reflectValue.Len() == 0 {
					return ErrorMissingWhereClause
				}
			}
		}
	}

	return r.instance.Find(dest, conds...).Error
}

func (r *Query) First(dest any) error {
	err := r.instance.First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	return err
}

func (r *Query) FirstOrCreate(dest any, conds ...any) error {
	var err error
	if len(conds) > 1 {
		err = r.instance.Attrs([]any{conds[1]}...).FirstOrCreate(dest, []any{conds[0]}...).Error
	} else {
		err = r.instance.FirstOrCreate(dest, conds...).Error
	}

	return err
}

func (r *Query) ForceDelete(value any, conds ...any) error {
	return r.instance.Unscoped().Delete(value, conds...).Error
}

func (r *Query) Get(dest any) error {
	return r.instance.Find(dest).Error
}

func (r *Query) Group(name string) contractsorm.Query {
	tx := r.instance.Group(name)

	return NewQuery(tx)
}

func (r *Query) Having(query any, args ...any) contractsorm.Query {
	tx := r.instance.Having(query, args...)

	return NewQuery(tx)
}

func (r *Query) Join(query string, args ...any) contractsorm.Query {
	tx := r.instance.Joins(query, args...)

	return NewQuery(tx)
}

func (r *Query) Limit(limit int) contractsorm.Query {
	tx := r.instance.Limit(limit)

	return NewQuery(tx)
}

func (r *Query) Load(model any, relation string, args ...any) error {
	if relation == "" {
		return errors.New("relation cannot be empty")
	}

	destType := reflect.TypeOf(model)
	if destType.Kind() != reflect.Pointer {
		return errors.New("model must be pointer")
	}

	if id := database.GetID(model); id == nil {
		return errors.New("id cannot be empty")
	}

	copyDest := copyStruct(model)
	query := r.With(relation, args...)
	err := query.Find(model)

	t := destType.Elem()
	v := reflect.ValueOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name != relation {
			v.Field(i).Set(copyDest.Field(i))
		}
	}

	return err
}

func (r *Query) LoadMissing(model any, relation string, args ...any) error {
	destType := reflect.TypeOf(model)
	if destType.Kind() != reflect.Pointer {
		return errors.New("model must be pointer")
	}

	t := reflect.TypeOf(model).Elem()
	v := reflect.ValueOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == relation {
			var id any
			if v.Field(i).Kind() == reflect.Pointer {
				if !v.Field(i).IsNil() {
					id = database.GetIDByReflect(v.Field(i).Type().Elem(), v.Field(i).Elem())
				}
			} else if v.Field(i).Kind() == reflect.Slice {
				if v.Field(i).Len() > 0 {
					return nil
				}
			} else {
				id = database.GetIDByReflect(v.Field(i).Type(), v.Field(i))
			}
			if cast.ToString(id) != "" {
				return nil
			}
		}
	}

	return r.Load(model, relation, args...)
}

func (r *Query) Model(value any) contractsorm.Query {
	tx := r.instance.Model(value)

	return NewQuery(tx)
}

func (r *Query) Offset(offset int) contractsorm.Query {
	tx := r.instance.Offset(offset)

	return NewQuery(tx)
}

func (r *Query) Omit(columns ...string) contractsorm.Query {
	tx := r.instance.Omit(columns...)

	return NewQuery(tx)
}

func (r *Query) Order(value any) contractsorm.Query {
	tx := r.instance.Order(value)

	return NewQuery(tx)
}

func (r *Query) OrWhere(query any, args ...any) contractsorm.Query {
	tx := r.instance.Or(query, args...)

	return NewQuery(tx)
}

func (r *Query) Paginate(page, limit int, dest any, total *int64) error {
	offset := (page - 1) * limit
	if total != nil {
		if r.instance.Statement.Table == "" && r.instance.Statement.Model == nil {
			if err := r.Model(dest).Count(total); err != nil {
				return err
			}
		} else {
			if err := r.Count(total); err != nil {
				return err
			}
		}
	}

	return r.Offset(offset).Limit(limit).Find(dest)
}

func (r *Query) Pluck(column string, dest any) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *Query) Raw(sql string, values ...any) contractsorm.Query {
	tx := r.instance.Raw(sql, values...)

	return NewQuery(tx)
}

func (r *Query) Save(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				return r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Save(value).Error
			}
		}

		return r.instance.Save(value).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				return r.instance.Omit(orm.Associations).Save(value).Error
			}
		}

		return r.instance.Save(value).Error
	}

	return r.instance.Omit(orm.Associations).Save(value).Error
}

func (r *Query) Scan(dest any) error {
	return r.instance.Scan(dest).Error
}

func (r *Query) Select(query any, args ...any) contractsorm.Query {
	tx := r.instance.Select(query, args...)

	return NewQuery(tx)
}

func (r *Query) Table(name string, args ...any) contractsorm.Query {
	tx := r.instance.Table(name, args...)

	return NewQuery(tx)
}

func (r *Query) Update(column string, value any) error {
	return r.instance.Update(column, value).Error
}

func (r *Query) Updates(values any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				return r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Updates(values).Error
			}
		}

		return r.instance.Updates(values).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				return r.instance.Omit(orm.Associations).Updates(values).Error
			}
		}

		return r.instance.Updates(values).Error
	}

	return r.instance.Omit(orm.Associations).Updates(values).Error
}

func (r *Query) Where(query any, args ...any) contractsorm.Query {
	tx := r.instance.Where(query, args...)

	return NewQuery(tx)
}

func (r *Query) WithTrashed() contractsorm.Query {
	tx := r.instance.Unscoped()

	return NewQuery(tx)
}

func (r *Query) With(query string, args ...any) contractsorm.Query {
	if len(args) == 1 {
		switch arg := args[0].(type) {
		case func(contractsorm.Query) contractsorm.Query:
			newArgs := []any{
				func(db *gorm.DB) *gorm.DB {
					query := arg(NewQuery(db))

					return query.(*Query).instance
				},
			}

			tx := r.instance.Preload(query, newArgs...)

			return NewQuery(tx)
		}
	}

	tx := r.instance.Preload(query, args...)

	return NewQuery(tx)
}

func (r *Query) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	var gormFuncs []func(*gorm.DB) *gorm.DB
	for _, item := range funcs {
		gormFuncs = append(gormFuncs, func(db *gorm.DB) *gorm.DB {
			item(&Query{db})

			return db
		})
	}

	tx := r.instance.Scopes(gormFuncs...)

	return NewQuery(tx)
}
