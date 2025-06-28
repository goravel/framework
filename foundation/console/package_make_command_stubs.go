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

// Bindings returns what bindings the service provider will register.
func (r *ServiceProvider) Bindings() []string {
	return []string{Binding}
}

// Dependencies returns what dependencies the service provider needs.
// For example, the Cache package needs the Log facade.
// You can get the binding from package service provider files.
func (r *ServiceProvider) Dependencies() []string {
	return []string{}
}

// ProvideFor returns what services the service provider provides for.
// For example, the Redis package provides services for the Cache facade.
// You can get the binding from package service provider files.
func (r *ServiceProvider) ProvideFor() []string {
	return []string{}
}

// Register registers the service provider.
func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return nil, nil
	})
}

// Boot boots the service provider, will be called after all service providers are registered.
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

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	packages.Setup(os.Args).
		Install(
			modify.File(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.AddProvider("&DummyName.ServiceProvider{}")),
		).
		Uninstall(
			modify.File(path.Config("app.go")).
				Find(match.Providers()).Modify(modify.RemoveProvider("&DummyName.ServiceProvider{}")).
				Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
		).
		Execute()
}

`
	content = strings.ReplaceAll(content, "DummyName", r.name)

	return content
}
