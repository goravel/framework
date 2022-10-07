package command

type Extend struct {
	Category string
	Flags    []Flag
}

type Context interface {
	Argument(index int) string
	Arguments() []string
	Option(key string) string
}

type Flag struct {
	Name     string
	Aliases  []string
	Usage    string
	Required bool
	Value    string
}
