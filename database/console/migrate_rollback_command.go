package console

import (
	"errors"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type MigrateRollbackCommand struct {
	config config.Config
}

func NewMigrateRollbackCommand(config config.Config) *MigrateRollbackCommand {
	return &MigrateRollbackCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateRollbackCommand) Signature() string {
	return "migrate:rollback"
}

// Description The console command description.
func (receiver *MigrateRollbackCommand) Description() string {
	return "Rollback the database migrations"
}

// Extend The console command extend.
func (receiver *MigrateRollbackCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "step",
				Value: "1",
				Usage: "rollback steps",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *MigrateRollbackCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellow().Println("Please fill database config first")

		return nil
	}

	stepString := "-" + ctx.Option("step")
	step, err := strconv.Atoi(stepString)
	if err != nil {
		color.Red().Println("Migration rollback failed: invalid step", ctx.Option("step"))

		return nil
	}

	if err = m.Steps(step); err != nil && !errors.Is(err, migrate.ErrNoChange) && !errors.Is(err, migrate.ErrNilVersion) {
		var errShortLimit migrate.ErrShortLimit
		switch {
		case errors.As(err, &errShortLimit):
		default:
			color.Red().Println("Migration rollback failed:", err.Error())

			return nil
		}
	}

	color.Green().Println("Migration rollback success")

	return nil
}
