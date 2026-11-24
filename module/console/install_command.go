package console

import (
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
)

type InstallCommand struct {
	process process.Process
}

func NewInstallCommand(process process.Process) *InstallCommand {
	return &InstallCommand{process: process}
}

// Signature The name and signature of the console command.
func (r *InstallCommand) Signature() string {
	return "module:install"
}

// Description The console command description.
func (r *InstallCommand) Description() string {
	return "Install third-party modules"
}

// Extend The console command extend.
func (r *InstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <package@version>...",
		Category:  "module",
	}
}

// Handle Execute the console command: go get, register in bootstrap/modules.go, publish.
func (r *InstallCommand) Handle(ctx console.Context) error {
	if !env.IsBootstrapSetup() {
		return errors.ModuleRequiresBootstrapSetup.Args(r.Signature())
	}

	packages := ctx.Arguments()
	if len(packages) == 0 {
		return errors.ModuleInstallPackageEmpty
	}

	for _, pkg := range packages {
		if err := r.install(pkg); err != nil {
			return err
		}
		ctx.Success("Module installed successfully: " + pkg)
	}

	if res := r.process.Run("go", "mod", "tidy"); res.Failed() {
		return errors.ModuleInstallFailed.Args("go mod tidy", res.Error())
	}

	// The running binary does not have the modules: a new one publishes them.
	res := r.process.WithSpinner("Publishing modules").Run("go", "run", ".", "artisan", "module:publish")
	if output := strings.TrimSpace(res.Output()); output != "" {
		ctx.Line(output)
	}
	if res.Failed() {
		return errors.ModuleInstallFailed.Args("module:publish", res.Error())
	}

	return nil
}

// install gets pkg, the Go package holding the Module struct, and lists &<name>.Module{}.
func (r *InstallCommand) install(pkg string) error {
	importPath, _, _ := strings.Cut(pkg, "@")

	if res := r.process.WithSpinner("Getting "+pkg).Run("go", "get", pkg); res.Failed() {
		return errors.ModuleInstallFailed.Args("go get "+pkg, res.Error())
	}

	res := r.process.Run("go", "list", "-f", "{{.Name}}", importPath)
	if res.Failed() {
		return errors.ModuleInstallFailed.Args("go list "+importPath, res.Error())
	}
	name := strings.TrimSpace(res.Output())

	registered, err := moduleRegistered(importPath, name, importPath)
	if err != nil || registered {
		return err
	}

	spec := importPath
	if name != importPath[strings.LastIndex(importPath, "/")+1:] {
		spec = name + " " + importPath
	}

	return modify.AddModule(spec, "&"+name+".Module{}")
}
