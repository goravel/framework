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
	Model(value any) Query
	Offset(offset int) Query
	Omit(columns ...string) Query
	Order(value any) Query
	OrWhere(query any, args ...any) Query
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
	Load(dest any, relation string, relations ...string) error
}
