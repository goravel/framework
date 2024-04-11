package schema

type ForeignKeyDefinition interface {
	CascadeOnDelete() ForeignKeyDefinition
	CascadeOnUpdate() ForeignKeyDefinition
	On(table string) ForeignKeyDefinition
	NoActionOnDelete() ForeignKeyDefinition
	NoActionOnUpdate() ForeignKeyDefinition
	NullOnDelete() ForeignKeyDefinition
	References(columns ...string) ForeignKeyDefinition
	RestrictOnDelete() ForeignKeyDefinition
	RestrictOnUpdate() ForeignKeyDefinition
}
