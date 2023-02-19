package console

import (
	"io/ioutil"
	"strings"

	"github.com/gookit/color"
	"github.com/manifoldco/promptui"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
)

type KeyGenerateCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *KeyGenerateCommand) Signature() string {
	return "key:generate"
}

//Description The console command description.
func (receiver *KeyGenerateCommand) Description() string {
	return "Set the application key"
}

//Extend The console command extend.
func (receiver *KeyGenerateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "key",
	}
}

//Handle Execute the console command.
func (receiver *KeyGenerateCommand) Handle(ctx console.Context) error {
	if facades.Config.GetString("app.env") == "production" {
		color.Yellowln("**************************************")
		color.Yellowln("*     Application In Production!     *")
		color.Yellowln("**************************************")
		prompt := promptui.Prompt{
			Label: color.New(color.Green).Sprintf("Do you really wish to run this command?(yes/no)") + "[" + color.New(color.Yellow).Sprintf("no") + "]",
		}
		result, err := prompt.Run()
		if err != nil {
			color.Redln(err.Error())

			return nil
		}
		if result != "yes" {
			color.Yellowln("Command Canceled")

			return nil
		}
	}

	key := receiver.generateRandomKey()
	if err := receiver.writeNewEnvironmentFileWith(key); err != nil {
		color.Redln(err.Error())

		return nil
	}

	color.Greenln("Application key set successfully")

	return nil
}

//generateRandomKey Generate a random key for the application.
func (receiver *KeyGenerateCommand) generateRandomKey() string {
	return str.Random(32)
}

//writeNewEnvironmentFileWith Write a new environment file with the given key.
func (receiver *KeyGenerateCommand) writeNewEnvironmentFileWith(key string) error {
	content, err := ioutil.ReadFile(".env")
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(content), "APP_KEY="+facades.Config.GetString("app.key"), "APP_KEY="+key, 1)

	err = ioutil.WriteFile(".env", []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
