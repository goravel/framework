package console

import (
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateRefreshCommand struct {
	config config.Config
}

func NewMigrateRefreshCommand(config config.Config) *MigrateRefreshCommand {
	return &MigrateRefreshCommand{
		config: config,
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

	color.Greenln("Migration refresh success")

	return nil
}
