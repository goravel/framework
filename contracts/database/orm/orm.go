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
	Methods
	Begin() (Transaction, error)
}

//go:generate mockery --name=Transaction
type Transaction interface {
	Methods
	Commit() error
	Rollback() error
}

type Methods interface {
	Count(count *int64) error
	Create(value interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	Exec(sql string, values ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	First(dest interface{}) error
	FirstOrCreate(dest interface{}, conds ...interface{}) error
	ForceDelete(value interface{}, conds ...interface{}) error
	Get(dest interface{}) error
	Group(name string) Methods
	Having(query interface{}, args ...interface{}) Methods
	Join(query string, args ...interface{}) Methods
	Limit(limit int) Methods
	Model(value interface{}) Methods
	Offset(offset int) Methods
	Order(value interface{}) Methods
	OrWhere(query interface{}, args ...interface{}) Methods
	Pluck(column string, dest interface{}) error
	Raw(sql string, values ...interface{}) Methods
	Save(value interface{}) error
	Scan(dest interface{}) error
	Select(query interface{}, args ...interface{}) Methods
	Table(name string, args ...interface{}) Methods
	Update(column string, value interface{}) error
	Updates(values interface{}) error
	Where(query interface{}, args ...interface{}) Methods
	WithTrashed() Methods
	Scopes(funcs ...func(Methods) Methods) Methods
}
