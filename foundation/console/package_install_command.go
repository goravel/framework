package console

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type PackageInstallCommand struct {
}

func NewPackageInstallCommand() *PackageInstallCommand {
	return &PackageInstallCommand{}
}

// Signature The name and signature of the console command.
func (r *PackageInstallCommand) Signature() string {
	return "package:install"
}

// Description The console command description.
func (r *PackageInstallCommand) Description() string {
	return "Install a package"
}

// Extend The console command extend.
func (r *PackageInstallCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <package@version>",
		Category:  "package",
	}
}

// Handle Execute the console command.
func (r *PackageInstallCommand) Handle(ctx console.Context) error {
	var (
		err error
		pkg = ctx.Argument(0)
	)
	if pkg == "" {
		pkg, err = ctx.Ask("Enter the package name to install", console.AskOption{
			Description: "If no version is specified, install the latest",
			Placeholder: " E.g example.com/pkg or example.com/pkg@v1.0.0",
			Prompt:      ">",
			Validate: func(s string) error {
				if s == "" {
					return errors.CommandEmptyPackageName
				}

				return nil
			},
		})
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}

	// handle package version
	pkg, version, ok := strings.Cut(pkg, "@")
	manager := pkg + "/manager"
	if ok {
		pkg = pkg + "@" + version
	}

	// get package
	var output []byte
	get := exec.Command("go", "get", pkg)
	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(get.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = get.CombinedOutput()

			return err
		},
	}); err != nil {
		color.Errorf("failed to get package: %s\n", err.Error())
		if len(output) > 0 {
			color.Red().Println(string(output))
		}

		return nil
	}

	// install package
	install := exec.Command("go", "run", manager, "install")
	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(install.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = install.CombinedOutput()

			return err
		},
	}); err != nil {
		color.Errorf("failed to install package: %s\n", err.Error())
		if len(output) > 0 {
			color.Red().Println(string(output))
		}

		return nil
	}

	// tidy go.mod file
	tidy := exec.Command("go", "mod", "tidy")
	if err = ctx.Spinner(fmt.Sprintf("> @%s", strings.Join(tidy.Args, " ")), console.SpinnerOption{
		Action: func() error {
			output, err = tidy.CombinedOutput()

			return err
		},
	}); err != nil {
		color.Errorf("failed to tidy go.mod file: %s\n", err.Error())
		if len(output) > 0 {
			color.Red().Println(string(output))
		}

		return nil
	}

	color.Successf("Package %s installed successfully\n", pkg)

	return nil
}
