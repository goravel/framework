package db

type Conditions struct {
	table   string
	where   []Where
	orderBy []string
	selects []string
	limit   *uint64
}

type Where struct {
	query any
	args  []any
	or    bool
}
