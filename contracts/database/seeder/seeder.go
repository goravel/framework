package seeder

type Facade interface {
	// Register registers seeders.
	Register(seeders []Seeder)

	// GetSeeder gets a seeder instance from the seeders.
	GetSeeder(name string) Seeder
}
type Seeder interface {
	// Run executes the seeder logic.
	Run() error
}
