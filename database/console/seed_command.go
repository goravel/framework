package console

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	color "github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/foundation"
)

type SeedCommand struct {
	config config.Config
	app    foundation.Application
}

func NewSeedCommand(config config.Config, app foundation.Application) *SeedCommand {
	return &SeedCommand{
		config: config,
		app:    app,
	}
}

// Signature The name and signature of the console command.
func (receiver *SeedCommand) Signature() string {
	return "db:seed"
}

// Description The console command description.
func (receiver *SeedCommand) Description() string {
	return "Seed the database with records"
}

// Extend The console command extend.
func (receiver *SeedCommand) Extend() command.Extend {
	return command.Extend{
		Category: "db",
	}
}

// Handle executes the console command.
func (receiver *SeedCommand) Handle(ctx console.Context) error {
	if !receiver.ConfirmToProceed(ctx) {
		log.Println("Confirmation to proceed denied.")
		return nil
	}

	color.Greenln("Seeding database.")
	seeder := receiver.GetSeeder(ctx)
	if seeder == nil {
		log.Println("No valid seeder instance found.")
		return nil
	}
	seeder.Run(ctx)
	return nil
}

// ConfirmToProceed determines if the command should proceed based on user confirmation.
func (receiver *SeedCommand) ConfirmToProceed(ctx console.Context) bool {
	force := ctx.Option("force")
	if force == "true" || (force == "" && receiver.config.Env("APP_ENV") != "production") {
		return true
	}

	// Display confirmation message
	receiver.config.Add("alert", "Application In Production")
	receiver.config.Add("confirm", "Do you really wish to run this command?")
	receiver.config.Add("cancel", "Command canceled.")

	alert := receiver.config.GetString("alert")
	cancel := receiver.config.GetString("cancel")

	fmt.Println(alert)

	confirmed := receiver.config.GetBool("confirmed")
	if !confirmed {
		fmt.Println(cancel)
		return false
	}
	return true
}

// GetSeeder returns a seeder instance from the container.
func (receiver *SeedCommand) GetSeeder(ctx console.Context) database.Seeder {
	class := ctx.Argument(0)
	if class == "" {
		class = "DatabaseSeeder"
	}
	class = "seeders." + class

	instance, err := receiver.app.Make(class)
	if err != nil {
		log.Println("Failed to resolve seeder instance:", err)
		return nil
	}

	seeder, ok := instance.(database.Seeder)
	if !ok {
		log.Println("Resolved instance does not implement the Seeder interface")
		return nil
	}

	seeder.SetCommand(ctx)
	return seeder
}
