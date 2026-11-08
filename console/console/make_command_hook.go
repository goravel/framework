package console

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type MakeCommandHook struct {
}

func NewMakeCommandHook() *MakeCommandHook {
	return &MakeCommandHook{}
}

// Signature The name and signature of the console command.
func (r *MakeCommandHook) Signature() string {
	return "make:command-hook"
}

// Description The console command description.
func (r *MakeCommandHook) Description() string {
	return "Create a new Artisan command hook"
}

// Extend The console command extend.
func (r *MakeCommandHook) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:               "force",
				Aliases:            []string{"f"},
				Value:              false,
				Usage:              "Create the command hook even if it already exists",
				DisableDefaultText: true,
			},
			&command.StringFlag{
				Name:    "command",
				Aliases: []string{"c"},
				Usage:   "The command signature or '*' pattern the hook runs for, default is '*'",
			},
		},
	}
}

// Handle Execute the console command.
func (r *MakeCommandHook) Handle(ctx console.Context) error {
	commandKey := ctx.Option("command")

	supportconsole.HandleEach(ctx, func(ctx console.Context, name string) error {
		return r.makeOne(ctx, name, commandKey)
	})

	return nil
}

func (r *MakeCommandHook) makeOne(ctx console.Context, name, commandKey string) error {
	make, err := supportconsole.NewMake(ctx, "command hook", name, support.Config.Paths.Hooks)
	if err != nil {
		return err
	}

	commandKey, err = r.commandKey(ctx, commandKey)
	if err != nil {
		return err
	}

	if err := file.PutContent(make.GetFilePath(), r.populateStub(r.getStub(), make.GetPackageName(), make.GetStructName())); err != nil {
		return err
	}

	ctx.Success("Command hook created successfully")

	if !env.IsBootstrapSetup() {
		ctx.Warning("Command hook created but not registered, add it to WithCommandHooks in bootstrap/app.go manually")

		return nil
	}

	hook := fmt.Sprintf("&%s.%s{}", make.GetPackageName(), make.GetStructName())
	if err := modify.AddCommandHook(make.GetPackageImportPath(), commandKey, hook); err != nil {
		return errors.ConsoleCommandHookRegisterFailed.Args(make.GetSignature(), err)
	}

	ctx.Success("Command hook registered successfully")

	return nil
}

// commandKey returns the key given with --command, otherwise it prompts for one
// with "*" as the default.
func (r *MakeCommandHook) commandKey(ctx console.Context, key string) (string, error) {
	if key = strings.TrimSpace(key); key != "" {
		return key, nil
	}

	key, err := ctx.Ask("Enter the command signature the hook runs for", console.AskOption{Default: "*"})
	if err != nil {
		return "", err
	}

	if key = strings.TrimSpace(key); key == "" {
		key = "*"
	}

	return key, nil
}

func (r *MakeCommandHook) getStub() string {
	return Stubs{}.CommandHook()
}

// populateStub Populate the place-holders in the command hook stub.
func (r *MakeCommandHook) populateStub(stub, packageName, structName string) string {
	stub = strings.ReplaceAll(stub, "DummyCommandHook", structName)
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
