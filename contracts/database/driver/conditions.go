package driver

// WhereType type of where condition
type WhereType int

const (
	WhereTypeBase WhereType = iota
	WhereTypeJsonContains
	WhereTypeJsonContainsKey
	WhereTypeJsonLength
	// WhereRelation used for where cause with relation subqueries, like WhereHas, OrWhereHas etc.
	WhereRelation
)

type Conditions struct {
	CrossJoin     []Join
	Distinct      *bool
	GroupBy       []string
	Having        *Having
	Join          []Join
	InRandomOrder *bool
	LeftJoin      []Join
	LockForUpdate *bool
	Limit         *uint64
	Offset        *uint64
	OrderBy       []string
	RightJoin     []Join
	Selects       []string
	SharedLock    *bool
	Table         string
	Where         []Where
}

type Having struct {
	Query any
	Args  []any
}

type Join struct {
	Query string
	Args  []any
}

type Where struct {
	Query    any
	Args     []any
	Type     WhereType
	Relation string
	Or       bool
	IsNot    bool
}
