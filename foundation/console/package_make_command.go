package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type PackageMakeCommand struct{}

func NewPackageMakeCommand() *PackageMakeCommand {
	return &PackageMakeCommand{}
}

// Signature The name and signature of the console command.
func (r *PackageMakeCommand) Signature() string {
	return "make:package"
}

// Description The console command description.
func (r *PackageMakeCommand) Description() string {
	return "Create a package template"
}

// Extend The console command extend.
func (r *PackageMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "root",
				Aliases: []string{"r"},
				Usage:   "The root path of package, default: packages",
				Value:   support.Config.Paths.Packages,
			},
		},
	}
}

// Handle Execute the console command.
func (r *PackageMakeCommand) Handle(ctx console.Context) error {
	for _, pkg := range supportconsole.MakeNames(ctx) {
		if err := r.makeOne(ctx, pkg); err != nil {
			ctx.Error(err.Error())
		}
	}

	return nil
}

func (r *PackageMakeCommand) makeOne(ctx console.Context, pkg string) error {
	if pkg == "" {
		var err error
		pkg, err = ctx.Ask("Enter the package name", console.AskOption{
			Validate: func(s string) error {
				if s == "" {
					return errors.CommandEmptyPackageName
				}

				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	pkg = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(pkg, "/", "_"), "-", "_"), ".", "_")
	root := ctx.Option("root") + "/" + pkg

	if file.Exists(root) {
		ctx.Error("Package " + pkg + " already exists")
		return nil
	}

	packageName := packageName(pkg)
	packageMakeCommandStubs := NewPackageMakeCommandStubs(pkg, root)
	files := map[string]func() string{
		"README.md":           packageMakeCommandStubs.Readme,
		"service_provider.go": packageMakeCommandStubs.ServiceProvider,
		packageName + ".go":   packageMakeCommandStubs.Main,
		filepath.Join("contracts", packageName+".go"): packageMakeCommandStubs.Contracts,
		filepath.Join("facades", packageName+".go"):   packageMakeCommandStubs.Facades,
	}
	if env.IsBootstrapSetup() {
		files[filepath.Join("setup", "stubs.go")] = packageMakeCommandStubs.Config
		files[filepath.Join("setup", "setup.go")] = packageMakeCommandStubs.Setup
	} else {
		files[filepath.Join("setup", "config", packageName+".go")] = packageMakeCommandStubs.OldConfig
		files[filepath.Join("setup", "setup.go")] = packageMakeCommandStubs.OldSetup
	}

	for path, content := range files {
		if err := file.PutContent(filepath.Join(root, path), content()); err != nil {
			return err
		}
	}

	ctx.Success("Package created successfully: " + root)

	return nil
}

func packageName(name string) string {
	nameSlice := strings.Split(name, "/")
	lastName := nameSlice[len(nameSlice)-1]

	return strings.ReplaceAll(strings.ReplaceAll(lastName, "-", "_"), ".", "_")
}
