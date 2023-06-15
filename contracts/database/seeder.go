package database

import (
	"github.com/goravel/framework/contracts/console"
)

type Seeder interface {
	// Register registers seeders.
	Register(seeders []Seeder)

	// Run executes the seeder logic.
	Run(ctx console.Context) error

	// SetCommand sets the console command instance on the seeder.
	SetCommand(command console.Context)

	GetSeeder(name string) Seeder
}
