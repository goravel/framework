package gorm

import (
	"context"
	"database/sql"
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
	"github.com/goravel/framework/contracts/database/gorm"
	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm/hints"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/database"
)

var QuerySet = wire.NewSet(BuildQueryImpl, wire.Bind(new(ormcontract.Query), new(*QueryImpl)))
var _ ormcontract.Query = &QueryImpl{}

type QueryImpl struct {
	conditions Conditions
	config     config.Config
	connection string
	ctx        context.Context
	instance   *gormio.DB
	queries    map[string]*QueryImpl
}

func NewQueryImpl(ctx context.Context, config config.Config, connection string, db *gormio.DB, conditions *Conditions) *QueryImpl {
	queryImpl := &QueryImpl{
		config:     config,
		connection: connection,
		ctx:        ctx,
		instance:   db,
		queries:    make(map[string]*QueryImpl),
	}

	if conditions != nil {
		queryImpl.conditions = *conditions
	}

	return queryImpl
}

func BuildQueryImpl(ctx context.Context, config config.Config, connection string, gorm gorm.Gorm) (*QueryImpl, error) {
	db, err := gorm.Make()
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		db = db.WithContext(ctx)
	}

	return NewQueryImpl(ctx, config, connection, db, nil), nil
}

func (r *QueryImpl) Association(association string) ormcontract.Association {
	query := r.buildConditions()

	return query.instance.Association(association)
}

func (r *QueryImpl) Begin() (ormcontract.Transaction, error) {
	tx := r.instance.Begin()

	return NewTransaction(tx, r.config, r.connection), tx.Error
}

func (r *QueryImpl) Driver() ormcontract.Driver {
	return ormcontract.Driver(r.instance.Dialector.Name())
}

func (r *QueryImpl) Count(count *int64) error {
	query := r.buildConditions()

	return query.instance.Count(count).Error
}

func (r *QueryImpl) Create(value any) error {
	query, err := r.refreshConnection(value)
	if err != nil {
		return err
	}
	query = query.buildConditions()

	if len(query.instance.Statement.Selects) > 0 && len(query.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	if len(query.instance.Statement.Selects) > 0 {
		return query.selectCreate(value)
	}

	if len(query.instance.Statement.Omits) > 0 {
		return query.omitCreate(value)
	}

	return query.create(value)
}

func (r *QueryImpl) Cursor() (chan ormcontract.Cursor, error) {
	with := r.conditions.with
	query := r.buildConditions()
	r.conditions.with = with

	var err error
	cursorChan := make(chan ormcontract.Cursor)
	go func() {
		var rows *sql.Rows
		rows, err = query.instance.Rows()
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			val := make(map[string]any)
			err = query.instance.ScanRows(rows, val)
			if err != nil {
				return
			}
			cursorChan <- &CursorImpl{query: r, row: val}
		}
		close(cursorChan)
	}()
	return cursorChan, err
}

func (r *QueryImpl) Delete(dest any, conds ...any) (*ormcontract.Result, error) {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return nil, err
	}
	query = query.buildConditions()

	if err := query.deleting(dest); err != nil {
		return nil, err
	}

	res := query.instance.Delete(dest, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if err := query.deleted(dest); err != nil {
		return nil, err
	}

	return &ormcontract.Result{
		RowsAffected: res.RowsAffected,
	}, nil
}

func (r *QueryImpl) Distinct(args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.distinct = append(conditions.distinct, args...)

	return r.setConditions(conditions)
}

func (r *QueryImpl) Exec(sql string, values ...any) (*ormcontract.Result, error) {
	query := r.buildConditions()
	result := query.instance.Exec(sql, values...)

	return &ormcontract.Result{
		RowsAffected: result.RowsAffected,
	}, result.Error
}

func (r *QueryImpl) Exists(exists *bool) error {
	query := r.buildConditions()

	return query.instance.Select("1").Limit(1).Find(exists).Error
}

func (r *QueryImpl) Find(dest any, conds ...any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if err := filterFindConditions(conds...); err != nil {
		return err
	}
	if err := query.instance.Find(dest, conds...).Error; err != nil {
		return err
	}

	return query.retrieved(dest)
}

func (r *QueryImpl) FindOrFail(dest any, conds ...any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if err := filterFindConditions(conds...); err != nil {
		return err
	}

	res := query.instance.Find(dest, conds...)
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return orm.ErrRecordNotFound
	}

	return query.retrieved(dest)
}

