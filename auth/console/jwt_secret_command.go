package console

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"

	"github.com/gookit/color"
)

type JwtSecretCommand struct {
}

//Signature The name and signature of the console command.
func (receiver *JwtSecretCommand) Signature() string {
	return "jwt:secret"
}

//Description The console command description.
func (receiver *JwtSecretCommand) Description() string {
	return "Set the JWTAuth secret key used to sign the tokens"
}

//Extend The console command extend.
func (receiver *JwtSecretCommand) Extend() command.Extend {
	return command.Extend{
		Category: "jwt",
	}
}

//Handle Execute the console command.
func (receiver *JwtSecretCommand) Handle(ctx console.Context) error {
	key := receiver.generateRandomKey()

	if err := receiver.setSecretInEnvironmentFile(key); err != nil {
		color.Redln(err.Error())

		return nil
	}

	color.Greenln("Jwt Secret set successfully")

	return nil
}

//generateRandomKey Generate a random key for the application.
func (receiver *JwtSecretCommand) generateRandomKey() string {
	return str.Random(32)
}

//setSecretInEnvironmentFile Set the application key in the environment file.
func (receiver *JwtSecretCommand) setSecretInEnvironmentFile(key string) error {
	currentKey := facades.Config.GetString("jwt.secret")

	if currentKey != "" {
		return errors.New("Exist jwt secret")
	}

	err := receiver.writeNewEnvironmentFileWith(key)

	if err != nil {
		return err
	}

	return nil
}

//writeNewEnvironmentFileWith Write a new environment file with the given key.
func (receiver *JwtSecretCommand) writeNewEnvironmentFileWith(key string) error {
	content, err := ioutil.ReadFile(".env")
	if err != nil {
		return err
	}

	newContent := strings.Replace(string(content), "JWT_SECRET="+facades.Config.GetString("jwt.secret"), "JWT_SECRET="+key, 1)

	err = ioutil.WriteFile(".env", []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
