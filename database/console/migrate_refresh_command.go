package console

import (
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateRefreshCommand struct {
	config  config.Config
	artisan console.Artisan
}

func NewMigrateRefreshCommand(config config.Config, artisan console.Artisan) *MigrateRefreshCommand {
	return &MigrateRefreshCommand{
		config:  config,
		artisan: artisan,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateRefreshCommand) Signature() string {
	return "migrate:refresh"
}

// Description The console command description.
func (receiver *MigrateRefreshCommand) Description() string {
	return "Reset and re-run all migrations"
}

// Extend The console command extend.
func (receiver *MigrateRefreshCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "step",
				Value: "",
				Usage: "refresh steps",
			},
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
func (receiver *MigrateRefreshCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")

		return nil
	}

	if step := ctx.Option("step"); step != "" {
		stepString := "-" + step
		s, err := strconv.Atoi(stepString)
		if err != nil {
			color.Redln("Migration refresh failed: invalid step", ctx.Option("step"))

			return nil
		}

		if err = m.Steps(s); err != nil && err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
			switch err.(type) {
			case migrate.ErrShortLimit:
			default:
				color.Redln("Migration refresh failed:", err.Error())

				return nil
			}
		}
	} else {
		if err = m.Down(); err != nil && err != migrate.ErrNoChange {
			color.Redln("Migration reset failed:", err.Error())

			return nil
		}
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration refresh failed:", err.Error())

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
	color.Greenln("Migration refresh success")

	return nil
}
