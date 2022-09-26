package orm

import (
	"context"
)

//go:generate mockery --name=Orm
type Orm interface {
	WithContext(ctx context.Context) Orm
	Connection(name string) Orm
	Query() DB
	Transaction(txFunc func(tx Transaction) error) error
}

//go:generate mockery --name=DB
type DB interface {
	methods
	Begin() (Transaction, error)
}

//go:generate mockery --name=Transaction
type Transaction interface {
	methods
	Commit() error
	Rollback() error
}

type methods interface {
	Model(value interface{}) Transaction
	Table(name string, args ...interface{}) Transaction
	Select(query interface{}, args ...interface{}) Transaction
	Where(query interface{}, args ...interface{}) Transaction
	Join(query string, args ...interface{}) Transaction
	Group(name string) Transaction
	Having(query interface{}, args ...interface{}) Transaction
	Order(value interface{}) Transaction
	Limit(limit int) Transaction
	Offset(offset int) Transaction
	Scopes(funcs ...func(Transaction) Transaction) Transaction
	Raw(sql string, values ...interface{}) Transaction
	WithTrashed() Transaction
	Create(value interface{}) error
	Save(value interface{}) error
	First(dest interface{}, conds ...interface{}) error
	Last(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	FirstOrCreate(dest interface{}, conds ...interface{}) error
	Update(column string, value interface{}) error
	Updates(values interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	ForceDelete(value interface{}, conds ...interface{}) error
	Count(count *int64) error
	Pluck(column string, dest interface{}) error
	Scan(dest interface{}) error
	Exec(sql string, values ...interface{}) error
}
