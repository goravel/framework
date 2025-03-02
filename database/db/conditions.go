package db

type Conditions struct {
	Table   string
	Where   []Where
	OrderBy []string
	Selects []string
	Limit   *uint64
}

type Where struct {
	query any
	args  []any
	or    bool
}
