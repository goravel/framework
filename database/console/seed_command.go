package console

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	color "github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/seeder"
)

type SeedCommand struct {
	config config.Config
	facade seeder.Facade
}

func NewSeedCommand(config config.Config, facade seeder.Facade) *SeedCommand {
	return &SeedCommand{
		config: config,
		facade: facade,
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
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "force the operation to run when in production",
			},
			&command.StringFlag{
				Name:    "seeder",
				Value:   "",
				Aliases: []string{"s"},
				Usage:   "name of the seeder to run",
			},
		},
	}
}

// Handle executes the console command.
func (receiver *SeedCommand) Handle(ctx console.Context) error {
	if !receiver.ConfirmToProceed(ctx) {
		return nil
	}

	color.Greenln("Seeding database.")
	err := receiver.RunSeeder(ctx)
	if err != nil {
		log.Println(err)
	}
	return nil
}

// ConfirmToProceed determines if the command should proceed based on user confirmation.
func (receiver *SeedCommand) ConfirmToProceed(ctx console.Context) bool {
	force := ctx.OptionBool("force")
	if force || (receiver.config.Env("APP_ENV") != "production") {
		return true
	}

	// Display production alert message
	receiver.config.Add("alert", "Application In Production")

	alert := receiver.config.GetString("alert")

	color.Yellowln(alert)
	return false
}

// GetSeeder returns a seeder instance from the container.
func (receiver *SeedCommand) RunSeeder(ctx console.Context) error {
	class := ctx.Argument(0)
	seeders := receiver.facade
	if class == "" {
		class = ctx.Option("seeder")
	}
	if class == "" {
		// Run all seeders
		for _, item := range seeders.GetAllSeeder() {
			if item == nil {
				log.Println("No seeder found.")
				continue
			}
			item.Run()
		}
		return nil
	}
	class = "seeders." + class
	seeder := seeders.GetSeeder(class)
	if seeder == nil {
		log.Printf("No seeder of type %s found\n", class)
	}
	seeder.Run()
	return nil
}
