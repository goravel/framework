package gorm

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/google/wire"
	"github.com/spf13/cast"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/goravel/framework/contracts/config"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm/hints"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/database"
)

var QuerySet = wire.NewSet(NewQueryImpl, wire.Bind(new(ormcontract.Query), new(*QueryImpl)))
var _ ormcontract.Query = &QueryImpl{}

type QueryImpl struct {
	config        config.Config
	ctx           context.Context
	instance      *gormio.DB
	origin        *QueryImpl
	with          map[string][]any
	withoutEvents bool
}

func NewQueryImpl(ctx context.Context, config config.Config, gorm Gorm) (*QueryImpl, error) {
	db, err := gorm.Make()
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return &QueryImpl{
		instance: db,
		config:   config,
		ctx:      ctx,
	}, nil
}

func NewQueryImplByInstance(db *gormio.DB, instance *QueryImpl) *QueryImpl {
	queryImpl := &QueryImpl{config: instance.config, ctx: db.Statement.Context, instance: db, origin: instance.origin, with: instance.with, withoutEvents: instance.withoutEvents}

	// The origin is used by the With method to load the relationship.
	if instance.origin == nil && instance.instance != nil {
		queryImpl.origin = instance
	}

	return queryImpl
}

func (r *QueryImpl) Association(association string) ormcontract.Association {
	return r.instance.Association(association)
}

func (r *QueryImpl) Begin() (ormcontract.Transaction, error) {
	tx := r.instance.Begin()

	return NewTransaction(tx, r.config), tx.Error
}

func (r *QueryImpl) Driver() ormcontract.Driver {
	return ormcontract.Driver(r.instance.Dialector.Name())
}

func (r *QueryImpl) Count(count *int64) error {
	return r.instance.Count(count).Error
}

func (r *QueryImpl) Create(value any) error {
	if err := r.refreshConnection(value); err != nil {
		return err
	}
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

func (r *QueryImpl) Cursor() (chan ormcontract.Cursor, error) {
	var err error
	cursorChan := make(chan ormcontract.Cursor)
	go func() {
		rows, err := r.instance.Rows()
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			val := make(map[string]any)
			err := r.instance.ScanRows(rows, val)
			if err != nil {
				return
			}
			cursorChan <- &CursorImpl{row: val, query: r}
		}
		close(cursorChan)
	}()
	return cursorChan, err
}

func (r *QueryImpl) Delete(dest any, conds ...any) (*ormcontract.Result, error) {
	if err := r.refreshConnection(dest); err != nil {
		return nil, err
	}
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

	return &ormcontract.Result{
		RowsAffected: res.RowsAffected,
	}, nil
}

func (r *QueryImpl) Distinct(args ...any) ormcontract.Query {
	tx := r.instance.Distinct(args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Exec(sql string, values ...any) (*ormcontract.Result, error) {
	result := r.instance.Exec(sql, values...)

	return &ormcontract.Result{
		RowsAffected: result.RowsAffected,
	}, result.Error
}

func (r *QueryImpl) Find(dest any, conds ...any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	if err := filterFindConditions(conds...); err != nil {
		return err
	}
	if err := r.instance.Find(dest, conds...).Error; err != nil {
		return err
	}

	return r.retrieved(dest)
}

func (r *QueryImpl) FindOrFail(dest any, conds ...any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
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

func (r *QueryImpl) First(dest any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	res := r.instance.First(dest)
	if res.Error != nil {
		if errors.Is(res.Error, gormio.ErrRecordNotFound) {
			return nil
		}

		return res.Error
	}

	return r.retrieved(dest)
}

func (r *QueryImpl) FirstOr(dest any, callback func() error) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	err := r.instance.First(dest).Error
	if err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return callback()
		}

		return err
	}

	return r.retrieved(dest)
}

func (r *QueryImpl) FirstOrCreate(dest any, conds ...any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	if len(conds) == 0 {
		return errors.New("query condition is require")
	}

	var res *gormio.DB
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

func (r *QueryImpl) FirstOrFail(dest any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	err := r.instance.First(dest).Error
	if err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return orm.ErrRecordNotFound
		}

		return err
	}

	return r.retrieved(dest)
}

