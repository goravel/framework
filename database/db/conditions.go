package db

type Conditions struct {
	CrossJoin []Join
	Distinct  *bool
	GroupBy   []string
	Having    *Having
	Join      []Join
	LeftJoin  []Join
	Limit     *uint64
	OrderBy   []string
	RightJoin []Join
	Selects   []string
	Table     string
	Where     []Where
}

type Having struct {
	query any
	args  []any
}

type Join struct {
	query string
	args  []any
}

type Where struct {
	query any
	args  []any
	or    bool
}
