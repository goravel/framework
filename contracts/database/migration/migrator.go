package migration

const (
	MigratorDefault = "default"
	MigratorSql     = "sql"
)

type Status struct {
	Name  string
	Batch int
	Ran   bool
}

type Migrator interface {
	// Create a new migration file.
	Create(name string, modelName string) (string, error)
	// Fresh the migrations.
	Fresh() error
	// Reset the migrations.
	Reset() error
	// Rollback the last migration operation.
	Rollback(step, batch int) error
	// Run the migrations according to paths.
	Run() error
	// Status get the migration's status.
	Status() ([]Status, error)
}
