package console

import (
	"errors"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type MigrateFreshCommand struct {
	config  config.Config
	artisan console.Artisan
}

func NewMigrateFreshCommand(config config.Config, artisan console.Artisan) *MigrateFreshCommand {
	return &MigrateFreshCommand{
		config:  config,
		artisan: artisan,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateFreshCommand) Signature() string {
	return "migrate:fresh"
}

// Description The console command description.
func (receiver *MigrateFreshCommand) Description() string {
	return "Drop all tables and re-run all migrations"
}

// Extend The console command extend.
func (receiver *MigrateFreshCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "seed",
				Usage: "seed the database after running migrations",
			},
			&command.StringSliceFlag{
				Name:  "seeder",
				Usage: "specify the seeder(s) to use for seeding the database",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *MigrateFreshCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellow().Println("Please fill database config first")
		return nil
	}

	if err = m.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		color.Red().Println("Migration failed:", err.Error())
		return nil
	}

	m2, err2 := getMigrate(receiver.config)
	if err2 != nil {
		return err2
	}
	if m2 == nil {
		color.Yellow().Println("Please fill database config first")
		return nil
	}

	if err2 = m2.Up(); err2 != nil && !errors.Is(err2, migrate.ErrNoChange) {
		color.Red().Println("Migration failed:", err2.Error())
		return nil
	}

	// Seed the database if the "seed" flag is provided
	if ctx.OptionBool("seed") {
		seeders := ctx.OptionSlice("seeder")
		seederFlag := ""
		if len(seeders) > 0 {
			seederFlag = " --seeder " + strings.Join(seeders, ",")
		}
		receiver.artisan.Call("db:seed" + seederFlag)
	}

	color.Green().Println("Migration fresh success")

	return nil
}
