package db

type Conditions struct {
	Distinct *bool
	GroupBy  []string
	Having   *Having
	Limit    *uint64
	OrderBy  []string
	Selects  []string
	Table    string
	Where    []Where
}

type Having struct {
	query any
	args  []any
}

type Where struct {
	query any
	args  []any
	or    bool
}
