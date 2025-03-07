package driver

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
	Query any
	Args  []any
	Or    bool
}
