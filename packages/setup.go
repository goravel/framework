package packages

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/contracts/packages/modify"
	"github.com/goravel/framework/packages/options"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
)

type setup struct {
	command     string
	driver      string
	facade      string
	force       bool
	onInstall   []modify.Apply
	onUninstall []modify.Apply
	packageName string
}

var osExit = os.Exit

func Setup(args []string) packages.Setup {
	st := &setup{}
	var packageName string

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
		if strings.HasPrefix(arg, "--driver=") {
			st.driver = strings.TrimPrefix(arg, "--driver=")
		}
		if strings.HasPrefix(arg, "--package-name=") {
			packageName = strings.TrimPrefix(arg, "--package-name=")
		}
		if strings.HasPrefix(arg, "--paths=") {
			if err := json.Unmarshal([]byte(strings.TrimPrefix(arg, "--paths=")), &support.Config.Paths); err != nil {
				panic(fmt.Sprintf("failed to unmarshal paths: %s", err))
			}
		}
	}

	if packageName == "" {
		packageName = "goravel"
	}

	st.packageName = packageName

	return st
}

func (r *setup) Execute() {
	if r.command == "install" {
		for i := range r.onInstall {
			r.reportError(r.onInstall[i].Apply(options.Driver(r.driver), options.Force(r.force), options.Facade(r.facade)))
		}

		color.Successln("package installed successfully")
	}

	if r.command == "uninstall" {
		for i := range r.onUninstall {
			r.reportError(r.onUninstall[i].Apply(options.Driver(r.driver), options.Force(r.force), options.Facade(r.facade)))
		}

		color.Successln("package uninstalled successfully")
	}
}

// ModulePath returns the module path of package, it may be a sub-package, eg: github.com/goravel/framework/auth.
func (r *setup) ModulePath() string {
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "setup") {
		return path.Dir(info.Path)
	}

	return ""
}

func (r *setup) PackageName() string {
	return r.packageName
}

func (r *setup) Install(modifiers ...modify.Apply) packages.Setup {
	r.onInstall = modifiers

	return r
}

func (r *setup) Uninstall(modifiers ...modify.Apply) packages.Setup {
	r.onUninstall = modifiers

	return r
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
