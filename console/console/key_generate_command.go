package console

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/gookit/color"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	"github.com/urfave/cli/v2"
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
func (receiver *KeyGenerateCommand) Extend() consolecontract.CommandExtend {
	return consolecontract.CommandExtend{
		Category: "key",
	}
}

//Handle Execute the console command.
func (receiver *KeyGenerateCommand) Handle(c *cli.Context) error {
	key := receiver.generateRandomKey()

	if err := receiver.setKeyInEnvironmentFile(key); err != nil {
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

//setKeyInEnvironmentFile Set the application key in the environment file.
func (receiver *KeyGenerateCommand) setKeyInEnvironmentFile(key string) error {
	currentKey := facades.Config.GetString("app.key")

	if currentKey != "" {
		return errors.New("Exist application key")
	}

	err := receiver.writeNewEnvironmentFileWith(key)

	if err != nil {
		return err
	}

	return nil
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
