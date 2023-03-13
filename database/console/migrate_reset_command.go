package console

import (

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateResetCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *MigrateResetCommand) Signature() string {
	return "migrate:reset"
}

//Description The console command description.
func (receiver *MigrateResetCommand) Description() string {
	return "Rollback all database migrations"
}

//Extend The console command extend.
func (receiver *MigrateResetCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

//Handle Execute the console command.
func (receiver *MigrateResetCommand) Handle(ctx console.Context) error {
	m, err := getMigrate()
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")

		return nil
	}

	// Rollback all migrations
    for {
        err := m.Down()
        if err == migrate.ErrNoChange || err == migrate.ErrNilVersion {
            // All migrations have been rolled back
            break
        }
        if err != nil {
            return err
        }
    }

	color.Greenln("Migration reset success")

	return nil
}
