package console

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/console"
	console2 "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/support/facades"
	"github.com/goravel/framework/support/str"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"strings"
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
func (receiver *KeyGenerateCommand) Extend() console2.CommandExtend {
	return console2.CommandExtend{
		Category: "key",
	}
}

//Handle Execute the console command.
func (receiver *KeyGenerateCommand) Handle(c *cli.Context) error {
	key := receiver.generateRandomKey()

	if err := receiver.setKeyInEnvironmentFile(key); err != nil {
		return err
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
		log.Fatalln("Exist application key.")
	}

	err := receiver.writeNewEnvironmentFileWith(key)

	if err != nil {
		return err
	}

	return nil
}

//writeNewEnvironmentFileWith Write a new environment file with the given key.
func (receiver *KeyGenerateCommand) writeNewEnvironmentFileWith(key string) error {
	content, err := ioutil.ReadFile(console.EnvironmentFile)
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(content), "APP_KEY="+facades.Config.GetString("app.key"), "APP_KEY="+key, 1)

	err = ioutil.WriteFile(console.EnvironmentFile, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
