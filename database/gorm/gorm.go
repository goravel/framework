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

type Transaction struct {
	contractsorm.Query
	instance *gorm.DB
}

func NewTransaction(tx *gorm.DB) *Transaction {
	return &Transaction{Query: NewQueryWithInstance(nil, tx), instance: tx}
}

func (r *Transaction) Commit() error {
	return r.instance.Commit().Error
}

func (r *Transaction) Rollback() error {
	return r.instance.Rollback().Error
}

type Query struct {
	instance      *gorm.DB
	withoutEvents bool
}

func NewQuery(ctx context.Context, connection string) (*Query, error) {
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

	return NewQueryWithInstance(nil, db), nil
}

func NewQueryWithInstance(query *Query, instance *gorm.DB) *Query {
	if query == nil {
		return &Query{instance: instance}
	}

	return &Query{instance: instance, withoutEvents: query.withoutEvents}
}

func NewQueryWithWithoutEvents(query *Query) *Query {
	return &Query{instance: query.instance, withoutEvents: true}
}

func (r *Query) Association(association string) contractsorm.Association {
	return r.instance.Association(association)
}

func (r *Query) Begin() (contractsorm.Transaction, error) {
	tx := r.instance.Begin()

	return NewTransaction(tx), tx.Error
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
		return r.selectCreate(value)
	}

	if len(r.instance.Statement.Omits) > 0 {
		return r.omitCreate(value)
	}

	return r.create(value)
}

func (r *Query) Delete(dest any, conds ...any) (*contractsorm.Result, error) {
	if err := deleting(r, dest); err != nil {
		return nil, err
	}

	res := r.instance.Delete(dest, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if err := deleted(r, dest); err != nil {
		return nil, err
	}

	return &contractsorm.Result{
		RowsAffected: res.RowsAffected,
	}, nil
}

func (r *Query) Distinct(args ...any) contractsorm.Query {
	tx := r.instance.Distinct(args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Exec(sql string, values ...any) (*contractsorm.Result, error) {
	result := r.instance.Exec(sql, values...)

	return &contractsorm.Result{
		RowsAffected: result.RowsAffected,
	}, result.Error
}

func (r *Query) Find(dest any, conds ...any) error {
	if err := filterFindConditions(conds...); err != nil {
		return err
	}
	if err := r.instance.Find(dest, conds...).Error; err != nil {
		return err
	}

	return retrieved(r, dest)
}

func (r *Query) FindOrFail(dest any, conds ...any) error {
	if err := filterFindConditions(conds...); err != nil {
		return err
	}

	res := r.instance.Find(dest, conds...)
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return orm.ErrRecordNotFound
	}

	return retrieved(r, dest)
}

func (r *Query) First(dest any) error {
	res := r.instance.First(dest)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}

		return res.Error
	}

	return retrieved(r, dest)
}

func (r *Query) FirstOr(dest any, callback func() error) error {
	err := r.instance.First(dest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return callback()
		}

		return err
	}

	return retrieved(r, dest)
}

func (r *Query) FirstOrCreate(dest any, conds ...any) error {
	if len(conds) == 0 {
		return errors.New("query condition is require")
	}

	var res *gorm.DB
	if len(conds) > 1 {
		res = r.instance.Attrs(conds[1]).FirstOrInit(dest, conds[0])
	} else {
		res = r.instance.FirstOrInit(dest, conds[0])
	}

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return retrieved(r, dest)
	}

	return r.Create(dest)
}

func (r *Query) FirstOrFail(dest any) error {
	err := r.instance.First(dest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orm.ErrRecordNotFound
		}

		return err
	}

	return retrieved(r, dest)
}

func (r *Query) FirstOrNew(dest any, attributes any, values ...any) error {
	var res *gorm.DB
	if len(values) > 0 {
		res = r.instance.Attrs(values[0]).FirstOrInit(dest, attributes)
	} else {
		res = r.instance.FirstOrInit(dest, attributes)
	}

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return retrieved(r, dest)
	}

	return nil
}

