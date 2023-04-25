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
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm/hints"
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
	if err := r.deleting(dest); err != nil {
		return nil, err
	}

	res := r.instance.Delete(dest, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if err := r.deleted(dest); err != nil {
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

	return r.retrieved(dest)
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

	return r.retrieved(dest)
}

func (r *Query) First(dest any) error {
	res := r.instance.First(dest)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}

		return res.Error
	}

	return r.retrieved(dest)
}

func (r *Query) FirstOr(dest any, callback func() error) error {
	err := r.instance.First(dest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return callback()
		}

		return err
	}

	return r.retrieved(dest)
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
		return r.retrieved(dest)
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

	return r.retrieved(dest)
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
		return r.retrieved(dest)
	}

	return nil
}

func (r *Query) ForceDelete(value any, conds ...any) (*contractsorm.Result, error) {
	if err := r.forceDeleting(value); err != nil {
		return nil, err
	}

	res := r.instance.Unscoped().Delete(value, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected > 0 {
		if err := r.forceDeleted(value); err != nil {
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

func (r *Query) LockForUpdate() contractsorm.Query {
	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		tx := r.instance.Clauses(clause.Locking{Strength: "UPDATE"})

		return NewQueryWithInstance(r, tx)
	} else if driver == sqlserverDialector.Name() {
		tx := r.instance.Clauses(hints.With("rowlock", "updlock", "holdlock"))

		return NewQueryWithInstance(r, tx)
	}

	return r
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

	model := r.instance.Statement.Model
	id := database.GetID(value)
	update := id != nil

	if err := r.saving(model, value); err != nil {
		return err
	}
	if update {
		if err := r.updating(model, value); err != nil {
			return err
		}
	} else {
		if err := r.creating(value); err != nil {
			return err
		}
	}

	if len(r.instance.Statement.Selects) > 0 {
		if err := r.selectSave(value); err != nil {
			return err
		}
	} else if len(r.instance.Statement.Omits) > 0 {
		if err := r.omitSave(value); err != nil {
			return err
		}
	} else {
		if err := r.save(value); err != nil {
			return err
		}
	}

	if update {
		if err := r.updated(model, value); err != nil {
			return err
		}
	} else {
		if err := r.created(value); err != nil {
			return err
		}
	}
	if err := r.saved(model, value); err != nil {
		return err
	}

	return nil
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

func (r *Query) SharedLock() contractsorm.Query {
	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		tx := r.instance.Clauses(clause.Locking{Strength: "SHARE"})

		return NewQueryWithInstance(r, tx)
	} else if driver == sqlserverDialector.Name() {
		tx := r.instance.Clauses(hints.With("rowlock", "holdlock"))

		return NewQueryWithInstance(r, tx)
	}

	return r
}

func (r *Query) Table(name string, args ...any) contractsorm.Query {
	tx := r.instance.Table(name, args...)

	return NewQueryWithInstance(r, tx)
}

func (r *Query) Update(column any, value ...any) (*contractsorm.Result, error) {
	if _, ok := column.(string); !ok && len(value) > 0 {
		return nil, errors.New("parameter error, please check the document")
	}

	var singleUpdate bool
	model := r.instance.Statement.Model
	if model != nil {
		id := database.GetID(model)
		singleUpdate = id != nil
	}

	if c, ok := column.(string); ok && len(value) > 0 {
		r.instance.Statement.Dest = map[string]any{c: value[0]}
	}
	if len(value) == 0 {
		r.instance.Statement.Dest = column
	}

	if singleUpdate {
		if err := r.saving(model, r.instance.Statement.Dest); err != nil {
			return nil, err
		}
		if err := r.updating(model, r.instance.Statement.Dest); err != nil {
			return nil, err
		}
	}

	res, err := r.updates(r.instance.Statement.Dest)

	if singleUpdate && err == nil {
		if err := r.updated(model, r.instance.Statement.Dest); err != nil {
			return nil, err
		}
		if err := r.saved(model, r.instance.Statement.Dest); err != nil {
			return nil, err
		}
	}

	return res, err
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
	if len(r.instance.Statement.Selects) > 1 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				return errors.New("cannot set orm.Associations and other fields at the same time")
			}
		}
	}

	if len(r.instance.Statement.Selects) == 1 && r.instance.Statement.Selects[0] == orm.Associations {
		r.instance.Statement.Selects = []string{}
	}

	if err := r.saving(nil, value); err != nil {
		return err
	}
	if err := r.creating(value); err != nil {
		return err
	}

	if err := r.instance.Create(value).Error; err != nil {
		return err
	}

	if err := r.created(value); err != nil {
		return err
	}
	if err := r.saved(nil, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) omitCreate(value any) error {
	if len(r.instance.Statement.Omits) > 1 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				return errors.New("cannot set orm.Associations and other fields at the same time")
			}
		}
	}

	if len(r.instance.Statement.Omits) == 1 && r.instance.Statement.Omits[0] == orm.Associations {
		r.instance.Statement.Selects = []string{}
	}

	if err := r.saving(nil, value); err != nil {
		return err
	}
	if err := r.creating(value); err != nil {
		return err
	}

	if len(r.instance.Statement.Omits) == 1 && r.instance.Statement.Omits[0] == orm.Associations {
		if err := r.instance.Omit(orm.Associations).Create(value).Error; err != nil {
			return err
		}
	} else {
		if err := r.instance.Create(value).Error; err != nil {
			return err
		}
	}

	if err := r.created(value); err != nil {
		return err
	}
	if err := r.saved(nil, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) create(value any) error {
	if err := r.saving(nil, value); err != nil {
		return err
	}
	if err := r.creating(value); err != nil {
		return err
	}

	if err := r.instance.Omit(orm.Associations).Create(value).Error; err != nil {
		return err
	}

	if err := r.created(value); err != nil {
		return err
	}
	if err := r.saved(nil, value); err != nil {
		return err
	}

	return nil
}

