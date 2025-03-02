package db

type Conditions struct {
	table   string
	where   []Where
	orderBy []string
}

type Where struct {
	query any
	args  []any
	or    bool
}