func (r *Query) ForceDelete(value any, conds ...any) (*contractsorm.Result, error) {
	if err := forceDeleting(r, value); err != nil {
		return nil, err
	}

	res := r.instance.Unscoped().Delete(value, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected > 0 {
		if err := forceDeleted(r, value); err != nil {
			return nil, err
		}
	}

	return &contractsorm.Result{
		RowsAffected: res.RowsAffected,
	}, res.Error
}

func (r *Query) Get(dest any) error {
	return r.Find(dest)
}

func (r *Query) Group(name string) contractsorm.Query {
	tx := r.instance.Group(name)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Having(query any, args ...any) contractsorm.Query {
	tx := r.instance.Having(query, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Instance() *gorm.DB {
	return r.instance
}

func (r *Query) Join(query string, args ...any) contractsorm.Query {
	tx := r.instance.Joins(query, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Limit(limit int) contractsorm.Query {
	tx := r.instance.Limit(limit)

	return NewQueryWithInstance(r, tx)
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

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Offset(offset int) contractsorm.Query {
	tx := r.instance.Offset(offset)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Omit(columns ...string) contractsorm.Query {
	tx := r.instance.Omit(columns...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Order(value any) contractsorm.Query {
	tx := r.instance.Order(value)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) OrWhere(query any, args ...any) contractsorm.Query {
	tx := r.instance.Or(query, args...)

	return NewQueryWithInstance(r, tx)
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

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Save(value any) error {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		return r.selectSave(value)
	}

	if len(r.instance.Statement.Omits) > 0 {
		return r.omitSave(value)
	}

	return r.save(value)
}

func (r *Query) SaveQuietly(value any) error {
	return r.WithoutEvents().Save(value)
}

func (r *Query) Scan(dest any) error {
	return r.instance.Scan(dest).Error
}

func (r *Query) Select(query any, args ...any) contractsorm.Query {
	tx := r.instance.Select(query, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Scopes(funcs ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	var gormFuncs []func(*gorm.DB) *gorm.DB
	for _, item := range funcs {
		gormFuncs = append(gormFuncs, func(tx *gorm.DB) *gorm.DB {
			item(NewQueryWithInstance(r, tx))

			return tx
		})
	}

	tx := r.instance.Scopes(gormFuncs...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Table(name string, args ...any) contractsorm.Query {
	tx := r.instance.Table(name, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Update(column string, value any) error {
	return r.instance.Update(column, value).Error
}

func (r *Query) Updates(values any) (*contractsorm.Result, error) {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return nil, errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				result := r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Updates(values)
				return &contractsorm.Result{
					RowsAffected: result.RowsAffected,
				}, result.Error
			}
		}

		result := r.instance.Updates(values)

		return &contractsorm.Result{
			RowsAffected: result.RowsAffected,
		}, result.Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				result := r.instance.Omit(orm.Associations).Updates(values)

				return &contractsorm.Result{
					RowsAffected: result.RowsAffected,
				}, result.Error
			}
		}
		result := r.instance.Updates(values)

		return &contractsorm.Result{
			RowsAffected: result.RowsAffected,
		}, result.Error
	}
	result := r.instance.Omit(orm.Associations).Updates(values)

	return &contractsorm.Result{
		RowsAffected: result.RowsAffected,
	}, result.Error
}

func (r *Query) UpdateOrCreate(dest any, attributes any, values any) error {
	res := r.instance.Assign(values).FirstOrInit(dest, attributes)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return r.Save(dest)
	}

	return r.Create(dest)
}

func (r *Query) Where(query any, args ...any) contractsorm.Query {
	tx := r.instance.Where(query, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) WithoutEvents() contractsorm.Query {
	return NewQueryWithWithoutEvents(r)
}

func (r *Query) WithTrashed() contractsorm.Query {
	tx := r.instance.Unscoped()

	return NewQueryWithInstance(r, tx)
}

func (r *Query) With(query string, args ...any) contractsorm.Query {
	if len(args) == 1 {
		switch arg := args[0].(type) {
		case func(contractsorm.Query) contractsorm.Query:
			newArgs := []any{
				func(tx *gorm.DB) *gorm.DB {
					query := arg(NewQueryWithInstance(r, tx))

					return query.(*Query).instance
				},
			}

			tx := r.instance.Preload(query, newArgs...)

			return NewQueryWithInstance(r, tx)
		}
	}

	tx := r.instance.Preload(query, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) selectCreate(value any) error {
	if len(r.instance.Statement.Selects) == 1 && r.instance.Statement.Selects[0] == orm.Associations {
		r.instance.Statement.Selects = []string{}

		return create(r, value)
	}

	for _, val := range r.instance.Statement.Selects {
		if val == orm.Associations {
			return errors.New("cannot set orm.Associations and other fields at the same time")
		}
	}

	return create(r, value)
}

func (r *Query) omitCreate(value any) error {
	if len(r.instance.Statement.Omits) == 1 && r.instance.Statement.Omits[0] == orm.Associations {
		r.instance.Statement.Selects = []string{}
		if err := saving(r, value); err != nil {
			return err
		}
		if err := creating(r, value); err != nil {
			return err
		}
		if err := r.instance.Omit(orm.Associations).Create(value).Error; err != nil {
			return err
		}
		if err := created(r, value); err != nil {
			return err
		}
		if err := saved(r, value); err != nil {
			return err
		}

		return nil
	}

	for _, val := range r.instance.Statement.Omits {
		if val == orm.Associations {
			return errors.New("cannot set orm.Associations and other fields at the same time")
		}
	}

	return create(r, value)
}

func (r *Query) create(value any) error {
	if err := saving(r, value); err != nil {
		return err
	}
	if err := creating(r, value); err != nil {
		return err
	}

	if err := r.instance.Omit(orm.Associations).Create(value).Error; err != nil {
		return err
	}

	if err := created(r, value); err != nil {
		return err
	}
	if err := saved(r, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) selectSave(value any) error {
	for _, val := range r.instance.Statement.Selects {
		if val == orm.Associations {
			if err := saving(r, value); err != nil {
				return err
			}
			if err := updating(r, value); err != nil {
				return err
			}
			if err := r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Save(value).Error; err != nil {
				return err
			}
			if err := updated(r, value); err != nil {
				return err
			}
			if err := saved(r, value); err != nil {
				return err
			}

			return nil
		}
	}

	if err := saving(r, value); err != nil {
		return err
	}
	if err := updating(r, value); err != nil {
		return err
	}
	if err := r.instance.Save(value).Error; err != nil {
		return err
	}
	if err := updated(r, value); err != nil {
		return err
	}
	if err := saved(r, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) omitSave(value any) error {
	for _, val := range r.instance.Statement.Omits {
		if val == orm.Associations {
			if err := saving(r, value); err != nil {
				return err
			}
			if err := updating(r, value); err != nil {
				return err
			}
			if err := r.instance.Omit(orm.Associations).Save(value).Error; err != nil {
				return err
			}
			if err := updated(r, value); err != nil {
				return err
			}
			if err := saved(r, value); err != nil {
				return err
			}

			return nil
		}
	}

	if err := saving(r, value); err != nil {
		return err
	}
	if err := updating(r, value); err != nil {
		return err
	}
	if err := r.instance.Save(value).Error; err != nil {
		return err
	}
	if err := updated(r, value); err != nil {
		return err
	}
	if err := saved(r, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) save(value any) error {
	if err := saving(r, value); err != nil {
		return err
	}
	if err := updating(r, value); err != nil {
		return err
	}
	if err := r.instance.Omit(orm.Associations).Save(value).Error; err != nil {
		return err
	}
	if err := updated(r, value); err != nil {
		return err
	}
	if err := saved(r, value); err != nil {
		return err
	}

	return nil
}

func retrieved(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if retrievedModel, ok := dest.(contractsorm.Retrieved); ok {
		if err := retrievedModel.Retrieved(query); err != nil {
			return err
		}
	}

	return nil
}

func updating(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if updatingModel, ok := dest.(contractsorm.Updating); ok {
		if err := updatingModel.Updating(query); err != nil {
			return err
		}
	}

	return nil
}

func updated(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if updatedModel, ok := dest.(contractsorm.Updated); ok {
		if err := updatedModel.Updated(query); err != nil {
			return err
		}
	}

	return nil
}

func saving(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if savingModel, ok := dest.(contractsorm.Saving); ok {
		if err := savingModel.Saving(query); err != nil {
			return err
		}
	}

	return nil
}

func saved(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if savedModel, ok := dest.(contractsorm.Saved); ok {
		if err := savedModel.Saved(query); err != nil {
			return err
		}
	}

	return nil
}

func creating(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if creatingModel, ok := dest.(contractsorm.Creating); ok {
		if err := creatingModel.Creating(query); err != nil {
			return err
		}
	}

	return nil
}

func created(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if createdModel, ok := dest.(contractsorm.Created); ok {
		if err := createdModel.Created(query); err != nil {
			return err
		}
	}

	return nil
}

func deleting(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if deletingModel, ok := dest.(contractsorm.Deleting); ok {
		if err := deletingModel.Deleting(query); err != nil {
			return err
		}
	}

	return nil
}

func deleted(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if deletedModel, ok := dest.(contractsorm.Deleted); ok {
		if err := deletedModel.Deleted(query); err != nil {
			return err
		}
	}

	return nil
}

func forceDeleting(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if forceDeletingModel, ok := dest.(contractsorm.ForceDeleting); ok {
		if err := forceDeletingModel.ForceDeleting(query); err != nil {
			return err
		}
	}

	return nil
}

func forceDeleted(query *Query, dest any) error {
	if query.withoutEvents {
		return nil
	}

	if forceDeletedModel, ok := dest.(contractsorm.ForceDeleted); ok {
		if err := forceDeletedModel.ForceDeleted(query); err != nil {
			return err
		}
	}

	return nil
}

func create(query *Query, dest any) error {
	if err := saving(query, dest); err != nil {
		return err
	}
	if err := creating(query, dest); err != nil {
		return err
	}

	if err := query.instance.Create(dest).Error; err != nil {
		return err
	}

	if err := created(query, dest); err != nil {
		return err
	}
	if err := saved(query, dest); err != nil {
		return err
	}

	return nil
}

func filterFindConditions(conds ...any) error {
	if len(conds) > 0 {
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

	return nil
}
