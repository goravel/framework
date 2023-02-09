package orm

import (
	"context"
	"database/sql"
)

//go:generate mockery --name=Orm
type Orm interface {
	Connection(name string) Orm
	DB() (*sql.DB, error)
	Query() DB
	Transaction(txFunc func(tx Transaction) error) error
	WithContext(ctx context.Context) Orm
}

//go:generate mockery --name=DB
type DB interface {
	Query
	Begin() (Transaction, error)
}

//go:generate mockery --name=Transaction
type Transaction interface {
	Query
	Commit() error
	Rollback() error
}

type Query interface {
	Association(association string) Association
	Driver() Driver
	Count(count *int64) error
	Create(value any) error
	Delete(value any, conds ...any) error
	Distinct(args ...any) Query
	Exec(sql string, values ...any) error
	Find(dest any, conds ...any) error
	First(dest any) error
	FirstOrCreate(dest any, conds ...any) error
	ForceDelete(value any, conds ...any) error
	Get(dest any) error
	Group(name string) Query
	Having(query any, args ...any) Query
	Join(query string, args ...any) Query
	Limit(limit int) Query
	Load(dest any, relation string, args ...any) error
	LoadMissing(dest any, relation string, args ...any) error
	Model(value any) Query
	Offset(offset int) Query
	Omit(columns ...string) Query
	Order(value any) Query
	OrWhere(query any, args ...any) Query
	Paginate(page, limit int, dest any, total *int64) error
	Pluck(column string, dest any) error
	Raw(sql string, values ...any) Query
	Save(value any) error
	Scan(dest any) error
	Scopes(funcs ...func(Query) Query) Query
	Select(query any, args ...any) Query
	Table(name string, args ...any) Query
	Update(column string, value any) error
	Updates(values any) error
	Where(query any, args ...any) Query
	WithTrashed() Query
	With(query string, args ...any) Query
}

//go:generate mockery --name=Association
type Association interface {
	Find(out any, conds ...any) error
	Append(values ...any) error
	Replace(values ...any) error
	Delete(values ...any) error
	Clear() error
	Count() int64
}
