package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("ai", map[string]any{
		// Default AI Provider
		//
		// This option controls the default AI provider that will be used.
		"default": "",

		// AI Providers
		//
		// Here you may configure each AI provider used by your application.
		// A variety of drivers are available, and each provider may also
		// configure the models available to your application.
		"providers": map[string]any{
			"openai": map[string]any{
				"key": "",
				"url": "",
				"via": "",
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

func (s Stubs) AIFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/ai"
)

func AI() ai.AI {
	return App().MakeAI()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