func (r *QueryImpl) First(dest any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	res := query.instance.First(dest)
	if res.Error != nil {
		if errors.Is(res.Error, gormio.ErrRecordNotFound) {
			return nil
		}

		return res.Error
	}

	return query.retrieved(dest)
}

func (r *QueryImpl) FirstOr(dest any, callback func() error) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if err := query.instance.First(dest).Error; err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return callback()
		}

		return err
	}

	return query.retrieved(dest)
}

func (r *QueryImpl) FirstOrCreate(dest any, conds ...any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if len(conds) == 0 {
		return errors.New("query condition is require")
	}

	var res *gormio.DB
	if len(conds) > 1 {
		res = query.instance.Attrs(conds[1]).FirstOrInit(dest, conds[0])
	} else {
		res = query.instance.FirstOrInit(dest, conds[0])
	}

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return query.retrieved(dest)
	}

	return query.Create(dest)
}

func (r *QueryImpl) FirstOrFail(dest any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if err := query.instance.First(dest).Error; err != nil {
		if errors.Is(err, gormio.ErrRecordNotFound) {
			return orm.ErrRecordNotFound
		}

		return err
	}

	return query.retrieved(dest)
}

func (r *QueryImpl) FirstOrNew(dest any, attributes any, values ...any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	var res *gormio.DB
	if len(values) > 0 {
		res = query.instance.Attrs(values[0]).FirstOrInit(dest, attributes)
	} else {
		res = query.instance.FirstOrInit(dest, attributes)
	}

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return query.retrieved(dest)
	}

	return nil
}

func (r *QueryImpl) ForceDelete(value any, conds ...any) (*ormcontract.Result, error) {
	query, err := r.refreshConnection(value)
	if err != nil {
		return nil, err
	}

	query = query.buildConditions()

	if err := query.forceDeleting(value); err != nil {
		return nil, err
	}

	res := query.instance.Unscoped().Delete(value, conds...)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected > 0 {
		if err := query.forceDeleted(value); err != nil {
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
	conditions := r.conditions
	conditions.group = name

	return r.setConditions(conditions)
}

func (r *QueryImpl) Having(query any, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.having = &Having{
		query: query,
		args:  args,
	}

	return r.setConditions(conditions)
}

func (r *QueryImpl) Instance() *gormio.DB {
	return r.instance
}

func (r *QueryImpl) Join(query string, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.join = append(conditions.join, Join{
		query: query,
		args:  args,
	})

	return r.setConditions(conditions)
}

func (r *QueryImpl) Limit(limit int) ormcontract.Query {
	conditions := r.conditions
	conditions.limit = &limit

	return r.setConditions(conditions)
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
	conditions := r.conditions
	conditions.lockForUpdate = true

	return r.setConditions(conditions)
}

func (r *QueryImpl) Model(value any) ormcontract.Query {
	conditions := r.conditions
	conditions.model = value

	return r.setConditions(conditions)
}

func (r *QueryImpl) Offset(offset int) ormcontract.Query {
	conditions := r.conditions
	conditions.offset = &offset

	return r.setConditions(conditions)
}

func (r *QueryImpl) Omit(columns ...string) ormcontract.Query {
	conditions := r.conditions
	conditions.omit = columns

	return r.setConditions(conditions)
}

func (r *QueryImpl) Order(value any) ormcontract.Query {
	conditions := r.conditions
	conditions.order = append(r.conditions.order, value)

	return r.setConditions(conditions)
}

func (r *QueryImpl) OrderBy(column string, direction ...string) ormcontract.Query {
	var orderDirection string
	if len(direction) > 0 {
		orderDirection = direction[0]
	} else {
		orderDirection = "ASC"
	}
	return r.Order(fmt.Sprintf("%s %s", column, orderDirection))
}

func (r *QueryImpl) OrderByDesc(column string) ormcontract.Query {
	return r.Order(fmt.Sprintf("%s DESC", column))
}

func (r *QueryImpl) InRandomOrder() ormcontract.Query {
	order := ""
	switch r.Driver() {
	case ormcontract.DriverMysql:
		order = "RAND()"
	case ormcontract.DriverSqlserver:
		order = "NEWID()"
	case ormcontract.DriverPostgres:
		order = "RANDOM()"
	case ormcontract.DriverSqlite:
		order = "RANDOM()"
	}
	return r.Order(order)
}

func (r *QueryImpl) OrWhere(query any, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
		or:    true,
	})

	return r.setConditions(conditions)
}

