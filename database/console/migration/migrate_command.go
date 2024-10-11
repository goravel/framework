package migration

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
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
	driver, err := GetDriver(receiver.config)
	if err != nil {
		return err
	}

	// support for multiple migration paths in the future
	path := []string{"./database/migrations"}
	if err := driver.Run(path); err != nil {
		color.Red().Println("Migration failed:", err.Error())
		return nil
	}

	color.Green().Println("Migration success")

	return nil
}
