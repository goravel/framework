package schema

type Schema interface {
	// Create a new table on the schema.
	//Create(table string, callback func(table Blueprint))
	// Connection Get the connection for the schema.
	Connection() Schema
	// DropIfExists Drop a table from the schema if exists.
	//DropIfExists(table string)
	// Register migrations.
	Register([]Migration)
	// Sql Execute a sql directly.
	Sql(callback func(table Blueprint))
	// Table Modify a table on the schema.
	//Table(table string, callback func(table Blueprint))
}

type Migration interface {
	// Signature Get the migration signature.
	Signature() string
	// Up Run the migrations.
	Up()
	// Down Reverse the migrations.
	Down()
}
