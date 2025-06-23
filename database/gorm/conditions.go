package gorm

import (
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

type Conditions struct {
	dest          any
	distinct      bool
	groupBy       []string
	having        *contractsdriver.Having
	join          []contractsdriver.Join
	limit         *int
	lockForUpdate bool
	model         any
	offset        *int
	omit          []string
	order         []any
	scopes        []func(contractsorm.Query) contractsorm.Query
	selectColumns []string
	sharedLock    bool
	table         *Table
	where         []contractsdriver.Where
	with          []With
	withoutEvents bool
	withTrashed   bool
}

type Table struct {
	name string
	args []any
}

type With struct {
	query string
	args  []any
}
