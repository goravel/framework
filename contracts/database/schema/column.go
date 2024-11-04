package schema

type ColumnDefinition interface {
	// AutoIncrement set the column as auto increment
	AutoIncrement() ColumnDefinition
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
	// GetDefault returns the default value
	GetDefault() any
	// GetLength returns the length value
	GetLength() int
	// GetName returns the name value
	GetName() string
	// GetNullable returns the nullable value
	GetNullable() bool
	// GetType returns the type value
	GetType() string
	// Nullable allow NULL values to be inserted into the column
	Nullable() ColumnDefinition
	// Unsigned set the column as unsigned
	Unsigned() ColumnDefinition
}
