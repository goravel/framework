package gorm

import (
	ormcontract "github.com/goravel/framework/contracts/database/orm"
)

type Conditions struct {
	distinct      []any
	group         string
	having        *Having
	join          []Join
	limit         *int
	lockForUpdate bool
	model         any
	offset        *int
	omit          []string
	order         []any
	scopes        []func(ormcontract.Query) ormcontract.Query
	selectColumns *Select
	sharedLock    bool
	table         *Table
	where         []Where
	with          []With
	withoutEvents bool
	withTrashed   bool
}

type Having struct {
	query any
	args  []any
}

type Join struct {
	query string
	args  []any
}

type Select struct {
	query any
	args  []any
}

type Table struct {
	name string
	args []any
}

type Where struct {
	query any
	args  []any
	or    bool
}

type With struct {
	query string
	args  []any
}
