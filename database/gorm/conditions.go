package gorm

import (
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

type Conditions struct {
	dest                any
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
	selectSubs          []selectSub
	where               []contractsdriver.Where
	eagerLoad           []eagerLoadEntry
	relations           []relationExistence
	oneOfMany           *oneOfManyConfig
	distinct            bool
	lockForUpdate       bool
	sharedLock          bool
	withoutEvents       bool
	withoutGlobalScopes []string
	withTrashed         bool
}

// oneOfManyConfig captures the column + aggregate for OfMany / LatestOfMany / OldestOfMany when
// they are called inside a With() eager-load callback. It's read by runRelatedQuery, which
// rewrites the inner query into an INNER JOIN over a per-parent aggregate subquery.
type oneOfManyConfig struct {
	column    string
	aggregate string // "MAX" | "MIN" | other SQL aggregate
}

type Select struct {
	query any
	args  []any
}

type Table struct {
	name string
	args []any
}

// selectSub describes a deferred sub-select aggregate (WithCount / WithMax / etc.).
// The relation is resolved at buildConditions() time, when the parent model is known.
type selectSub struct {
	relation string
	column   string
	function string // count | max | min | sum | avg | exists
	alias    string
	callback contractsorm.RelationCallback
}

// relationExistence describes a deferred relationship existence/absence condition.
// Building is deferred so the parent model can be resolved from conditions.model or conditions.dest
// (the latter is set by Find/First/Get when the user passes a dest).
type relationExistence struct {
	relation    string
	operator    string
	count       int
	conjunction string // "and" | "or"
	callback    contractsorm.RelationCallback

	// morph specifics (zero-valued for non-morph queries)
	morphTypes    []any
	morphCallback contractsorm.MorphRelationCallback
}
