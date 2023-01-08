package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/database"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type GormDB struct {
	ormcontract.Query
	instance *gorm.DB
}

func NewGormDB(ctx context.Context, connection string) (ormcontract.DB, error) {
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
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	return gorm.Open(gormConfig, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.LogMode(logLevel),
	})
}

func (r *GormDB) Begin() (ormcontract.Transaction, error) {
	tx := r.instance.Begin()

	return NewGormTransaction(tx), tx.Error
}

type GormTransaction struct {
	ormcontract.Query
	instance *gorm.DB
}

func NewGormTransaction(instance *gorm.DB) ormcontract.Transaction {
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

func NewGormQuery(instance *gorm.DB) ormcontract.Query {
	return &GormQuery{instance}
}

func (r *GormQuery) Driver() ormcontract.Driver {
	return ormcontract.Driver(r.instance.Dialector.Name())
}

func (r *GormQuery) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *GormQuery) Create(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Relationships {
				return r.instance.Create(value).Error
			}
		}

		return r.instance.Create(value).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Relationships {
				return r.instance.Omit(orm.Relationships).Create(value).Error
			}
		}

		return r.instance.Create(value).Error
	}

	return r.instance.Omit(orm.Relationships).Create(value).Error
}

func (r *GormQuery) Delete(value any, conds ...any) error {
	return r.instance.Delete(value, conds...).Error
}

func (r *GormQuery) Distinct(args ...any) ormcontract.Query {
	tx := r.instance.Distinct(args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Exec(sql string, values ...any) error {
	return r.instance.Exec(sql, values...).Error
}

func (r *GormQuery) Find(dest any, conds ...any) error {
	if len(conds) == 1 {
		switch conds[0].(type) {
		case string:
			if conds[0].(string) == "" {
				return ErrorMissingWhereClause
			}
		default:
			reflectValue := reflect.Indirect(reflect.ValueOf(conds[0]))
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

func (r *GormQuery) First(dest any) error {
	err := r.instance.First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	return err
}

func (r *GormQuery) FirstOrCreate(dest any, conds ...any) error {
	var err error
	if len(conds) > 1 {
		err = r.instance.Attrs([]any{conds[1]}...).FirstOrCreate(dest, []any{conds[0]}...).Error
	} else {
		err = r.instance.FirstOrCreate(dest, conds...).Error
	}

	return err
}

func (r *GormQuery) ForceDelete(value any, conds ...any) error {
	return r.instance.Unscoped().Delete(value, conds...).Error
}

func (r *GormQuery) Get(dest any) error {
	return r.instance.Find(dest).Error
}

func (r *GormQuery) Group(name string) ormcontract.Query {
	tx := r.instance.Group(name)

	return NewGormQuery(tx)
}

func (r *GormQuery) Having(query any, args ...any) ormcontract.Query {
	tx := r.instance.Having(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Join(query string, args ...any) ormcontract.Query {
	tx := r.instance.Joins(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Limit(limit int) ormcontract.Query {
	tx := r.instance.Limit(limit)

	return NewGormQuery(tx)
}

func (r *GormQuery) Model(value any) ormcontract.Query {
	tx := r.instance.Model(value)

	return NewGormQuery(tx)
}

func (r *GormQuery) Offset(offset int) ormcontract.Query {
	tx := r.instance.Offset(offset)

	return NewGormQuery(tx)
}

func (r *GormQuery) Omit(columns ...string) ormcontract.Query {
	tx := r.instance.Omit(columns...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Order(value any) ormcontract.Query {
	tx := r.instance.Order(value)

	return NewGormQuery(tx)
}

func (r *GormQuery) OrWhere(query any, args ...any) ormcontract.Query {
	tx := r.instance.Or(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Pluck(column string, dest any) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *GormQuery) Raw(sql string, values ...any) ormcontract.Query {
	tx := r.instance.Raw(sql, values...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Save(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Relationships {
				return r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Save(value).Error
			}
		}

		return r.instance.Save(value).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Relationships {
				return r.instance.Omit(orm.Relationships).Save(value).Error
			}
		}

		return r.instance.Save(value).Error
	}

	return r.instance.Omit(orm.Relationships).Save(value).Error
}

func (r *GormQuery) Scan(dest any) error {
	return r.instance.Scan(dest).Error
}

func (r *GormQuery) Select(query any, args ...any) ormcontract.Query {
	tx := r.instance.Select(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Table(name string, args ...any) ormcontract.Query {
	tx := r.instance.Table(name, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Update(column string, value any) error {
	return r.instance.Update(column, value).Error
}

func (r *GormQuery) Updates(values any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Relationships {
				return r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Updates(values).Error
			}
		}

		return r.instance.Updates(values).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Relationships {
				return r.instance.Omit(orm.Relationships).Updates(values).Error
			}
		}

		return r.instance.Updates(values).Error
	}

	return r.instance.Omit(orm.Relationships).Updates(values).Error
}

func (r *GormQuery) Where(query any, args ...any) ormcontract.Query {
	tx := r.instance.Where(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) WithTrashed() ormcontract.Query {
	tx := r.instance.Unscoped()

	return NewGormQuery(tx)
}

func (r *GormQuery) With(query string, args ...any) ormcontract.Query {
	tx := r.instance.Preload(query, args...)

	return NewGormQuery(tx)
}

func (r *GormQuery) Load(dest any, relation string, relations ...string) error {
	if relation == "" {
		return errors.New("relation cannot be empty")
	}

	id := database.GetID(dest)
	if id == nil {
		return errors.New("id cannot be empty")
	}

	copyDest := copyStruct(dest)
	query := r.With(relation)
	if len(relations) > 0 {
		for _, rel := range relations {
			query = query.With(rel)
		}
	}
	err := query.Find(dest, id)

	relations = append(relations, relation)
	t := reflect.TypeOf(dest).Elem()
	v := reflect.ValueOf(dest).Elem()
	for i := 0; i < t.NumField(); i++ {
		var isRelation bool
		for _, rel := range relations {
			if t.Field(i).Name == rel {
				isRelation = true
				break
			}
		}

		if !isRelation {
			v.Field(i).Set(copyDest.Field(i))
		}
	}

	return err
}

func (r *GormQuery) Scopes(funcs ...func(ormcontract.Query) ormcontract.Query) ormcontract.Query {
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

func copyStruct(dest any) reflect.Value {
	t := reflect.TypeOf(dest).Elem()
	v := reflect.ValueOf(dest).Elem()
	destFields := make([]reflect.StructField, 0)
	for i := 0; i < t.NumField(); i++ {
		destFields = append(destFields, t.Field(i))
	}
	copyDestStruct := reflect.StructOf(destFields)

	return v.Convert(copyDestStruct)
}
