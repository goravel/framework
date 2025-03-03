package db

type Conditions struct {
	Distinct *bool
	Limit    *uint64
	OrderBy  []string
	Selects  []string
	Table    string
	Where    []Where
}

type Where struct {
	query any
	args  []any
	or    bool
}
