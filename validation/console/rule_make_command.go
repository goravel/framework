package console

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type RuleMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *RuleMakeCommand) Signature() string {
	return "make:rule"
}

// Description The console command description.
func (r *RuleMakeCommand) Description() string {
	return "Create a new rule class"
}

// Extend The console command extend.
func (r *RuleMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the rule even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *RuleMakeCommand) Handle(ctx console.Context) error {
	m, err := supportconsole.NewMake(ctx, "rule", ctx.Argument(0), support.Config.Paths.Rule)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := file.PutContent(m.GetFilePath(), r.populateStub(r.getStub(), m.GetPackageName(), m.GetStructName(), m.GetSignature())); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Rule created successfully")

	if err = modify.GoFile(filepath.Join("app", "providers", "validation_service_provider.go")).
		Find(match.Imports()).Modify(modify.AddImport(m.GetPackageImportPath())).
		Find(match.ValidationRules()).Modify(modify.Register(fmt.Sprintf("&%s.%s{}", m.GetPackageName(), m.GetStructName()))).
		Apply(); err != nil {
		ctx.Warning(errors.ValidationRuleRegisterFailed.Args(err).Error())
		return nil
	}

	ctx.Success("Rule registered successfully")

	return nil
}

func (r *RuleMakeCommand) getStub() string {
	return Stubs{}.Rule()
}

// populateStub Populate the place-holders in the command stub.
func (r *RuleMakeCommand) populateStub(stub string, packageName, structName, signature string) string {
	stub = strings.ReplaceAll(stub, "DummyRule", structName)
	stub = strings.ReplaceAll(stub, "DummySignature", str.Of(signature).Snake().String())
	stub = strings.ReplaceAll(stub, "DummyPackage", packageName)

	return stub
}
