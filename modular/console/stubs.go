package console

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

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

// langFacade returns the facades package exposing Lang(ctx): the app's, the
// framework's when the app has no facades, or ok=false when the app's lacks it.
func langFacade() (importPath, pkg string, ok bool) {
	if file.Exists(path.Facade("lang.go")) {
		facades := packages.Paths().Facades()

		return facades.Import(), facades.Package(), true
	}
	if file.Exists(path.Facade()) {
		return "", "", false
	}

	return "github.com/goravel/framework/facades", "facades", true
}

func (r *moduleStubs) replace(content string) string {
	return strings.NewReplacer(
		"DummyModule", r.module,
		"DummyPackage", r.pkg,
		"DummyStudly", r.studly,
		"DummySnake", r.snake,
		"DummyTitle", r.title,
		"DummyName", r.kebab,
	).Replace(content)
}

// Module is module.go: web adds views, web routes and Vite entries, lang the
// translations, full the provider, API routes and config.
func (r *moduleStubs) Module(web, lang, full bool) string {
	var imports, vars, methods []string

	if web || lang {
		imports = append(imports, "\t\"embed\"\n\t\"io/fs\"\n")
	}
	if web {
		vars = append(vars, "//go:embed resources/views\nvar views embed.FS\n\n")
	}
	if lang {
		vars = append(vars, "//go:embed lang\nvar lang embed.FS\n\n")
	}
	if full {
		imports = append(imports, `	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/route"
`, `	"DummyModule/providers"
	"DummyModule/routes"
`)
		methods = append(methods, `
func (m *Module) ServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ServiceProvider{},
	}
}
`)
	}
	if web {
		methods = append(methods, `
func (m *Module) Routes(router route.Router) {
	routes.Web(router)
}
`)
	}
	if full {
		methods = append(methods, `
func (m *Module) ApiRoutes(router route.Router) {
	routes.Api(router)
}

// Config is available as "modules.DummyName".
func (m *Module) Config() map[string]any {
	return map[string]any{
		"enabled": true,
	}
}
`)
	}
	if web {
		methods = append(methods, `
// Views are named "DummyName/<file>", e.g. "DummyName/index.tmpl".
func (m *Module) Views() fs.FS {
	sub, _ := fs.Sub(views, "resources/views")

	return sub
}
`)
	}
	if lang {
		methods = append(methods, `
// Translations are grouped as "DummyName/<file>", e.g. "DummyName/messages.title".
func (m *Module) Translations() fs.FS {
	sub, _ := fs.Sub(lang, "lang")

	return sub
}
`)
	}
	if web {
		methods = append(methods, `
// ViteEntries are relative to the module directory, e.g. "resources/js/app.js".
// ViteAliases() and Static() are optional too, see contracts/modular.
func (m *Module) ViteEntries() []string {
	return []string{}
}
`)
	}
	if r.order != 0 {
		methods = append(methods, fmt.Sprintf(`
func (m *Module) Order() int {
	return %d
}
`, r.order))
	}

	content := "package DummyPackage\n\n"
	if len(imports) > 0 {
		content += "import (\n" + strings.Join(imports, "\n") + ")\n\n"
	}
	content += strings.Join(vars, "") + `type Module struct{}

func (m *Module) Name() string {
	return "DummyName"
}
` + strings.Join(methods, "")

	return r.replace(content)
}

func (r *moduleStubs) ServiceProvider() string {
	return `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
)

type ServiceProvider struct{}

// Register service bindings here.
func (r *ServiceProvider) Register(app foundation.Application) {
}

// Boot runs after all service providers have been registered.
func (r *ServiceProvider) Boot(app foundation.Application) {
}
`
}

func (r *moduleStubs) WebRoutes() string {
	return r.replace(`package routes

import (
	"github.com/goravel/framework/contracts/route"

	"DummyModule/http/controllers"
)

func Web(router route.Router) {
	router.Get("/DummyName", controllers.NewDummyStudlyController().Index)
}
`)
}

func (r *moduleStubs) ApiRoutes() string {
	return r.replace(`package routes

import (
	"github.com/goravel/framework/contracts/route"

	"DummyModule/http/controllers"
)

func Api(router route.Router) {
	router.Get("/DummyName", controllers.NewDummyStudlyController().Show)
}
`)
}

// Controller has Show for the API route and, with web, Index rendering the module view.
func (r *moduleStubs) Controller(web bool) string {
	imports := "import (\n\t\"github.com/goravel/framework/contracts/http\"\n)\n"
	index := ""
	if web {
		title := `"DummyTitle"`
		if facadesImport, facadesPackage, ok := langFacade(); ok {
			imports = "import (\n\t\"github.com/goravel/framework/contracts/http\"\n\n\t\"" + facadesImport + "\"\n)\n"
			title = facadesPackage + `.Lang(ctx).Get("DummyName/messages.title")`
		}
		index = `
func (r *DummyStudlyController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("DummyName/index.tmpl", map[string]any{
		"title": ` + title + `,
	})
}
`
	}

	return r.replace(`package controllers

` + imports + `
type DummyStudlyController struct{}

func NewDummyStudlyController() *DummyStudlyController {
	return &DummyStudlyController{}
}
` + index + `
func (r *DummyStudlyController) Show(ctx http.Context) http.Response {
	return ctx.Response().Success().Json(http.Json{
		"module": "DummyName",
	})
}
`)
}

// View has no {{ define }}: define names are global and would collide across modules.
func (r *moduleStubs) View() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .title }}</title>
</head>
<body>
    <h1>{{ .title }}</h1>
</body>
</html>
`
}

func (r *moduleStubs) Lang() string {
	return r.replace(`{
    "title": "DummyTitle"
}
`)
}

func (r *moduleStubs) Test() string {
	tests := packages.Paths().Tests()

	return r.replace(strings.NewReplacer(
		"DummyTestsImport", tests.Import(),
		"DummyTestsPackage", tests.Package(),
	).Replace(`package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"DummyTestsImport"
)

type DummyStudlyTestSuite struct {
	suite.Suite
	DummyTestsPackage.TestCase
}

func TestDummyStudlyTestSuite(t *testing.T) {
	suite.Run(t, new(DummyStudlyTestSuite))
}

func (s *DummyStudlyTestSuite) TestIndex() {
	// TODO
}
`))
}
