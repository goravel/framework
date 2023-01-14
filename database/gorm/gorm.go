package gorm

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
	databasesupport "github.com/goravel/framework/database/support"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/database"

	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func New(connection string) (*gorm.DB, error) {
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

	logger := NewLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
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

type DB struct {
	ormcontract.Query
	instance *gorm.DB
}

func NewDB(ctx context.Context, connection string) (ormcontract.DB, error) {
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

func (r *DB) Begin() (ormcontract.Transaction, error) {
	tx := r.instance.Begin()

	return NewTransaction(tx), tx.Error
}

type Transaction struct {
	ormcontract.Query
	instance *gorm.DB
}

func NewTransaction(instance *gorm.DB) ormcontract.Transaction {
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

func NewQuery(instance *gorm.DB) ormcontract.Query {
	return &Query{instance}
}

func (r *Query) Driver() ormcontract.Driver {
	return ormcontract.Driver(r.instance.Dialector.Name())
}

func (r *Query) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *Query) Create(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		if len(r.instance.Statement.Selects) == 1 && r.instance.Statement.Selects[0] == orm.Relationships {
			r.instance.Statement.Selects = []string{}
			return r.instance.Create(value).Error
		}

		for _, val := range r.instance.Statement.Selects {
			if val == orm.Relationships {
				return errors.New("cannot set orm.Relationships and other fields at the same time")
			}
		}

		return r.instance.Create(value).Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		if len(r.instance.Statement.Omits) == 1 && r.instance.Statement.Omits[0] == orm.Relationships {
			r.instance.Statement.Selects = []string{}
			return r.instance.Omit(orm.Relationships).Create(value).Error
		}

		for _, val := range r.instance.Statement.Omits {
			if val == orm.Relationships {
				return errors.New("cannot set orm.Relationships and other fields at the same time")
			}
		}

		return r.instance.Create(value).Error
	}

	return r.instance.Omit(orm.Relationships).Create(value).Error
}

func (r *Query) Delete(value any, conds ...any) error {
	return r.instance.Delete(value, conds...).Error
}

func (r *Query) Distinct(args ...any) ormcontract.Query {
	tx := r.instance.Distinct(args...)

	return NewQuery(tx)
}

func (r *Query) Exec(sql string, values ...any) error {
	return r.instance.Exec(sql, values...).Error
}

func (r *Query) Find(dest any, conds ...any) error {
	if len(conds) == 1 {
		switch conds[0].(type) {
		case string:
			if conds[0].(string) == "" {
				return databasesupport.ErrorMissingWhereClause
			}
		default:
			reflectValue := reflect.Indirect(reflect.ValueOf(conds[0]))
			switch reflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				if reflectValue.Len() == 0 {
					return databasesupport.ErrorMissingWhereClause
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

func (r *Query) Group(name string) ormcontract.Query {
	tx := r.instance.Group(name)

	return NewQuery(tx)
}

func (r *Query) Having(query any, args ...any) ormcontract.Query {
	tx := r.instance.Having(query, args...)

	return NewQuery(tx)
}

func (r *Query) Join(query string, args ...any) ormcontract.Query {
	tx := r.instance.Joins(query, args...)

	return NewQuery(tx)
}

func (r *Query) Limit(limit int) ormcontract.Query {
	tx := r.instance.Limit(limit)

	return NewQuery(tx)
}

func (r *Query) Load(dest any, relation string, args ...any) error {
	if relation == "" {
		return errors.New("relation cannot be empty")
	}

	id := database.GetID(dest)
	if id == nil {
		return errors.New("id cannot be empty")
	}

	copyDest := copyStruct(dest)
	query := r.With(relation, args...)
	err := query.Find(dest, id)

	t := reflect.TypeOf(dest).Elem()
	v := reflect.ValueOf(dest).Elem()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name != relation {
			v.Field(i).Set(copyDest.Field(i))
		}
	}

	return err
}

func (r *Query) Model(value any) ormcontract.Query {
	tx := r.instance.Model(value)

	return NewQuery(tx)
}

func (r *Query) Offset(offset int) ormcontract.Query {
	tx := r.instance.Offset(offset)

	return NewQuery(tx)
}

func (r *Query) Omit(columns ...string) ormcontract.Query {
	tx := r.instance.Omit(columns...)

	return NewQuery(tx)
}

func (r *Query) Order(value any) ormcontract.Query {
	tx := r.instance.Order(value)

	return NewQuery(tx)
}

func (r *Query) OrWhere(query any, args ...any) ormcontract.Query {
	tx := r.instance.Or(query, args...)

	return NewQuery(tx)
}

func (r *Query) Pluck(column string, dest any) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *Query) Raw(sql string, values ...any) ormcontract.Query {
	tx := r.instance.Raw(sql, values...)

	return NewQuery(tx)
}

func (r *Query) Save(value any) error {
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

func (r *Query) Scan(dest any) error {
	return r.instance.Scan(dest).Error
}

func (r *Query) Select(query any, args ...any) ormcontract.Query {
	tx := r.instance.Select(query, args...)

	return NewQuery(tx)
}

func (r *Query) Table(name string, args ...any) ormcontract.Query {
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

func (r *Query) Where(query any, args ...any) ormcontract.Query {
	tx := r.instance.Where(query, args...)

	return NewQuery(tx)
}

func (r *Query) WithTrashed() ormcontract.Query {
	tx := r.instance.Unscoped()

	return NewQuery(tx)
}

func (r *Query) With(query string, args ...any) ormcontract.Query {
	if len(args) == 1 {
		switch args[0].(type) {
		case func(ormcontract.Query) ormcontract.Query:
			newArgs := []any{
				func(db *gorm.DB) *gorm.DB {
					query := args[0].(func(query ormcontract.Query) ormcontract.Query)(NewQuery(db))

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

func (r *Query) Scopes(funcs ...func(ormcontract.Query) ormcontract.Query) ormcontract.Query {
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
