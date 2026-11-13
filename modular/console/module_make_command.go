package console

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
)

var (
	// moduleInputPattern is what make:module accepts; the kebab name it derives must pass modular.ValidName.
	moduleInputPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 _-]*$`)

	// reservedModulePackages are packages Go refuses to import.
	reservedModulePackages = map[string]bool{"main": true, "init": true, "internal": true, "vendor": true}
)

type ModuleMakeCommand struct {
	config config.Config
}

func NewModuleMakeCommand(config config.Config) *ModuleMakeCommand {
	return &ModuleMakeCommand{config: config}
}

// Signature The name and signature of the console command.
func (r *ModuleMakeCommand) Signature() string {
	return "make:module"
}

// Description The console command description.
func (r *ModuleMakeCommand) Description() string {
	return "Create a new module"
}

// Extend The console command extend.
func (r *ModuleMakeCommand) Extend() command.Extend {
	return command.Extend{
		ArgsUsage: " <name>...",
		Category:  "make",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "api",
				Usage: "Create an API module, without views and web routes",
			},
			&command.BoolFlag{
				Name:  "plain",
				Usage: "Create only module.go",
			},
			&command.IntFlag{
				Name:  "order",
				Usage: "The module order: lower first, negative last",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the module even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ModuleMakeCommand) Handle(ctx console.Context) error {
	supportconsole.HandleEach(ctx, r.makeOne)

	return nil
}

func (r *ModuleMakeCommand) makeOne(ctx console.Context, name string) error {
	if !env.IsBootstrapSetup() {
		return errors.ModularRequiresBootstrapSetup.Args(r.Signature())
	}
	if ctx.OptionBool("api") && ctx.OptionBool("plain") {
		return errors.ModularOptionsConflict.Args("api", "plain")
	}

	if name == "" {
		var err error
		name, err = ctx.Ask("Enter the module name", console.AskOption{
			Validate: func(s string) error {
				if s == "" {
					return errors.ModularModuleNameEmpty
				}

				return nil
			},
		})
		if err != nil {
			return err
		}
	}

	stubs, err := newModuleStubs(name, modulesDirectory(r.config))
	if err != nil {
		return err
	}
	stubs.order = ctx.OptionInt("order")

	root := path.Base(filepath.FromSlash(stubs.dir))
	if file.Exists(root) && !ctx.OptionBool("force") {
		return errors.ModularModuleExists.Args(stubs.kebab, stubs.dir)
	}

	registered, err := stubs.registered()
	if err != nil {
		return err
	}

	files := stubs.files(ctx.OptionBool("api"), ctx.OptionBool("plain"))
	if err := removeStaleStubs(root, stubs.files(false, false), files); err != nil {
		return err
	}
	for name, content := range files {
		if err := file.PutContent(filepath.Join(root, filepath.FromSlash(name)), content); err != nil {
			return err
		}
	}

	if !registered {
		if err := modify.AddModule(stubs.importSpec(), "&"+stubs.pkg+".Module{}"); err != nil {
			return err
		}
	}

	ctx.Success("Module created successfully: " + stubs.dir)

	// An assigned chain (builder := foundation.Setup()...) cannot be edited: the user wires it.
	if !file.Contain(path.Bootstrap("app.go"), "WithModules") {
		ctx.Warning("Add .WithModules(Modules) to foundation.Setup() in " + filepath.ToSlash(filepath.Join(support.Config.Paths.Bootstrap, "app.go")))
	}

	return nil
}

// removeStaleStubs deletes the stubs of an earlier --force run that this one no longer writes.
func removeStaleStubs(root string, all, current map[string]string) error {
	for name := range all {
		if _, ok := current[name]; ok {
			continue
		}

		stale := filepath.Join(root, filepath.FromSlash(name))
		if err := os.Remove(stale); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}
		// Remove the directories the stub leaves empty; os.Remove refuses the others.
		for dir := filepath.Dir(stale); dir != root && os.Remove(dir) == nil; dir = filepath.Dir(dir) {
		}
	}

	return nil
}

// moduleStubs holds the names derived from a module name, e.g. "BlogPosts".
type moduleStubs struct {
	kebab  string // blog-posts: module name and directory
	pkg    string // blogposts: Go package
	studly string // BlogPosts
	snake  string // blog_posts
	title  string // Blog Posts
	dir    string // modules/blog-posts
	module string // <main>/modules/blog-posts
	order  int
}

func newModuleStubs(name, directory string) (*moduleStubs, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.ModularModuleNameEmpty
	}

	kebab := str.Of(name).Kebab().String()
	pkg := strings.ReplaceAll(kebab, "-", "")
	if !moduleInputPattern.MatchString(name) || !modular.ValidName(kebab) {
		return nil, errors.ModularModuleNameInvalid.Args(name)
	}
	if token.IsKeyword(pkg) || reservedModulePackages[pkg] {
		return nil, errors.ModularModuleNameForbidden.Args(name, pkg)
	}

	dir := filepath.ToSlash(filepath.Join(directory, kebab))

	return &moduleStubs{
		kebab:  kebab,
		pkg:    pkg,
		studly: str.Of(kebab).Studly().String(),
		snake:  strings.ReplaceAll(kebab, "-", "_"),
		title:  str.Of(kebab).Headline().String(),
		dir:    dir,
		module: packages.Paths().Main().Import() + "/" + dir,
	}, nil
}

// importSpec is the import of the module package, aliased when the directory is not the package name.
func (r *moduleStubs) importSpec() string {
	if r.pkg == r.kebab {
		return r.module
	}

	return r.pkg + " " + r.module
}

// registered reports whether bootstrap/modules.go lists the module, and fails
// when another import there already uses the module's package name.
func (r *moduleStubs) registered() (bool, error) {
	modulesFile := path.Bootstrap("modules.go")
	source := modify.ModulesStub()
	if file.Exists(modulesFile) {
		content, err := file.GetContent(modulesFile)
		if err != nil {
			return false, err
		}
		source = content
	}

	parsed, err := parser.ParseFile(token.NewFileSet(), modulesFile, source, parser.ImportsOnly)
	if err != nil {
		return false, err
	}

	imported := false
	for _, spec := range parsed.Imports {
		importPath := strings.Trim(spec.Path.Value, "`\"")
		name := importPath[strings.LastIndex(importPath, "/")+1:]
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if importPath == r.module {
			imported = true
		} else if name == r.pkg {
			return false, errors.ModularModuleImportConflict.Args(r.kebab, r.pkg, importPath)
		}
	}

	return imported && strings.Contains(source, "&"+r.pkg+".Module{}"), nil
}

// files returns the module files by slash path relative to the module directory.
func (r *moduleStubs) files(api, plain bool) map[string]string {
	if plain {
		return map[string]string{"module.go": r.Module(false, false, false)}
	}

	files := map[string]string{
		"module.go":                     r.Module(!api, true, true),
		"providers/service_provider.go": r.ServiceProvider(),
		"routes/api.go":                 r.ApiRoutes(),
		"http/controllers/" + r.snake + "_controller.go": r.Controller(!api),
		"database/migrations/.gitkeep":                   "",
		"database/seeders/.gitkeep":                      "",
		"models/.gitkeep":                                "",
		"lang/en/messages.json":                          r.Lang(),
	}
	if !api {
		files["routes/web.go"] = r.WebRoutes()
		files["resources/views/index.tmpl"] = r.View()
	}
	if file.Exists(path.Test("test_case.go")) {
		files["tests/feature/"+r.snake+"_test.go"] = r.Test()
	}

	return files
}
