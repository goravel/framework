package schema

type ForeignKeyDefinition interface {
	CascadeOnDelete() ForeignKeyDefinition
	CascadeOnUpdate() ForeignKeyDefinition
	On(table string) ForeignKeyDefinition
	Name(name string) ForeignKeyDefinition
	NoActionOnDelete() ForeignKeyDefinition
	NoActionOnUpdate() ForeignKeyDefinition
	NullOnDelete() ForeignKeyDefinition
	References(columns ...string) ForeignKeyDefinition
	RestrictOnDelete() ForeignKeyDefinition
	RestrictOnUpdate() ForeignKeyDefinition
}

type IndexDefinition interface {
	Algorithm(algorithm string) IndexDefinition
	Deferrable() IndexDefinition
	InitiallyImmediate() IndexDefinition
	Language(name string) IndexDefinition
	Name(name string) IndexDefinition
}

type IndexConfig struct {
	Algorithm string
	Name      string
	Language  string
}
