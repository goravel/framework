package schema

type ColumnDefinition interface {
	// Change the column
	Change()
	// Comment sets the comment value
	Comment(comment string) ColumnDefinition
	// GetComment returns the comment value
	GetComment() (comment string)
	// GetAllowed returns the allowed value
	GetAllowed() []string
	// GetAutoIncrement returns the autoIncrement value
	GetAutoIncrement() bool
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

type DecimalLength struct {
	Places int
	Total  int
}
