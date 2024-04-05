package schema

type Schema interface {
	// Connection Get instance by database connection.
	Connection(name string) Schema
	// Create a new table on the schema.
	Create(table string, callback func(table Blueprint)) error
	// Drop a table from the schema.
	Drop(table string) error
	// DropAllTables Drop all tables from the database.
	DropAllTables() error
	// DropAllViews Drop all views from the database.
	DropAllViews() error
	// DropColumns Drop columns from a table schema.
	DropColumns(table string, columns []string) error
	// DropIfExists Drop a table from the schema if it exists.
	DropIfExists(table string) error
	// GetColumns Get the columns for a given table.
	GetColumns(table string) ([]Column, error)
	// GetColumnListing Get the column listing for a given table.
	GetColumnListing(table string) []string
	// GetIndexes Get the indexes for a given table.
	GetIndexes(table string) ([]Index, error)
	// GetIndexListing Get the names of the indexes for a given table.
	GetIndexListing(table string) []string
	// GetTableListing Get the names of the tables that belong to the database.
	GetTableListing() []string
	// GetTables Get the tables that belong to the database.
	GetTables() ([]Table, error)
	// GetViews Get the views that belong to the database.
	GetViews() []View
	// HasColumn Determine if the given table has a given column.
	HasColumn(table, column string) bool
	// HasColumns Determine if the given table has given columns.
	HasColumns(table string, columns []string) bool
	// HasIndex Determine if the given table has a given index.
	HasIndex(table, index string) bool
	// HasTable Determine if the given table exists.
	HasTable(table string) bool
	// HasView Determine if the given view exists.
	HasView(view string) bool
	// Register migrations.
	Register([]Migration)
	// Rename a table on the schema.
	Rename(from, to string)
	// Table Modify a table on the schema.
	Table(table string, callback func(table Blueprint)) error
}

type Command struct {
	Algorithm string
	Column    ColumnDefinition
	Columns   []string
	Value     string
	Name      string
}

type Index struct {
	Columns []string
	Name    string
	Primary bool
	Type    string
	Unique  bool
}

type Table struct {
	Comment string
	Name    string
	Size    int
}

type View struct {
	Definition string
	Name       string
}
