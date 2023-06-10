package console

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type MigrateCommand struct {
	config config.Config
}

func NewMigrateCommand(config config.Config) *MigrateCommand {
	return &MigrateCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateCommand) Signature() string {
	return "migrate"
}

// Description The console command description.
func (receiver *MigrateCommand) Description() string {
	return "Run the database migrations"
}

// Extend The console command extend.
func (receiver *MigrateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (receiver *MigrateCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellowln("Please fill database config first")

		return nil
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		color.Redln("Migration failed:", err.Error())

		return nil
	}

	color.Greenln("Migration success")

	return nil
}
