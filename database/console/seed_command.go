package console

import (
	"errors"

	color "github.com/gookit/color"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/seeder"
)

type SeedCommand struct {
	config config.Config
	seeder seeder.Facade
}

func NewSeedCommand(config config.Config, seeder seeder.Facade) *SeedCommand {
	return &SeedCommand{
		config: config,
		seeder: seeder,
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
			&command.StringSliceFlag{
				Name:    "seeder",
				Value:   []string{},
				Aliases: []string{"s"},
				Usage:   "name of the seeder to run",
			},
		},
	}
}

// Handle executes the console command.
func (receiver *SeedCommand) Handle(ctx console.Context) error {
	err := receiver.ConfirmToProceed(ctx)
	if err != nil {
		color.Redln(err)
		return nil
	}

	names := ctx.OptionSlice("seeder")
	seeders := receiver.GetSeeders(names)
	if seeders == nil {
		return nil
	}
	color.Greenln("Seeding database.")
	for _, seeder := range seeders {
		err := seeder.Run()
		if err != nil {
			color.Redf("Error running seeder: %v\n", err)
			return nil
		}
	}

	return nil
}

// ConfirmToProceed determines if the command should proceed based on user confirmation.
func (receiver *SeedCommand) ConfirmToProceed(ctx console.Context) error {
	force := ctx.OptionBool("force")
	if force || (receiver.config.Env("APP_ENV") != "production") {
		return nil
	}
	return errors.New("application in production use --force to run this command")
}

// GetSeeder returns a seeder instance from the container.
func (receiver *SeedCommand) GetSeeders(names []string) []seeder.Seeder {
	if len(names) == 0 {
		return receiver.seeder.GetSeeders()
	}
	var seeders []seeder.Seeder
	for _, name := range names {
		class := "seeders." + name
		seeder := receiver.seeder.GetSeeder(class)
		if seeder == nil {
			color.Redf("No seeder of type %s found\n", class)
			return nil
		}
		seeders = append(seeders, seeder)
	}
	return seeders
}
