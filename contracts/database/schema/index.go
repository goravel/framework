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
