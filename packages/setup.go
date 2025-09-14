package packages

import (
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/options"
	"github.com/goravel/framework/support/color"
)

type setup struct {
	command     string
	facade      string
	module      string
	force       bool
	onInstall   []modify.Apply
	onUninstall []modify.Apply
}

var osExit = os.Exit

// GetModulePath returns the module path of package, it may be a sub-package, eg: github.com/goravel/framework/auth.
func GetModulePath() string {
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "setup") {
		return path.Dir(info.Path)
	}

	return ""
}

// GetModuleName returns the module name of application, eg: goravel.
func GetModuleName() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Path
	}

	return "goravel"
}

// GetModuleNameFromArgs returns the module name from command line arguments, default is "goravel".
// It is used in the package:install command.
func GetModuleNameFromArgs(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--module=") {
			return strings.TrimPrefix(arg, "--module=")
		}
	}

	return "goravel"
}

func Setup(args []string) packages.Setup {
	st := &setup{}

	for _, arg := range args {
		if arg == "install" || arg == "uninstall" {
			st.command = arg
		}
		if arg == "--force" || arg == "-f" {
			st.force = true
		}
		if strings.HasPrefix(arg, "--facade=") {
			st.facade = strings.TrimPrefix(arg, "--facade=")
		}
	}

	st.module = GetModulePath()

	return st
}

func (r *setup) Install(modifiers ...modify.Apply) packages.Setup {
	r.onInstall = modifiers

	return r
}

func (r *setup) Uninstall(modifiers ...modify.Apply) packages.Setup {
	r.onUninstall = modifiers

	return r
}

func (r *setup) Execute() {
	if r.module == "" {
		color.Errorln(errors.PackageModuleNameEmpty)
		osExit(1)
	}

	if r.command == "install" {
		for i := range r.onInstall {
			r.reportError(r.onInstall[i].Apply(options.Force(r.force), options.Facade(r.facade)))
		}

		color.Successln("package installed successfully")
	}

	if r.command == "uninstall" {
		for i := range r.onUninstall {
			r.reportError(r.onUninstall[i].Apply(options.Force(r.force), options.Facade(r.facade)))
		}

		color.Successln("package uninstalled successfully")
	}
}

func (r *setup) reportError(err error) {
	if err != nil {
		if r.force {
			color.Warningln(err)
			return
		}

		color.Errorln(err)
		osExit(1)
	}
}
