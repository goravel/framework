package migration

import (
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
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
func (r *MigrateRefreshCommand) Signature() string {
	return "migrate:refresh"
}

// Description The console command description.
func (r *MigrateRefreshCommand) Description() string {
	return "Reset and re-run all migrations"
}

// Extend The console command extend.
func (r *MigrateRefreshCommand) Extend() command.Extend {
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
func (r *MigrateRefreshCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(r.config)
	if err != nil {
		return err
	}
	if m == nil {
		ctx.Error(errors.ConsoleEmptyDatabaseConfig.Error())

		return nil
	}

	if step := ctx.Option("step"); step != "" {
		stepString := "-" + step
		s, err := strconv.Atoi(stepString)
		if err != nil {
			ctx.Error(errors.MigrationRefreshFailed.Args(err).Error())

			return nil
		}

		if err = m.Steps(s); err != nil && !errors.Is(err, migrate.ErrNoChange) && !errors.Is(err, migrate.ErrNilVersion) {
			var errShortLimit migrate.ErrShortLimit
			switch {
			case errors.As(err, &errShortLimit):
			default:
				ctx.Error(errors.MigrationRefreshFailed.Args(err).Error())

				return nil
			}
		}
	} else {
		if err = m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			ctx.Error(errors.MigrationRefreshFailed.Args(err).Error())
			return nil
		}
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		ctx.Error(errors.MigrationRefreshFailed.Args(err).Error())
		return nil
	}

	// Seed the database if the "seed" flag is provided
	if ctx.OptionBool("seed") {
		seeders := ctx.OptionSlice("seeder")
		seederFlag := ""
		if len(seeders) > 0 {
			seederFlag = " --seeder " + strings.Join(seeders, ",")
		}

		if err := r.artisan.Call("db:seed" + seederFlag); err != nil {
			ctx.Error(errors.MigrationRefreshFailed.Args(err).Error())
			return nil
		}
	}
	ctx.Info("Migration refresh success")

	return nil
}
