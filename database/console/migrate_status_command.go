package console

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
)

type MigrateStatusCommand struct {
	config config.Config
}

func NewMigrateStatusCommand(config config.Config) *MigrateStatusCommand {
	return &MigrateStatusCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *MigrateStatusCommand) Signature() string {
	return "migrate:status"
}

// Description The console command description.
func (receiver *MigrateStatusCommand) Description() string {
	return "Show the status of each migration"
}

// Extend The console command extend.
func (receiver *MigrateStatusCommand) Extend() command.Extend {
	return command.Extend{
		Category: "migrate",
	}
}

// Handle Execute the console command.
func (receiver *MigrateStatusCommand) Handle(ctx console.Context) error {
	m, err := getMigrate(receiver.config)
	if err != nil {
		return err
	}
	if m == nil {
		color.Yellow().Println("Please fill database config first")
		return nil
	}

	version, dirty, err := m.Version()
	if err != nil {
		color.Red().Println("Migration status failed:", err.Error())

		return nil
	}

	if dirty {
		color.Yellow().Println("Migration status: dirty")
		color.Green().Println("Migration version:", version)

		return nil
	}

	color.Green().Println("Migration status: clean")
	color.Green().Println("Migration version:", version)

	return nil
}
