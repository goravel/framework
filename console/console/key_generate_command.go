package console

import (
	"os"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/str"
)

type KeyGenerateCommand struct {
	config config.Config
}

func NewKeyGenerateCommand(config config.Config) *KeyGenerateCommand {
	return &KeyGenerateCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (receiver *KeyGenerateCommand) Signature() string {
	return "key:generate"
}

// Description The console command description.
func (receiver *KeyGenerateCommand) Description() string {
	return "Set the application key"
}

// Extend The console command extend.
func (receiver *KeyGenerateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "key",
	}
}

// Handle Execute the console command.
func (receiver *KeyGenerateCommand) Handle(ctx console.Context) error {
	if receiver.config.GetString("app.env") == "production" {
		color.Warningln("**************************************")
		color.Warningln("*     Application In Production!     *")
		color.Warningln("**************************************")

		answer, err := ctx.Confirm("Do you really wish to run this command?")
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}

		if !answer {
			ctx.Warning("Command cancelled!")
			return nil
		}
	}

	key := receiver.generateRandomKey()
	if err := receiver.writeNewEnvironmentFileWith(key); err != nil {
		ctx.Error(err.Error())

		return nil
	}

	ctx.Success("Application key set successfully")

	return nil
}

// generateRandomKey Generate a random key for the application.
func (receiver *KeyGenerateCommand) generateRandomKey() string {
	return str.Random(32)
}

// writeNewEnvironmentFileWith Write a new environment file with the given key.
func (receiver *KeyGenerateCommand) writeNewEnvironmentFileWith(key string) error {
	content, err := os.ReadFile(support.EnvPath)
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(content), "APP_KEY="+receiver.config.GetString("app.key"), "APP_KEY="+key, 1)

	err = os.WriteFile(support.EnvPath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
