package gorm

import (
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

// WhereType type of where condition
type WhereType int

const (
	WhereTypeBase WhereType = iota
	WhereTypeJsonContains
	WhereTypeJsonContainsKey
	WhereTypeJsonLength
)

type Conditions struct {
	distinct      []any
	groupBy       []string
	having        *Having
	join          []Join
	limit         *int
	lockForUpdate bool
	model         any
	offset        *int
	omit          []string
	order         []any
	scopes        []func(contractsorm.Query) contractsorm.Query
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
	ttype WhereType
	query any
	args  []any
	or    bool
	isNot bool
}

type With struct {
	query string
	args  []any
}
