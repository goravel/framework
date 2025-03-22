package console

import (
	"strings"

	"github.com/goravel/framework/support/str"
)

type PackageMakeCommandStubs struct {
	pkg  string
	root string
	name string
}

func NewPackageMakeCommandStubs(pkg, root string) *PackageMakeCommandStubs {
	return &PackageMakeCommandStubs{pkg: pkg, root: root, name: packageName(pkg)}
}

func (r PackageMakeCommandStubs) Readme() string {
	content := `# DummyName
`

	return strings.ReplaceAll(content, "DummyName", r.name)
}

func (r PackageMakeCommandStubs) ServiceProvider() string {
	content := `package DummyName

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "DummyPackage"

var App foundation.Application

type ServiceProvider struct {
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return nil, nil
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {

}
`

	content = strings.ReplaceAll(content, "DummyPackage", r.pkg)
	content = strings.ReplaceAll(content, "DummyName", r.name)

	return content
}

func (r PackageMakeCommandStubs) Main() string {
	content := `package DummyName

type DummyCamelName struct {}
`

	content = strings.ReplaceAll(content, "DummyName", r.name)
	content = strings.ReplaceAll(content, "DummyCamelName", str.Of(r.name).Studly().String())

	return content
}

func (r PackageMakeCommandStubs) Config() string {
	content := `package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("DummyName", map[string]any{
		
	})
}
`

	return strings.ReplaceAll(content, "DummyName", r.name)
}

func (r PackageMakeCommandStubs) Contracts() string {
	content := `package contracts

type DummyCamelName interface {}
`

	return strings.ReplaceAll(content, "DummyCamelName", str.Of(r.name).Studly().String())
}

func (r PackageMakeCommandStubs) Facades() string {
	content := `package facades

import (
	"log"

	"goravel/DummyRoot"
	"goravel/DummyRoot/contracts"
)

func DummyCamelName() contracts.DummyCamelName {
	instance, err := DummyName.App.Make(DummyName.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.DummyCamelName)
}
`

	content = strings.ReplaceAll(content, "DummyRoot", r.root)
	content = strings.ReplaceAll(content, "DummyName", r.name)
	content = strings.ReplaceAll(content, "DummyCamelName", str.Of(r.name).Studly().String())

	return content
}

func (r PackageMakeCommandStubs) Setup() string {
	content := `package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	pkgcontracts "github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/path"
)

func main() {
	info, ok := debug.ReadBuildInfo()
	if !ok || !strings.HasSuffix(info.Path, "setup") {
		color.Errorln("Package module name is empty, please run command with module name.")
		return
	}
	module := filepath.Dir(info.Path)
	force := len(os.Args) == 3 && (os.Args[2] == "--force" || os.Args[2] == "-f")

	var pkg = &packages.Setup{
		Force: force,
		Module:          module,
		OnInstall: []pkgcontracts.FileModifier{
			packages.ModifyGoFile{
				File: path.Config("app.go"),
				Modifiers: []pkgcontracts.GoNodeModifier{
					packages.AddImportSpec(module),
					packages.AddProviderSpec(
						"&DummyName.ServiceProvider{}",
					),
				},
			},
		},
		OnUninstall: []pkgcontracts.FileModifier{
			packages.ModifyGoFile{
				File: path.Config("app.go"),
				Modifiers: []pkgcontracts.GoNodeModifier{
					packages.RemoveImportSpec(module),
					packages.RemoveProviderSpec("&DummyName.ServiceProvider{}"),
				},
			},
		},
	}

	if len(os.Args) > 1 {
		execute(pkg, os.Args[1])
	}
}

func execute(pkg pkgcontracts.Setup, command string) {
	var err error
	switch command {
	case "install":
		err = pkg.Install()
	case "uninstall":
		err = pkg.Uninstall()
	default:
		return
	}

	if err != nil {
		color.Errorln(err)
		os.Exit(1)
	}

	color.Successf("Package %sed successfully\n", command)
}
`
	content = strings.ReplaceAll(content, "DummyName", r.name)

	return content
}
