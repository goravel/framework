package console

import "strings"

type Stubs struct{}

func (r Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/http"

	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("modular", map[string]any{
		// Directory of the application's own modules, relative to the project root.
		"path": "modules",

		// Route groups wrapping what modules register with Routes() and ApiRoutes().
		"routes": map[string]any{
			"web": map[string]any{"prefix": "", "middleware": []http.Middleware{}},
			"api": map[string]any{"prefix": "api", "middleware": []http.Middleware{}},
		},
	})
}
`

	return strings.NewReplacer(
		"DummyPackage", pkg,
		"DummyFacadesImport", facadesImport,
		"DummyFacadesPackage", facadesPackage,
	).Replace(content)
}
