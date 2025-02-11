package db

type Conditions struct {
	table string
	where []Where
}

type Where struct {
	query any
	args  []any
	// or    bool
}
