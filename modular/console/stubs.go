package console

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type Stubs struct{}

func (r Stubs) Config(pkg, facadesImport, facadesPackage string) string {
	content := `package DummyPackage

import (
	"DummyFacadesImport"
)

func init() {
	config := DummyFacadesPackage.Config()
	config.Add("modular", map[string]any{
		// Directory of the application's own modules, relative to the project root.
		"path": "modules",
	})
}
`

	return strings.NewReplacer(
		"DummyPackage", pkg,
		"DummyFacadesImport", facadesImport,
		"DummyFacadesPackage", facadesPackage,
	).Replace(content)
}

// routeFacade returns the facades package exposing Route(): the app's when it
// has one, else the framework's.
func routeFacade() (importPath, pkg string) {
	if file.Exists(path.Facade("route.go")) {
		facades := packages.Paths().Facades()

		return facades.Import(), facades.Package()
	}

	return "github.com/goravel/framework/facades", "facades"
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

// Module is module.go: full adds the provider, routes, config, views,
// translations and Vite entries.
func (r *moduleStubs) Module(full bool) string {
	var imports, vars, methods []string

	if full {
		imports = append(imports, "\t\"embed\"\n\t\"io/fs\"\n", `	"github.com/goravel/framework/contracts/foundation"
`, `	"DummyModule/providers"
	"DummyModule/routes"
`)
		vars = append(vars, "//go:embed resources/views\nvar views embed.FS\n\n", "//go:embed lang\nvar lang embed.FS\n\n")
		methods = append(methods, `
func (m *Module) ServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ServiceProvider{},
	}
}

func (m *Module) Routes() {
	routes.Web()
	routes.Api()
}

// Config is available as "modules.DummyName".
func (m *Module) Config() map[string]any {
	return map[string]any{
		"enabled": true,
	}
}

// Views are named "DummyName/<file>", e.g. "DummyName/index.tmpl".
func (m *Module) Views() fs.FS {
	sub, _ := fs.Sub(views, "resources/views")

	return sub
}

// Translations are grouped as "DummyName/<file>", e.g. "DummyName/messages.title".
func (m *Module) Translations() fs.FS {
	sub, _ := fs.Sub(lang, "lang")

	return sub
}

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
	return r.routes(`func Web() {
	DummyFacadesPackage.Route().Get("/DummyName", controllers.NewDummyStudlyController().Index)
}
`)
}

func (r *moduleStubs) ApiRoutes() string {
	return r.routes(`func Api() {
	DummyFacadesPackage.Route().Prefix("api").Get("/DummyName", controllers.NewDummyStudlyController().Show)
}
`)
}

// routes is a routes file registering with the Route facade.
func (r *moduleStubs) routes(body string) string {
	facadesImport, facadesPackage := routeFacade()
	controllers := r.module + "/http/controllers"
	// The framework's facades form their own import group, the app's sort with the module.
	imports := "\t\"" + facadesImport + "\"\n\n\t\"" + controllers + "\"\n"
	if facadesImport != "github.com/goravel/framework/facades" {
		sorted := []string{facadesImport, controllers}
		sort.Strings(sorted)
		imports = "\t\"" + sorted[0] + "\"\n\t\"" + sorted[1] + "\"\n"
	}

	content := "package routes\n\nimport (\n" + imports + ")\n\n" + body

	return r.replace(strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage))
}

// Controller has Index rendering the module view and Show for the API route.
func (r *moduleStubs) Controller() string {
	imports := "import (\n\t\"github.com/goravel/framework/contracts/http\"\n)\n"
	title := `"DummyTitle"`
	if facadesImport, facadesPackage, ok := langFacade(); ok {
		imports = "import (\n\t\"github.com/goravel/framework/contracts/http\"\n\n\t\"" + facadesImport + "\"\n)\n"
		title = facadesPackage + `.Lang(ctx).Get("DummyName/messages.title")`
	}
	index := `
func (r *DummyStudlyController) Index(ctx http.Context) http.Response {
	return ctx.Response().View().Make("DummyName/index.tmpl", map[string]any{
		"title": ` + title + `,
	})
}
`

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
