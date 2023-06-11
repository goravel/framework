package console

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	// "strings"

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
}

func NewSeedCommand(config config.Config) *SeedCommand {
	return &SeedCommand{
		config: config,
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

// Handle Execute the console command.
func (receiver *SeedCommand) Handle(ctx console.Context) error {
	if !receiver.ConfirmToProceed(ctx) {
		return nil
	}

	color.Greenln("Seeding database.")

	previousConnection := receiver.config.GetString("database.default")

	receiver.SetDatabase(ctx)

	receiver.GetSeeder(ctx).Run(ctx)

	// Reset the previous connection if available
	if previousConnection != "" {
		receiver.config.Add("database.default", previousConnection)
	}

	return nil
}

// ConfirmToProceed determines if the command should proceed.
func (receiver *SeedCommand) ConfirmToProceed(ctx console.Context) bool {
	force := ctx.Option("force")
	if force == "true" || (force == "" && receiver.config.Env("APP_ENV") != "production") {
		return true
	}

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

// getDatabase returns the name of the database connection to use.
func (receiver *SeedCommand) GetDatabase(ctx console.Context) string {
	database := ctx.Option("database")
	if database == "" {
		return receiver.config.GetString("database.default")
	}
	return database
}

// SetDatabase sets the database connection based on the provided option or the default value from the config.
func (receiver *SeedCommand) SetDatabase(ctx console.Context) {
	database := ctx.Option("database")
	if database == "" {
		database = receiver.config.GetString("database.default")
	}

	receiver.config.Add("database.default", database)
}

// GetSeeder returns a seeder instance from the container.
// GetSeeder returns a seeder instance from the container.
func (receiver *SeedCommand) GetSeeder(ctx console.Context) database.Seeder {
	class := ctx.Argument(0)
	if class == "" {
		class = ctx.Option("class")
	}

	if class != "" && !strings.Contains(class, "\\") {
		class = "Database\\Seeders\\" + class
	}

	if class == "Database\\Seeders\\DatabaseSeeder" && !receiver.ClassExists(class) {
		class = "DatabaseSeeder"
	}

	var container foundation.Container
	instance, err := container.Make(class)

	if err != nil {
		// Handle the error if necessary
		return nil
	}
	// Check if the resolved instance implements the Seeder interface
	seeder, ok := instance.(database.Seeder)
	if !ok {
		// Handle the case where the resolved instance does not implement the Seeder interface
		return nil
	}
	// Set the container and command on the seeder instance
	seeder.SetContainer(container)
	seeder.SetCommand(ctx)
	log.Println("new class", class)
	return seeder
}

// ClassExists checks if a class exists in the project.
// ClassExists checks if a class exists in the project.
func (receiver *SeedCommand) ClassExists(class string) bool {
	_, found := reflect.TypeOf(receiver.config).Elem().FieldByName(class)
	return found
}
