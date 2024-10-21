package migration

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/errors"
)

type MigrateRollbackCommand struct {
	migrator migration.Migrator
}

func NewMigrateRollbackCommand(migrator migration.Migrator) *MigrateRollbackCommand {
	return &MigrateRollbackCommand{
		migrator: migrator,
	}
}

// Signature The name and signature of the console command.
func (r *MigrateRollbackCommand) Signature() string {
	return "migrate:rollback"
}

// Description The console command description.
func (r *MigrateRollbackCommand) Description() string {
	return "Rollback the database migrations"
}

// Extend The console command extend.
func (r *MigrateRollbackCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
		Flags: []command.Flag{
			&command.IntFlag{
				Name:  "step",
				Value: 1,
				Usage: "rollback steps",
			},
			&command.IntFlag{
				Name:  "batch",
				Value: 0,
				Usage: "rollback batch number (only can be used in the default driver)",
			},
		},
	}
}

// Handle Execute the console command.
func (r *MigrateRollbackCommand) Handle(ctx console.Context) error {
	var step, batch int
	if step = ctx.OptionInt("step"); step == 0 {
		if batch = ctx.OptionInt("step"); batch == 0 {
			step = 1
		}
	}

	if err := r.migrator.Rollback(step, batch); err != nil {
		ctx.Error(errors.MigrationMigrateFailed.Args(err).Error())
		return nil
	}

	ctx.Info("Migration rollback success")

	return nil
}