func (r *QueryImpl) FirstOrNew(dest any, attributes any, values ...any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	var res *gormio.DB
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

func (r *QueryImpl) ForceDelete(value any, conds ...any) (*ormcontract.Result, error) {
	if err := r.refreshConnection(value); err != nil {
		return nil, err
	}
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

	return &ormcontract.Result{
		RowsAffected: res.RowsAffected,
	}, res.Error
}

func (r *QueryImpl) Get(dest any) error {
	return r.Find(dest)
}

func (r *QueryImpl) Group(name string) ormcontract.Query {
	tx := r.instance.Group(name)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Having(query any, args ...any) ormcontract.Query {
	tx := r.instance.Having(query, args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Instance() *gormio.DB {
	return r.instance
}

func (r *QueryImpl) Join(query string, args ...any) ormcontract.Query {
	tx := r.instance.Joins(query, args...)
	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Limit(limit int) ormcontract.Query {
	tx := r.instance.Limit(limit)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Load(model any, relation string, args ...any) error {
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
	err := r.With(relation, args...).Find(model)

	t := destType.Elem()
	v := reflect.ValueOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}
		if t.Field(i).Name != relation {
			v.Field(i).Set(copyDest.Field(i))
		}
	}

	return err
}

func (r *QueryImpl) LoadMissing(model any, relation string, args ...any) error {
	destType := reflect.TypeOf(model)
	if destType.Kind() != reflect.Pointer {
		return errors.New("model must be pointer")
	}

	t := reflect.TypeOf(model).Elem()
	v := reflect.ValueOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}
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

func (r *QueryImpl) LockForUpdate() ormcontract.Query {
	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		tx := r.instance.Clauses(clause.Locking{Strength: "UPDATE"})

		return NewQueryImplByInstance(tx, r)
	} else if driver == sqlserverDialector.Name() {
		tx := r.instance.Clauses(hints.With("rowlock", "updlock", "holdlock"))

		return NewQueryImplByInstance(tx, r)
	}

	return r
}

func (r *QueryImpl) Model(value any) ormcontract.Query {
	if err := r.refreshConnection(value); err != nil {
		return nil
	}
	tx := r.instance.Model(value)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Offset(offset int) ormcontract.Query {
	tx := r.instance.Offset(offset)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Omit(columns ...string) ormcontract.Query {
	tx := r.instance.Omit(columns...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Order(value any) ormcontract.Query {
	tx := r.instance.Order(value)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) OrWhere(query any, args ...any) ormcontract.Query {
	tx := r.instance.Or(query, args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Paginate(page, limit int, dest any, total *int64) error {
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

func (r *QueryImpl) Pluck(column string, dest any) error {
	return r.instance.Pluck(column, dest).Error
}

func (r *QueryImpl) Raw(sql string, values ...any) ormcontract.Query {
	tx := r.instance.Raw(sql, values...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Save(value any) error {
	if err := r.refreshConnection(value); err != nil {
		return err
	}
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

func (r *QueryImpl) SaveQuietly(value any) error {
	return r.WithoutEvents().Save(value)
}

func (r *QueryImpl) Scan(dest any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}

	return r.instance.Scan(dest).Error
}

func (r *QueryImpl) Select(query any, args ...any) ormcontract.Query {
	tx := r.instance.Select(query, args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Scopes(funcs ...func(ormcontract.Query) ormcontract.Query) ormcontract.Query {
	var gormFuncs []func(*gormio.DB) *gormio.DB
	for _, item := range funcs {
		gormFuncs = append(gormFuncs, func(tx *gormio.DB) *gormio.DB {
			item(NewQueryImplByInstance(tx, r))

			return tx
		})
	}

	tx := r.instance.Scopes(gormFuncs...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) SharedLock() ormcontract.Query {
	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		tx := r.instance.Clauses(clause.Locking{Strength: "SHARE"})

		return NewQueryImplByInstance(tx, r)
	} else if driver == sqlserverDialector.Name() {
		tx := r.instance.Clauses(hints.With("rowlock", "holdlock"))

		return NewQueryImplByInstance(tx, r)
	}

	return r
}

func (r *QueryImpl) Sum(column string, dest any) error {
	return r.instance.Select("SUM(" + column + ")").Row().Scan(dest)
}

func (r *QueryImpl) Table(name string, args ...any) ormcontract.Query {
	tx := r.instance.Table(name, args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) Update(column any, value ...any) (*ormcontract.Result, error) {
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

func (r *QueryImpl) UpdateOrCreate(dest any, attributes any, values any) error {
	if err := r.refreshConnection(dest); err != nil {
		return err
	}
	res := r.instance.Assign(values).FirstOrInit(dest, attributes)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return r.Save(dest)
	}

	return r.Create(dest)
}

func (r *QueryImpl) Where(query any, args ...any) ormcontract.Query {
	tx := r.instance.Where(query, args...)

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) WithoutEvents() ormcontract.Query {
	return NewQueryImplByInstance(r.instance, &QueryImpl{
		config:        r.config,
		withoutEvents: true,
	})
}

func (r *QueryImpl) WithTrashed() ormcontract.Query {
	tx := r.instance.Unscoped()

	return NewQueryImplByInstance(tx, r)
}

func (r *QueryImpl) With(query string, args ...any) ormcontract.Query {
	if len(args) == 1 {
		switch arg := args[0].(type) {
		case func(ormcontract.Query) ormcontract.Query:
			newArgs := []any{
				func(tx *gormio.DB) *gormio.DB {
					query := arg(NewQueryImplByInstance(tx, r))

					return query.(*QueryImpl).instance
				},
			}

			tx := r.instance.Preload(query, newArgs...)

			return NewQueryImplByInstance(tx, r)
		}
	}

	tx := r.instance.Preload(query, args...)

	queryImpl := NewQueryImplByInstance(tx, r)
	if queryImpl.with == nil {
		queryImpl.with = make(map[string][]any)
	}

	queryImpl.with[query] = args

	return queryImpl
}

func (r *QueryImpl) refreshConnection(value any) error {
	model, ok := value.(ormcontract.ConnectionModel)
	if !ok {
		return nil
	}
	conn := model.Connection()
	if conn == "" {
		conn = r.config.GetString("database.default")
	}
	driver := driver2gorm(r.config.GetString(fmt.Sprintf("database.connections.%s.driver", conn)))
	if driver == "" {
		return fmt.Errorf("connection %s driver is not supported", conn)
	}
	// if a driver is not the same, we need to refresh the connection
	if driver != r.instance.Name() {
		query, err := InitializeQuery(r.ctx, r.config, conn)
		if err != nil {
			return err
		}
		dbInstance := query.instance
		stmt := r.instance.Statement
		stmt.DB = dbInstance.Statement.DB
		stmt.ConnPool = dbInstance.ConnPool
		if r.ctx != nil {
			dbInstance = dbInstance.WithContext(r.ctx)
		}
		dbInstance.Statement = stmt
		r.instance = dbInstance
	}
	return nil
}

func (r *QueryImpl) selectCreate(value any) error {
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

func (r *QueryImpl) omitCreate(value any) error {
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

func (r *QueryImpl) create(value any) error {
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

func (r *QueryImpl) selectSave(value any) error {
	for _, val := range r.instance.Statement.Selects {
		if val == orm.Associations {
			return r.instance.Session(&gormio.Session{FullSaveAssociations: true}).Save(value).Error
		}
	}

	if err := r.instance.Save(value).Error; err != nil {
		return err
	}

	return nil
}

func (r *QueryImpl) omitSave(value any) error {
	for _, val := range r.instance.Statement.Omits {
		if val == orm.Associations {
			return r.instance.Omit(orm.Associations).Save(value).Error
		}
	}

	return r.instance.Save(value).Error
}

func (r *QueryImpl) save(value any) error {
	return r.instance.Omit(orm.Associations).Save(value).Error
}

func (r *QueryImpl) updates(values any) (*ormcontract.Result, error) {
	if len(r.instance.Statement.Selects) > 0 && len(r.instance.Statement.Omits) > 0 {
		return nil, errors.New("cannot set Select and Omits at the same time")
	}

	if len(r.instance.Statement.Selects) > 0 {
		for _, val := range r.instance.Statement.Selects {
			if val == orm.Associations {
				result := r.instance.Session(&gormio.Session{FullSaveAssociations: true}).Updates(values)
				return &ormcontract.Result{
					RowsAffected: result.RowsAffected,
				}, result.Error
			}
		}

		result := r.instance.Updates(values)

		return &ormcontract.Result{
			RowsAffected: result.RowsAffected,
		}, result.Error
	}

	if len(r.instance.Statement.Omits) > 0 {
		for _, val := range r.instance.Statement.Omits {
			if val == orm.Associations {
				result := r.instance.Omit(orm.Associations).Updates(values)

				return &ormcontract.Result{
					RowsAffected: result.RowsAffected,
				}, result.Error
			}
		}
		result := r.instance.Updates(values)

		return &ormcontract.Result{
			RowsAffected: result.RowsAffected,
		}, result.Error
	}
	result := r.instance.Omit(orm.Associations).Updates(values)

	return &ormcontract.Result{
		RowsAffected: result.RowsAffected,
	}, result.Error
}

func (r *QueryImpl) retrieved(dest any) error {
	return r.event(ormcontract.EventRetrieved, nil, dest)
}

func (r *QueryImpl) updating(model, dest any) error {
	return r.event(ormcontract.EventUpdating, model, dest)
}

func (r *QueryImpl) updated(model, dest any) error {
	return r.event(ormcontract.EventUpdated, model, dest)
}

func (r *QueryImpl) saving(model, dest any) error {
	return r.event(ormcontract.EventSaving, model, dest)
}

func (r *QueryImpl) saved(model, dest any) error {
	return r.event(ormcontract.EventSaved, model, dest)
}

func (r *QueryImpl) creating(dest any) error {
	return r.event(ormcontract.EventCreating, nil, dest)
}

func (r *QueryImpl) created(dest any) error {
	return r.event(ormcontract.EventCreated, nil, dest)
}

func (r *QueryImpl) deleting(dest any) error {
	return r.event(ormcontract.EventDeleting, nil, dest)
}

func (r *QueryImpl) deleted(dest any) error {
	return r.event(ormcontract.EventDeleted, nil, dest)
}

func (r *QueryImpl) forceDeleting(dest any) error {
	return r.event(ormcontract.EventForceDeleting, nil, dest)
}

func (r *QueryImpl) forceDeleted(dest any) error {
	return r.event(ormcontract.EventForceDeleted, nil, dest)
}

func (r *QueryImpl) event(event ormcontract.EventType, model, dest any) error {
	if r.withoutEvents {
		return nil
	}

	instance := NewEvent(r, model, dest)

	if dispatchesEvents, exist := dest.(ormcontract.DispatchesEvents); exist {
		if event, exist := dispatchesEvents.DispatchesEvents()[event]; exist {
			return event(instance)
		}

		return nil
	}
	if model != nil {
		if dispatchesEvents, exist := model.(ormcontract.DispatchesEvents); exist {
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

func observer(dest any) ormcontract.Observer {
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

func observerEvent(event ormcontract.EventType, observer ormcontract.Observer) func(ormcontract.Event) error {
	switch event {
	case ormcontract.EventRetrieved:
		return observer.Retrieved
	case ormcontract.EventCreating:
		return observer.Creating
	case ormcontract.EventCreated:
		return observer.Created
	case ormcontract.EventUpdating:
		return observer.Updating
	case ormcontract.EventUpdated:
		return observer.Updated
	case ormcontract.EventSaving:
		return observer.Saving
	case ormcontract.EventSaved:
		return observer.Saved
	case ormcontract.EventDeleting:
		return observer.Deleting
	case ormcontract.EventDeleted:
		return observer.Deleted
	case ormcontract.EventForceDeleting:
		return observer.ForceDeleting
	case ormcontract.EventForceDeleted:
		return observer.ForceDeleted
	}

	return nil
}

func driver2gorm(driver string) string {
	switch driver {
	case "mysql":
		return "mysql"
	case "postgresql":
		return "postgres"
	case "sqlite":
		return "sqlite"
	case "sqlserver":
		return "sqlserver"
	default:
		return ""
	}
}
