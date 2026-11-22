package gorm

import (
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

type Conditions struct {
	dest                any
	err                 error
	model               any
	having              *contractsdriver.Having
	limit               *int
	offset              *int
	table               *Table
	groupBy             []string
	has                 []hasCondition
	join                []contractsdriver.Join
	omit                []string
	order               []any
	scopes              []func(contractsorm.Query) contractsorm.Query
	selectColumns       []string
	selectRaw           *Select
	where               []contractsdriver.Where
	with                []With
	distinct            bool
	lockForUpdate       bool
	sharedLock          bool
	withoutEvents       bool
	withoutGlobalScopes []string
	withTrashed         bool
}

type Select struct {
	query any
	args  []any
}

// hasCondition is a pending relationship filter. It is resolved when the query is built, because the parent model
// may only be known from the destination of a terminal operation.
type hasCondition struct {
	relationship string
	callback     func(contractsorm.Query) contractsorm.Query
	operator     string
	count        int64
}

type Table struct {
	name string
	args []any
}

type With struct {
	query string
	args  []any
}
