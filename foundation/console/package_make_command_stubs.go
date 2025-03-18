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

func (r PackageMakeCommandStubs) Manager() string {
	content := `package main

import (
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
    "strings"

	pkgcontracts "github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/color"
)

var (
	module string
	dir    string
	force  bool
)

func init() {
	for i, arg := range os.Args {
		if arg == "--force" || arg == "-f" {
			force = true
		}

		if (arg == "--dir" || arg == "-d") && len(os.Args) > i+1 {
			dir = os.Args[i+1]
		}
	}

	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "manager") {
		module = path.Dir(info.Path)
	}

	if dir == "" {
		dir, _ = os.Getwd()
	}
}

func main() {
	var pkg = packages.Manager{
		ContinueOnError: force,
		Module:          module,
		OnInstall: []pkgcontracts.FileModifier{
			packages.ModifyGoFile{
				File: filepath.Join("config", "app.go"),
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
				File: filepath.Join("config", "app.go"),
				Modifiers: []pkgcontracts.GoNodeModifier{
					packages.RemoveImportSpec(module),
					packages.RemoveProviderSpec("&DummyName.ServiceProvider{}"),
				},
			},
		},
	}

    if module == "" {
		color.Errorln("Package module name is empty, please run command with module name.")
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "install" {
		err := pkg.Install(dir)
		if err != nil {
			color.Errorln(err)
			return
		}
		color.Successf("Package %s installed successfully\n", module)
	}

	if len(os.Args) > 1 && os.Args[1] == "uninstall" {
		err := pkg.Uninstall(dir)
		if err != nil {
			color.Errorln(err)
			return
		}
		color.Successf("Package %s uninstalled successfully\n", module)
	}
}

`
	content = strings.ReplaceAll(content, "DummyName", r.name)

	return content
}
