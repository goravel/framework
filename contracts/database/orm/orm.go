package orm

import "context"

//go:generate mockery --name=Orm
type Orm interface {
	Connection(name string) Orm
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
	Driver() Driver
	Count(count *int64) error
	Create(value interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	Distinct(args ...interface{}) Query
	Exec(sql string, values ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	First(dest interface{}) error
	FirstOrCreate(dest interface{}, conds ...interface{}) error
	ForceDelete(value interface{}, conds ...interface{}) error
	Get(dest interface{}) error
	Group(name string) Query
	Having(query interface{}, args ...interface{}) Query
	Join(query string, args ...interface{}) Query
	Limit(limit int) Query
	Model(value interface{}) Query
	Offset(offset int) Query
	Order(value interface{}) Query
	OrWhere(query interface{}, args ...interface{}) Query
	Pluck(column string, dest interface{}) error
	Raw(sql string, values ...interface{}) Query
	Save(value interface{}) error
	Scan(dest interface{}) error
	Scopes(funcs ...func(Query) Query) Query
	Select(query interface{}, args ...interface{}) Query
	Table(name string, args ...interface{}) Query
	Update(column string, value interface{}) error
	Updates(values interface{}) error
	Where(query interface{}, args ...interface{}) Query
	WithTrashed() Query
}
