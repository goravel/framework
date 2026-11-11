package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type InitCommand struct {
	config config.Config
}

func NewInitCommand(config config.Config) *InitCommand {
	return &InitCommand{config: config}
}

// Signature The name and signature of the console command.
func (r *InitCommand) Signature() string {
	return "modular:init"
}

// Description The console command description.
func (r *InitCommand) Description() string {
	return "Prepare the application for modules"
}

// Extend The console command extend.
func (r *InitCommand) Extend() command.Extend {
	return command.Extend{
		Category: "modular",
	}
}

// Handle Execute the console command. Running it again only reports what exists.
func (r *InitCommand) Handle(ctx console.Context) error {
	if !env.IsBootstrapSetup() {
		return errors.ModularRequiresBootstrapSetup.Args(r.Signature())
	}

	directory := modulesDirectory(r.config)
	if err := putIfMissing(ctx, filepath.ToSlash(filepath.Join(directory, ".gitkeep")), path.Base(directory, ".gitkeep"), ""); err != nil {
		return err
	}

	facadesImport, facadesPackage := facadesImport()
	configStub := Stubs{}.Config(packages.Paths().Config().Package(), facadesImport, facadesPackage)
	if err := putIfMissing(ctx, packages.Paths().Config().String("modular.go"), path.Config("modular.go"), configStub); err != nil {
		return err
	}

	if err := r.wireModules(ctx); err != nil {
		return err
	}

	return r.addGenerate(ctx)
}

// generateDirective makes go generate publish third-party modules.
const generateDirective = "//go:generate go run . artisan modular:publish"

// addGenerate adds generateDirective to main.go, above func main.
func (r *InitCommand) addGenerate(ctx console.Context) error {
	label := "main.go go:generate modular:publish"
	mainFile := path.Base("main.go")

	content, err := file.GetContent(mainFile)
	if err != nil {
		ctx.TwoColumnDetail(label, statusManual)
		ctx.Warning("Add " + generateDirective + " to the main package")

		return nil
	}
	if strings.Contains(content, generateDirective) {
		ctx.TwoColumnDetail(label, statusExists)

		return nil
	}

	if before, after, found := strings.Cut(content, "\nfunc main("); found {
		content = before + "\n" + generateDirective + "\n\nfunc main(" + after
	} else {
		content = strings.TrimRight(content, "\n") + "\n\n" + generateDirective + "\n"
	}
	if err := file.PutContent(mainFile, content); err != nil {
		return err
	}
	ctx.TwoColumnDetail(label, statusCreated)

	return nil
}

// wireModules creates bootstrap/modules.go and adds WithModules(Modules) to foundation.Setup().
func (r *InitCommand) wireModules(ctx console.Context) error {
	bootstrap := support.Config.Paths.Bootstrap
	modulesLabel := filepath.ToSlash(filepath.Join(bootstrap, "modules.go"))
	appLabel := filepath.ToSlash(filepath.Join(bootstrap, "app.go")) + " WithModules(Modules)"

	if file.Contain(path.Bootstrap("app.go"), "WithModules") {
		if file.Exists(path.Bootstrap("modules.go")) {
			ctx.TwoColumnDetail(modulesLabel, statusExists)
		}
		ctx.TwoColumnDetail(appLabel, statusExists)

		return nil
	}

	err := modify.InitModules()
	if err == nil {
		ctx.TwoColumnDetail(modulesLabel, statusCreated)
		ctx.TwoColumnDetail(appLabel, statusCreated)

		return nil
	}
	if !isManualWiring(err) {
		return err
	}

	// The chain could not be edited: keep modules.go ready and leave app.go to the user.
	if err := putIfMissing(ctx, modulesLabel, path.Bootstrap("modules.go"), modify.ModulesStub()); err != nil {
		return err
	}
	ctx.TwoColumnDetail(appLabel, statusManual)
	ctx.Warning("Add .WithModules(Modules) to foundation.Setup() in " + filepath.ToSlash(filepath.Join(bootstrap, "app.go")))

	return nil
}

func isManualWiring(err error) bool {
	return errors.Is(err, errors.PackageSetupNotFound) ||
		errors.Is(err, errors.PackageModulesFileExists) ||
		errors.Is(err, errors.PackageMatchGoNodeFail) ||
		errors.Is(err, errors.PackageModifyGoFileFail)
}
