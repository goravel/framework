package migration

type File struct {
	ID        uint
	Migration string
	Batch     int
}

type Repository interface {
	// CreateRepository Create the migration repository data store.
	CreateRepository() error
	// Delete Remove a migration from the log.
	Delete(migration string) error
	// DeleteRepository Delete the migration repository data store.
	DeleteRepository() error
	// GetLast Get the last migration batch.
	GetLast() ([]File, error)
	// GetMigrations Get the list of migrations.
	GetMigrations(steps int) ([]File, error)
	// GetMigrationsByBatch Get the list of the migrations by batch.
	GetMigrationsByBatch(batch int) ([]File, error)
	// GetNextBatchNumber Get the next migration batch number.
	GetNextBatchNumber() (int, error)
	// GetRan Get the completed migrations.
	GetRan() ([]string, error)
	// Log that a migration was run.
	Log(file string, batch int) error
	// RepositoryExists Determine if the migration repository exists.
	RepositoryExists() bool
}
