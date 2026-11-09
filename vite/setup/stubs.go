package main

import "strings"

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("vite", map[string]any{
		// Build Directory
		//
		// The directory inside the public directory that ` + "`vite build`" + ` writes to.
		// goravel-vite-plugin reads these paths from the application.
		"build_directory": "build",

		// Hot File
		//
		// The file the dev server announces its URL in, relative to the project
		// root. While it exists, {{ vite }} points every tag at the dev server.
		"hot_file": "public/hot",

		// Manifest
		//
		// The manifest ` + "`vite build`" + ` writes, relative to the build directory.
		"manifest": ".vite/manifest.json",

		// Inputs and Assets
		//
		// Entries and assets used outside views; ` + "`vite:inputs`" + ` merges them with view calls.
		//   "inputs": []string{"resources/inertia/app.ts"},
		//   "assets": []string{"resources/images/logo.svg"},
		"inputs": []string{},
		"assets": []string{},
	})
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyFacadesImport", facadesImport)
	content = strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage)

	return content
}

func (s Stubs) ViteFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/vite"
)

func Vite() vite.Vite {
	return App().MakeVite()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
