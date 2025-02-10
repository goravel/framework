package schema

type ColumnDefinition interface {
	// After Place the column "after" another column (MySQL only)
	After(column string) ColumnDefinition
	// AutoIncrement set the column as auto increment
	AutoIncrement() ColumnDefinition
	// Change the column (MySQL / PostgreSQL / SQL Server)
	Change() ColumnDefinition
	// Comment sets the comment value (MySQL / PostgreSQL)
	Comment(comment string) ColumnDefinition
	// Default set the default value
	Default(def any) ColumnDefinition
	// First Place the column "first" in the table (MySQL only)
	First() ColumnDefinition
	// GetAfter returns the after value
	GetAfter() string
	// GetAllowed returns the allowed value
	GetAllowed() []any
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
	// GetOnUpdate returns the onUpdate value
	GetOnUpdate() any
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
	// GetUseCurrent returns the useCurrent value
	GetUseCurrent() bool
	// GetUseCurrentOnUpdate returns the useCurrentOnUpdate value
	GetUseCurrentOnUpdate() bool
	// IsChange returns true if the column has changed
	IsChange() bool
	// IsFirst returns true if the column is first
	IsFirst() bool
	// IsSetComment returns true if the comment value is set
	IsSetComment() bool
	// OnUpdate sets the column to use the value on update (Mysql only)
	OnUpdate(value any) ColumnDefinition
	// Places set the decimal places
	Places(places int) ColumnDefinition
	// Total set the decimal total
	Total(total int) ColumnDefinition
	// Nullable allow NULL values to be inserted into the column
	Nullable() ColumnDefinition
	// Unsigned set the column as unsigned (Mysql only)
	Unsigned() ColumnDefinition
	// UseCurrent set the column to use the current timestamp
	UseCurrent() ColumnDefinition
	// UseCurrentOnUpdate set the column to use the current timestamp on update (Mysql only)
	UseCurrentOnUpdate() ColumnDefinition
}

type Column struct {
	Autoincrement bool
	Collation     string
	Comment       string
	Default       string
	Extra         string
	Name          string
	Nullable      bool
	Type          string
	TypeName      string
}
