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

type Table struct {
	name string
	args []any
	// alias is the name the query refers to the table by. Only the relation
	// subquery sets it, for a quoted "table AS alias" expression that Table
	// can't parse an alias from.
	alias string
}

type With struct {
	query string
	args  []any
}
