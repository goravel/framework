package console

import (
	"sort"
	"strings"

	foundationconsole "github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type Stubs struct{}

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

// Module is module.go: full implements every capability of an application module
// (not Located), the unused ones empty.
func (r *moduleStubs) Module(full bool) string {
	if !full {
		return r.replace(`package DummyPackage

type Module struct{}

func (m *Module) Name() string {
	return "DummyName"
}
`)
	}

	return r.replace(`package DummyPackage

import (
	"embed"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"

	"DummyModule/providers"
	"DummyModule/routes"
)

//go:embed resources/views
var views embed.FS

//go:embed lang
var lang embed.FS

type Module struct{}

func (m *Module) Name() string {
	return "DummyName"
}

// Dependencies are module names registered before this one.
func (m *Module) Dependencies() []string {
	return []string{}
}

func (m *Module) ServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ServiceProvider{},
	}
}

// Config is available as "modules.DummyName", or through Config().Module("DummyName").
func (m *Module) Config() map[string]any {
	return map[string]any{
		"enabled": true,
	}
}

func (m *Module) GlobalMiddleware() []http.Middleware {
	return []http.Middleware{}
}

func (m *Module) Routes() {
	routes.Web()
}

// LateRoutes run after every module's Routes(), e.g. for a catch-all route.
func (m *Module) LateRoutes() {
}

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{}
}

func (m *Module) Seeders() []seeder.Seeder {
	return []seeder.Seeder{}
}

func (m *Module) Commands() []console.Command {
	return []console.Command{}
}

func (m *Module) Jobs() []queue.Job {
	return []queue.Job{}
}

func (m *Module) Events() map[event.Event][]event.Listener {
	return map[event.Event][]event.Listener{}
}

func (m *Module) Schedule() []schedule.Event {
	return []schedule.Event{}
}

// Views are named "DummyName/<file>", e.g. "DummyName/index.tmpl".
func (m *Module) Views() module.Files {
	return module.Files{FS: views}
}

// Translations are grouped as "DummyName/<file>", e.g. "DummyName/messages.title".
func (m *Module) Translations() module.Files {
	return module.Files{FS: lang}
}

// Frontend is what Vite builds: Entries like "resources/js/app.js" and Aliases like
// "@DummyName-assets": "resources/assets", both relative to the module directory.
func (m *Module) Frontend() module.Frontend {
	return module.Frontend{}
}

// Static serves files as-is under a URL prefix, e.g. "/DummyName/assets": {FS: public, Dir: "public"}.
func (m *Module) Static() map[string]module.Files {
	return map[string]module.Files{}
}
`)
}

// ServiceProvider is the make:provider stub, as providers.ServiceProvider.
func (r *moduleStubs) ServiceProvider() string {
	return strings.NewReplacer(
		"DummyServiceProvider", "ServiceProvider",
		"DummyPackage", "providers",
	).Replace(foundationconsole.Stubs{}.ServiceProvider())
}

// Routes is routes/web.go, registering with the Route facade.
func (r *moduleStubs) Routes() string {
	facadesImport, facadesPackage := routeFacade()
	controllers := r.module + "/http/controllers"
	// The framework's facades form their own import group, the app's sort with the module.
	imports := "\t\"" + facadesImport + "\"\n\n\t\"" + controllers + "\"\n"
	if facadesImport != "github.com/goravel/framework/facades" {
		sorted := []string{facadesImport, controllers}
		sort.Strings(sorted)
		imports = "\t\"" + sorted[0] + "\"\n\t\"" + sorted[1] + "\"\n"
	}

	content := "package routes\n\nimport (\n" + imports + `)

func Web() {
	DummyFacadesPackage.Route().Get("/DummyName", controllers.NewDummyStudlyController().Index)
}
`

	return r.replace(strings.ReplaceAll(content, "DummyFacadesPackage", facadesPackage))
}

// Controller has Index rendering the module view.
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
` + index)
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
