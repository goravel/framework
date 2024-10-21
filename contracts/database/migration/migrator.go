package migration

const (
	MigratorDefault = "default"
	MigratorSql     = "sql"
)

type Migrator interface {
	// Create a new migration file.
	Create(name string) error
	// Fresh the migrations.
	Fresh() error
	// Run the migrations according to paths.
	Run() error
}
