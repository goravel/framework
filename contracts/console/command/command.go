package command

type Extend struct {
	Category string
	Flags    []Flag
}

type Flag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    string
}
