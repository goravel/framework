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
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "seeder",
				Value:   "DatabaseSeeder",
				Aliases: []string{"s"},
				Usage:   "name of the seeder to run",
			},
		},
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
		class = ctx.Option("seeder")
	}
	class = "seeders." + class
	instance, err := receiver.app.Make("goravel.seeder")
	if err != nil {
		log.Println("Failed to resolve seeder instance:", err)
		return nil
	}
	seeders, ok := instance.(database.Seeder)
	if !ok {
		log.Println("Resolved instance does not implement the Seeder interface")
		return nil
	}
	seeder := seeders.GetSeeder(class)
	if seeder == nil {
		log.Printf("No seeder of type %s found\n", class)
	}
	seeder.SetCommand(ctx)
	return seeder
}