func (r *Query) selectSave(value any) error {
	for _, val := range r.instance.Statement.Selects {
		if val == orm.Associations {
			return r.instance.Session(&gorm.Session{FullSaveAssociations: true}).Save(value).Error
		}
	}

	if err := r.instance.Save(value).Error; err != nil {
		return err
	}

	return nil
}

func (r *Query) omitSave(value any) error {
	for _, val := range r.instance.Statement.Omits {
		if val == orm.Associations {
			return r.instance.Omit(orm.Associations).Save(value).Error
		}
	}

	return r.instance.Save(value).Error
}

func (r *Query) save(value any) error {
	return r.instance.Omit(orm.Associations).Save(value).Error
}

func (r *Query) updates(values any) (*contractsorm.Result, error) {
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

func (r *Query) retrieved(dest any) error {
	return r.event(contractsorm.EventRetrieved, nil, dest)
}

func (r *Query) updating(model, dest any) error {
	return r.event(contractsorm.EventUpdating, model, dest)
}

func (r *Query) updated(model, dest any) error {
	return r.event(contractsorm.EventUpdated, model, dest)
}

func (r *Query) saving(model, dest any) error {
	return r.event(contractsorm.EventSaving, model, dest)
}

func (r *Query) saved(model, dest any) error {
	return r.event(contractsorm.EventSaved, model, dest)
}

func (r *Query) creating(dest any) error {
	return r.event(contractsorm.EventCreating, nil, dest)
}

func (r *Query) created(dest any) error {
	return r.event(contractsorm.EventCreated, nil, dest)
}

func (r *Query) deleting(dest any) error {
	return r.event(contractsorm.EventDeleting, nil, dest)
}

func (r *Query) deleted(dest any) error {
	return r.event(contractsorm.EventDeleted, nil, dest)
}

func (r *Query) forceDeleting(dest any) error {
	return r.event(contractsorm.EventForceDeleting, nil, dest)
}

func (r *Query) forceDeleted(dest any) error {
	return r.event(contractsorm.EventForceDeleted, nil, dest)
}

func (r *Query) event(event contractsorm.EventType, model, dest any) error {
	if r.withoutEvents {
		return nil
	}

	instance := NewEvent(r, model, dest)

	if dispatchesEvents, exist := dest.(contractsorm.DispatchesEvents); exist {
		if event, exist := dispatchesEvents.DispatchesEvents()[event]; exist {
			return event(instance)
		}

		return nil
	}
	if model != nil {
		if dispatchesEvents, exist := model.(contractsorm.DispatchesEvents); exist {
			if event, exist := dispatchesEvents.DispatchesEvents()[event]; exist {
				return event(instance)
			}
		}

		return nil
	}

	if observer := observer(dest); observer != nil {
		if observerEvent := observerEvent(event, observer); observerEvent != nil {
			return observerEvent(instance)
		}

		return nil
	}

	if model != nil {
		if observer := observer(model); observer != nil {
			if observerEvent := observerEvent(event, observer); observerEvent != nil {
				return observerEvent(instance)
			}

			return nil
		}
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

func observer(dest any) contractsorm.Observer {
	destType := reflect.TypeOf(dest)
	if destType.Kind() == reflect.Pointer {
		destType = destType.Elem()
	}

	for _, observer := range orm.Observers {
		modelType := reflect.TypeOf(observer.Model)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}
		if destType.Name() == modelType.Name() {
			return observer.Observer
		}
	}

	return nil
}

func observerEvent(event contractsorm.EventType, observer contractsorm.Observer) func(contractsorm.Event) error {
	switch event {
	case contractsorm.EventRetrieved:
		return observer.Retrieved
	case contractsorm.EventCreating:
		return observer.Creating
	case contractsorm.EventCreated:
		return observer.Created
	case contractsorm.EventUpdating:
		return observer.Updating
	case contractsorm.EventUpdated:
		return observer.Updated
	case contractsorm.EventSaving:
		return observer.Saving
	case contractsorm.EventSaved:
		return observer.Saved
	case contractsorm.EventDeleting:
		return observer.Deleting
	case contractsorm.EventDeleted:
		return observer.Deleted
	case contractsorm.EventForceDeleting:
		return observer.ForceDeleting
	case contractsorm.EventForceDeleted:
		return observer.ForceDeleted
	}

	return nil
}
