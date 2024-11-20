package schema

type ColumnDefinition interface {
	// AutoIncrement set the column as auto increment
	AutoIncrement() ColumnDefinition
	// Comment sets the comment value
	Comment(comment string) ColumnDefinition
	// GetAllowed returns the allowed value
	GetAllowed() []string
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
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
	// IsSetComment returns true if the comment value is set
	IsSetComment() bool
	// Places set the decimal places
	Places(places int) ColumnDefinition
	// Total set the decimal total
	Total(total int) ColumnDefinition
	// Nullable allow NULL values to be inserted into the column
	Nullable() ColumnDefinition
	// Unsigned set the column as unsigned
	Unsigned() ColumnDefinition
}

type Column struct {
	Autoincrement bool
	Collation     string
	Comment       string
	Default       string
	Name          string
	Nullable      bool
	Type          string
	TypeName      string
}
