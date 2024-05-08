package schema

type ColumnDefinition interface {
	// AutoIncrement set the column as auto increment
	AutoIncrement() ColumnDefinition
	// Change the column
	Change()
	// Comment sets the comment value
	Comment(comment string) ColumnDefinition
	// Default set the default value
	Default(def any) ColumnDefinition
	// GetAllowed returns the allowed value
	GetAllowed() []string
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
	// GetChange returns the change value
	GetChange() bool
	// GetComment returns the comment value
	GetComment() (comment string)
	// GetDefault returns the default value
	GetDefault() any
	// GetLength returns the length value
	GetLength() int
	// GetName returns the name value
	GetName() string
	// GetNullable returns the nullable value
	GetNullable() bool
	// GetPlaces returns the places value
	GetPlaces() int
	// GetPrecision returns the precision value
	GetPrecision() int
	// GetTotal returns the total value
	GetTotal() int
	// GetType returns the type value
	GetType() string
	// GetUnsigned returns the unsigned value
	GetUnsigned() bool
	// Nullable allow NULL values to be inserted into the column
	Nullable() ColumnDefinition
	// Places set the decimal places
	Places(places int) ColumnDefinition
	// Total set the decimal total
	Total(total int) ColumnDefinition
	// Unsigned set the column as unsigned
	Unsigned() ColumnDefinition
}

type Column struct {
	AutoIncrement bool
	Collation     string
	Comment       string
	Default       string
	Name          string
	Nullable      bool
	Type          string
	TypeName      string
}

type DecimalConfig struct {
	Places int
	Total  int
}
