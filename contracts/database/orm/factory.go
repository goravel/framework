package orm

//go:generate mockery --name=Factory
type Factory interface {
	// Count Sets the number of models that should be generated.
	Count(count int) Factory
	// Create Creates a model and persists it to the database.
	Create(value any, attributes ...map[string]any) error
	// CreateQuietly Creates a model and persists it to the database without firing any model events.
	CreateQuietly(value any, attributes ...map[string]any) error
	// Make Creates a model and returns it, but does not persist it to the database.
	Make(value any, attributes ...map[string]any) error
}
