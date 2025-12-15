package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/support/path"

	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("filesystems", map[string]any{
		// Default Filesystem Disk
		//
		// Here you may specify the default filesystem disk that should be used
		// by the framework. The "local" disk, as well as a variety of cloud
		// based disks are available to your application. Just store away!
		"default": "local",

		// Filesystem Disks
		//
		// Here you may configure as many filesystem "disks" as you wish, and you
		// may even configure multiple disks of the same driver. Defaults have
		// been set up for each driver as an example of the required values.
		//
		// Supported Drivers: "local", "custom"
		"disks": map[string]any{
			"local": map[string]any{
				"driver": "local",
				"root":   path.Storage("app"),
			},
			"public": map[string]any{
				"driver": "local",
				"root":   path.Storage("app/public"),
				"url":    config.Env("APP_URL", "").(string) + "/storage",
			},
		},
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) StorageFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/filesystem"
)

func Storage() filesystem.Storage {
	return App().MakeStorage()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
