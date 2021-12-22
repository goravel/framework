package console

import (
	"fmt"
	"github.com/goravel/framework/console"
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
func (receiver KeyGenerateCommand) Signature() string {
	return "key:generate"
}

//Description The console command description.
func (receiver KeyGenerateCommand) Description() string {
	return "Set the application key"
}

//Flags Set flags, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#flags
func (receiver KeyGenerateCommand) Flags() []cli.Flag {
	var flags []cli.Flag

	return flags
}

//Subcommands Set Subcommands, document: https://github.com/urfave/cli/blob/master/docs/v2/manual.md#subcommands
func (receiver KeyGenerateCommand) Subcommands() []*cli.Command {
	var subcommands []*cli.Command

	return subcommands
}

//Handle Execute the console command.
func (receiver KeyGenerateCommand) Handle(c *cli.Context) error {
	key := receiver.generateRandomKey()

	if err := receiver.setKeyInEnvironmentFile(key); err != nil {
		return err
	}

	fmt.Println("Application key set successfully.")

	return nil
}

//generateRandomKey Generate a random key for the application.
func (receiver KeyGenerateCommand) generateRandomKey() string {
	return str.Random(32)
}

//setKeyInEnvironmentFile Set the application key in the environment file.
func (receiver KeyGenerateCommand) setKeyInEnvironmentFile(key string) error {
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
func (receiver KeyGenerateCommand) writeNewEnvironmentFileWith(key string) error {
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
