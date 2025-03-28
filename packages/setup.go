package packages

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support/color"
)

type setup struct {
	Force       bool
	Module      string
	command     string
	onInstall   []packages.FileModifier
	onUninstall []packages.FileModifier
}

func Setup(args []string) *setup {
	st := &setup{}

	if len(args) > 1 && (args[1] == "install" || args[1] == "uninstall") {
		st.command = args[1]
	}

	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "setup") {
		st.Module = filepath.Dir(info.Path)
	}

	st.Force = len(args) == 3 && (args[2] == "--force" || args[2] == "-f")

	return st
}

func (r *setup) Install(modifiers ...packages.FileModifier) {
	r.onInstall = modifiers
}

func (r *setup) Uninstall(modifiers ...packages.FileModifier) {
	r.onUninstall = modifiers
}

func (r *setup) Execute() {
	if r.Module == "" {
		color.Errorln("package module name is empty, please run command with module name.")
		os.Exit(1)
	}

	if r.command == "install" {
		for i := range r.onInstall {
			r.reportError(r.onInstall[i].Apply())
		}

		color.Successln("Package installed successfully")
	}

	if r.command == "uninstall" {
		for i := range r.onUninstall {
			r.reportError(r.onUninstall[i].Apply())
		}

		color.Successln("Package uninstalled successfully")
	}

}

func (r *setup) reportError(err error) {
	if err != nil {
		if r.Force {
			color.Warningln(err)
			return
		}

		color.Errorln(err)
		os.Exit(1)
	}
}
