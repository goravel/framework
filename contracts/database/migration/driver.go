package migration

const (
	DriverDefault = "default"
	DriverSql     = "sql"
)

type Driver interface {
	// Create a new migration file.
	Create(name string) error
	// Run the migrations according to paths.
	Run(paths []string) error
}
