package packages

import (
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/color"
)

type setup struct {
	command     string
	module      string
	onInstall   []modify.Apply
	onUninstall []modify.Apply
	force       bool
}

var osExit = os.Exit

func GetModulePath() string {
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "setup") {
		return path.Dir(info.Path)
	}

	return ""
}

func Setup(args []string) packages.Setup {
	st := &setup{}

	if len(args) > 1 && (args[1] == "install" || args[1] == "uninstall") {
		st.command = args[1]
	}
	st.module = GetModulePath()
	st.force = len(args) == 3 && (args[2] == "--force" || args[2] == "-f")

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
			r.reportError(r.onInstall[i].Apply())
		}

		color.Successln("package installed successfully")
	}

	if r.command == "uninstall" {
		for i := range r.onUninstall {
			r.reportError(r.onUninstall[i].Apply())
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