func (r *QueryImpl) Paginate(page, limit int, dest any, total *int64) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	offset := (page - 1) * limit
	if total != nil {
		if query.conditions.table == nil && query.conditions.model == nil {
			if err := query.Model(dest).Count(total); err != nil {
				return err
			}
		} else {
			if err := query.Count(total); err != nil {
				return err
			}
		}
	}

	return query.Offset(offset).Limit(limit).Find(dest)
}

func (r *QueryImpl) Pluck(column string, dest any) error {
	query := r.buildConditions()

	return query.instance.Pluck(column, dest).Error
}

func (r *QueryImpl) Raw(sql string, values ...any) ormcontract.Query {
	return r.new(r.instance.Raw(sql, values...))
}

func (r *QueryImpl) Save(value any) error {
	query, err := r.refreshConnection(value)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	if len(query.instance.Statement.Selects) > 0 && len(query.instance.Statement.Omits) > 0 {
		return errors.New("cannot set Select and Omits at the same time")
	}

	model := query.instance.Statement.Model
	id := database.GetID(value)
	update := id != nil

	if err := query.saving(model, value); err != nil {
		return err
	}
	if update {
		if err := query.updating(model, value); err != nil {
			return err
		}
	} else {
		if err := query.creating(value); err != nil {
			return err
		}
	}

	if len(query.instance.Statement.Selects) > 0 {
		if err := query.selectSave(value); err != nil {
			return err
		}
	} else if len(query.instance.Statement.Omits) > 0 {
		if err := query.omitSave(value); err != nil {
			return err
		}
	} else {
		if err := query.save(value); err != nil {
			return err
		}
	}

	if update {
		if err := query.updated(model, value); err != nil {
			return err
		}
	} else {
		if err := query.created(value); err != nil {
			return err
		}
	}
	if err := query.saved(model, value); err != nil {
		return err
	}

	return nil
}

func (r *QueryImpl) SaveQuietly(value any) error {
	return r.WithoutEvents().Save(value)
}

func (r *QueryImpl) Scan(dest any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	return query.instance.Scan(dest).Error
}

func (r *QueryImpl) Scopes(funcs ...func(ormcontract.Query) ormcontract.Query) ormcontract.Query {
	conditions := r.conditions
	conditions.scopes = append(r.conditions.scopes, funcs...)

	return r.setConditions(conditions)
}

func (r *QueryImpl) Select(query any, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.selectColumns = &Select{
		query: query,
		args:  args,
	}

	return r.setConditions(conditions)
}

func (r *QueryImpl) SetContext(ctx context.Context) {
	r.ctx = ctx
	r.instance.Statement.Context = ctx
}

func (r *QueryImpl) SharedLock() ormcontract.Query {
	conditions := r.conditions
	conditions.sharedLock = true

	return r.setConditions(conditions)
}

func (r *QueryImpl) Sum(column string, dest any) error {
	query := r.buildConditions()

	return query.instance.Select("SUM(" + column + ")").Row().Scan(dest)
}

func (r *QueryImpl) Table(name string, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.table = &Table{
		name: name,
		args: args,
	}

	return r.setConditions(conditions)
}

func (r *QueryImpl) ToSql() ormcontract.ToSql {
	return NewToSql(r.setConditions(r.conditions), false)
}

func (r *QueryImpl) ToRawSql() ormcontract.ToSql {
	return NewToSql(r.setConditions(r.conditions), true)
}

func (r *QueryImpl) Update(column any, value ...any) (*ormcontract.Result, error) {
	query := r.buildConditions()

	if _, ok := column.(string); !ok && len(value) > 0 {
		return nil, errors.New("parameter error, please check the document")
	}

	var singleUpdate bool
	model := query.instance.Statement.Model
	if model != nil {
		id := database.GetID(model)
		singleUpdate = id != nil
	}

	if c, ok := column.(string); ok && len(value) > 0 {
		query.instance.Statement.Dest = map[string]any{c: value[0]}
	}
	if len(value) == 0 {
		query.instance.Statement.Dest = column
	}

	if singleUpdate {
		if err := query.saving(model, query.instance.Statement.Dest); err != nil {
			return nil, err
		}
		if err := query.updating(model, query.instance.Statement.Dest); err != nil {
			return nil, err
		}
	}

	res, err := query.updates(query.instance.Statement.Dest)

	if singleUpdate && err == nil {
		if err := query.updated(model, query.instance.Statement.Dest); err != nil {
			return nil, err
		}
		if err := query.saved(model, query.instance.Statement.Dest); err != nil {
			return nil, err
		}
	}

	return res, err
}

