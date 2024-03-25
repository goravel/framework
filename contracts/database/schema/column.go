package schema

type ColumnDefinition interface {
	// Change the column
	Change()
	// Comment sets the comment value
	Comment(comment string) ColumnDefinition
	// GetAllowed returns the allowed value
	GetAllowed() []string
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
	// GetComment returns the comment value
	GetComment() (comment string)
	// GetLength returns the length value
	GetLength() int
	// GetName returns the name value
	GetName() string
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

type IntegerConfig struct {
	AutoIncrement bool
	Unsigned      bool
}
