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
		// Entries and assets used outside resources/views, such as the Inertia root
		// template. ` + "`vite:inputs`" + ` merges them with the calls it finds in views.
		// Inputs are the entries built by ` + "`vite build`" + `, assets are the files passed
		// to ` + "`viteAsset`" + `. Paths are relative to the project root, for example:
		//
		//   "inputs": []string{"resources/inertia/app.ts", "resources/js/app.js"},
		//   "assets": []string{"resources/images/logo.svg"},
		"inputs": []string{},
		"assets": []string{},

		// Dev URL
		//
		// The Vite dev server URL, used when the hot file is absent; usually set through
		// the VITE_DEV_URL environment variable. The dev server writes the hot file when
		// it starts and removes it when it stops, so it is absent whenever the dev server
		// is not running, such as in production or a build-only environment. It is also
		// absent when the dev server runs somewhere the application cannot read that file
		// from, such as another container or a remote machine; set this to the server's
		// address, for example "http://localhost:5173", so the tags still reach it. The
		// hot file wins when both are present. Leave it empty in production so the tags
		// come from the build manifest.
		"dev_url": config.Env("VITE_DEV_URL", ""),
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