func (r *QueryImpl) UpdateOrCreate(dest any, attributes any, values any) error {
	query, err := r.refreshConnection(dest)
	if err != nil {
		return err
	}

	query = query.buildConditions()

	res := query.instance.Assign(values).FirstOrInit(dest, attributes)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return query.Save(dest)
	}

	return query.Create(dest)
}

func (r *QueryImpl) Where(query any, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.where = append(r.conditions.where, Where{
		query: query,
		args:  args,
	})

	return r.setConditions(conditions)
}

func (r *QueryImpl) WhereIn(column string, values []any) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s IN ?", column), values)
}

func (r *QueryImpl) OrWhereIn(column string, values []any) ormcontract.Query {
	return r.OrWhere(fmt.Sprintf("%s IN ?", column), values)
}

func (r *QueryImpl) WhereNotIn(column string, values []any) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s NOT IN ?", column), values)
}

func (r *QueryImpl) OrWhereNotIn(column string, values []any) ormcontract.Query {
	return r.OrWhere(fmt.Sprintf("%s NOT IN ?", column), values)
}

func (r *QueryImpl) WhereBetween(column string, x, y any) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s BETWEEN %v AND %v", column, x, y))
}

func (r *QueryImpl) WhereNotBetween(column string, x, y any) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s NOT BETWEEN %v AND %v", column, x, y))
}

func (r *QueryImpl) OrWhereBetween(column string, x, y any) ormcontract.Query {
	return r.OrWhere(fmt.Sprintf("%s BETWEEN %v AND %v", column, x, y))
}

func (r *QueryImpl) OrWhereNotBetween(column string, x, y any) ormcontract.Query {
	return r.OrWhere(fmt.Sprintf("%s NOT BETWEEN %v AND %v", column, x, y))
}

func (r *QueryImpl) OrWhereNull(column string) ormcontract.Query {
	return r.OrWhere(fmt.Sprintf("%s IS NULL", column))
}

func (r *QueryImpl) WhereNull(column string) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s IS NULL", column))
}

func (r *QueryImpl) WhereNotNull(column string) ormcontract.Query {
	return r.Where(fmt.Sprintf("%s IS NOT NULL", column))
}

func (r *QueryImpl) With(query string, args ...any) ormcontract.Query {
	conditions := r.conditions
	conditions.with = append(r.conditions.with, With{
		query: query,
		args:  args,
	})

	return r.setConditions(conditions)
}

func (r *QueryImpl) WithoutEvents() ormcontract.Query {
	conditions := r.conditions
	conditions.withoutEvents = true

	return r.setConditions(conditions)
}

func (r *QueryImpl) WithTrashed() ormcontract.Query {
	conditions := r.conditions
	conditions.withTrashed = true

	return r.setConditions(conditions)
}

func (r *QueryImpl) buildConditions() *QueryImpl {
	query := r.buildModel()
	db := query.instance
	db = query.buildDistinct(db)
	db = query.buildGroup(db)
	db = query.buildHaving(db)
	db = query.buildJoin(db)
	db = query.buildLockForUpdate(db)
	db = query.buildLimit(db)
	db = query.buildOrder(db)
	db = query.buildOffset(db)
	db = query.buildOmit(db)
	db = query.buildScopes(db)
	db = query.buildSelectColumns(db)
	db = query.buildSharedLock(db)
	db = query.buildTable(db)
	db = query.buildWith(db)
	db = query.buildWithTrashed(db)
	db = query.buildWhere(db)

	return query.new(db)
}

func (r *QueryImpl) buildDistinct(db *gormio.DB) *gormio.DB {
	if len(r.conditions.distinct) == 0 {
		return db
	}

	db = db.Distinct(r.conditions.distinct...)
	r.conditions.distinct = nil

	return db
}

func (r *QueryImpl) buildGroup(db *gormio.DB) *gormio.DB {
	if r.conditions.group == "" {
		return db
	}

	db = db.Group(r.conditions.group)
	r.conditions.group = ""

	return db
}

