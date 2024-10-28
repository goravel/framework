package migration

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/migration"
)

type MigrateStatusCommand struct {
	migrator migration.Migrator
}

func NewMigrateStatusCommand(migrator migration.Migrator) *MigrateStatusCommand {
	return &MigrateStatusCommand{
		migrator: migrator,
	}
}

// Signature The name and signature of the console command.
func (r *MigrateStatusCommand) Signature() string {
	return "migrate:status"
}

// Description The console command description.
func (r *MigrateStatusCommand) Description() string {
	return "Show the status of each migration"
}

// Extend The console command extend.
func (r *MigrateStatusCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (r *MigrateStatusCommand) Handle(ctx console.Context) error {
	if err := r.migrator.Status(); err != nil {
		ctx.Error(err.Error())
	}

	return nil
}