func (r *QueryImpl) buildHaving(db *gormio.DB) *gormio.DB {
	if r.conditions.having == nil {
		return db
	}

	db = db.Having(r.conditions.having.query, r.conditions.having.args...)
	r.conditions.having = nil

	return db
}

func (r *QueryImpl) buildJoin(db *gormio.DB) *gormio.DB {
	if r.conditions.join == nil {
		return db
	}

	for _, item := range r.conditions.join {
		db = db.Joins(item.query, item.args...)
	}

	r.conditions.join = nil

	return db
}

func (r *QueryImpl) buildLimit(db *gormio.DB) *gormio.DB {
	if r.conditions.limit == nil {
		return db
	}

	db = db.Limit(*r.conditions.limit)
	r.conditions.limit = nil

	return db
}

func (r *QueryImpl) buildLockForUpdate(db *gormio.DB) *gormio.DB {
	if !r.conditions.lockForUpdate {
		return db
	}

	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		return db.Clauses(clause.Locking{Strength: "UPDATE"})
	} else if driver == sqlserverDialector.Name() {
		return db.Clauses(hints.With("rowlock", "updlock", "holdlock"))
	}

	r.conditions.lockForUpdate = false

	return db
}

func (r *QueryImpl) buildModel() *QueryImpl {
	if r.conditions.model == nil {
		return r
	}

	query, err := r.refreshConnection(r.conditions.model)
	if err != nil {
		return nil
	}

	return query.new(query.instance.Model(r.conditions.model))
}

func (r *QueryImpl) buildOffset(db *gormio.DB) *gormio.DB {
	if r.conditions.offset == nil {
		return db
	}

	db = db.Offset(*r.conditions.offset)
	r.conditions.offset = nil

	return db
}

func (r *QueryImpl) buildOmit(db *gormio.DB) *gormio.DB {
	if len(r.conditions.omit) == 0 {
		return db
	}

	db = db.Omit(r.conditions.omit...)
	r.conditions.omit = nil

	return db
}

func (r *QueryImpl) buildOrder(db *gormio.DB) *gormio.DB {
	if len(r.conditions.order) == 0 {
		return db
	}

	for _, order := range r.conditions.order {
		db = db.Order(order)
	}

	r.conditions.order = nil

	return db
}

func (r *QueryImpl) buildSelectColumns(db *gormio.DB) *gormio.DB {
	if r.conditions.selectColumns == nil {
		return db
	}

	db = db.Select(r.conditions.selectColumns.query, r.conditions.selectColumns.args...)
	r.conditions.selectColumns = nil

	return db
}

func (r *QueryImpl) buildScopes(db *gormio.DB) *gormio.DB {
	if len(r.conditions.scopes) == 0 {
		return db
	}

	var gormFuncs []func(*gormio.DB) *gormio.DB
	for _, scope := range r.conditions.scopes {
		currentScope := scope
		gormFuncs = append(gormFuncs, func(tx *gormio.DB) *gormio.DB {
			queryImpl := r.new(tx)
			query := currentScope(queryImpl)
			queryImpl = query.(*QueryImpl)
			queryImpl = queryImpl.buildConditions()

			return queryImpl.instance
		})
	}

	db = db.Scopes(gormFuncs...)
	r.conditions.scopes = nil

	return db
}

func (r *QueryImpl) buildSharedLock(db *gormio.DB) *gormio.DB {
	if !r.conditions.sharedLock {
		return db
	}

	driver := r.instance.Name()
	mysqlDialector := mysql.Dialector{}
	postgresqlDialector := postgres.Dialector{}
	sqlserverDialector := sqlserver.Dialector{}

	if driver == mysqlDialector.Name() || driver == postgresqlDialector.Name() {
		return db.Clauses(clause.Locking{Strength: "SHARE"})
	} else if driver == sqlserverDialector.Name() {
		return db.Clauses(hints.With("rowlock", "holdlock"))
	}

	r.conditions.sharedLock = false

	return db
}

func (r *QueryImpl) buildTable(db *gormio.DB) *gormio.DB {
	if r.conditions.table == nil {
		return db
	}

	db = db.Table(r.conditions.table.name, r.conditions.table.args...)
	r.conditions.table = nil

	return db
}

func (r *QueryImpl) buildWhere(db *gormio.DB) *gormio.DB {
	if len(r.conditions.where) == 0 {
		return db
	}

	for _, item := range r.conditions.where {
		if item.or {
			db = db.Or(item.query, item.args...)
		} else {
			db = db.Where(item.query, item.args...)
		}
	}

	r.conditions.where = nil

	return db
}

func (r *QueryImpl) buildWith(db *gormio.DB) *gormio.DB {
	if len(r.conditions.with) == 0 {
		return db
	}

	for _, item := range r.conditions.with {
		isSet := false
		if len(item.args) == 1 {
			if arg, ok := item.args[0].(func(ormcontract.Query) ormcontract.Query); ok {
				newArgs := []any{
					func(tx *gormio.DB) *gormio.DB {
						queryImpl := NewQueryImpl(r.ctx, r.config, r.connection, tx, nil)
						query := arg(queryImpl)
						queryImpl = query.(*QueryImpl)
						queryImpl = queryImpl.buildConditions()

						return queryImpl.instance
					},
				}

				db = db.Preload(item.query, newArgs...)
				isSet = true
			}
		}

		if !isSet {
			db = db.Preload(item.query, item.args...)
		}
	}

	r.conditions.with = nil

	return db
}

func (r *QueryImpl) buildWithTrashed(db *gormio.DB) *gormio.DB {
	if !r.conditions.withTrashed {
		return db
	}

	db = db.Unscoped()
	r.conditions.withTrashed = false

	return db
}

func (r *QueryImpl) clearConditions() {
	r.conditions = Conditions{}
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

func (r *QueryImpl) created(dest any) error {
	return r.event(ormcontract.EventCreated, nil, dest)
}

func (r *QueryImpl) creating(dest any) error {
	return r.event(ormcontract.EventCreating, nil, dest)
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
	if r.conditions.withoutEvents {
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

			return nil
		}
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

func (r *QueryImpl) new(db *gormio.DB) *QueryImpl {
	query := NewQueryImpl(r.ctx, r.config, r.connection, db, &r.conditions)

	return query
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

func (r *QueryImpl) omitSave(value any) error {
	for _, val := range r.instance.Statement.Omits {
		if val == orm.Associations {
			return r.instance.Omit(orm.Associations).Save(value).Error
		}
	}

	return r.instance.Save(value).Error
}

func (r *QueryImpl) refreshConnection(model any) (*QueryImpl, error) {
	connection, err := getModelConnection(model)
	if err != nil {
		return nil, err
	}
	if connection == "" || connection == r.connection {
		return r, nil
	}

	query, ok := r.queries[connection]
	if !ok {
		var err error
		query, err = InitializeQuery(r.ctx, r.config, connection)
		if err != nil {
			return nil, err
		}

		if r.queries == nil {
			r.queries = make(map[string]*QueryImpl)
		}
		r.queries[connection] = query
	}

	query.conditions = r.conditions

	return query, nil
}

func (r *QueryImpl) retrieved(dest any) error {
	return r.event(ormcontract.EventRetrieved, nil, dest)
}

func (r *QueryImpl) save(value any) error {
	return r.instance.Omit(orm.Associations).Save(value).Error
}

func (r *QueryImpl) saved(model, dest any) error {
	return r.event(ormcontract.EventSaved, model, dest)
}

func (r *QueryImpl) saving(model, dest any) error {
	return r.event(ormcontract.EventSaving, model, dest)
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

func (r *QueryImpl) setConditions(conditions Conditions) *QueryImpl {
	query := r.new(r.instance)
	query.conditions = conditions

	return query
}

func (r *QueryImpl) updating(model, dest any) error {
	return r.event(ormcontract.EventUpdating, model, dest)
}

func (r *QueryImpl) updated(model, dest any) error {
	return r.event(ormcontract.EventUpdated, model, dest)
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

func getModelConnection(model any) (string, error) {
	value1 := reflect.ValueOf(model)
	if value1.Kind() == reflect.Ptr && value1.IsNil() {
		value1 = reflect.New(value1.Type().Elem())
	}
	modelType := reflect.Indirect(value1).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(model)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return "", errors.New("invalid model")
		}
		return "", fmt.Errorf("%s: %s.%s", "invalid model", modelType.PkgPath(), modelType.Name())
	}

	modelValue := reflect.New(modelType)
	connectionModel, ok := modelValue.Interface().(ormcontract.ConnectionModel)
	if !ok {
		return "", nil
	}

	return connectionModel.Connection(), nil
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
