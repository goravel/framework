package console

import (
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
)

type ModuleMakeCommandTestSuite struct {
	suite.Suite
}

func TestModuleMakeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleMakeCommandTestSuite))
}

func (s *ModuleMakeCommandTestSuite) SetupTest() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
}

func (s *ModuleMakeCommandTestSuite) TearDownTest() {
	removeScaffold(s.T())
}

// context mocks one make:module call.
func (s *ModuleMakeCommandTestSuite) context(name string, api, plain, force bool, order int) *mocksconsole.Context {
	ctx := mocksconsole.NewContext(s.T())
	ctx.EXPECT().Arguments().Return([]string{name}).Once()
	ctx.EXPECT().OptionBool("api").Return(api).Maybe()
	ctx.EXPECT().OptionBool("plain").Return(plain).Maybe()
	ctx.EXPECT().OptionBool("force").Return(force).Maybe()
	ctx.EXPECT().OptionInt("order").Return(order).Maybe()

	return ctx
}

// files lists the module's files and checks every Go file is gofmt-formatted.
func (s *ModuleMakeCommandTestSuite) files(root string) []string {
	var files []string
	s.Require().NoError(filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		rel, err := filepath.Rel(root, path)
		s.Require().NoError(err)
		files = append(files, filepath.ToSlash(rel))

		if strings.HasSuffix(path, ".go") {
			content, err := os.ReadFile(path)
			s.Require().NoError(err)
			formatted, err := format.Source(content)
			s.Require().NoError(err, rel)
			s.Equal(string(formatted), string(content), rel)
		}

		return nil
	}))

	return files
}

func (s *ModuleMakeCommandTestSuite) TestSignature() {
	cmd := NewModuleMakeCommand(nil)

	s.Equal("make:module", cmd.Signature())
	s.Equal("Create a new module", cmd.Description())
	s.Equal("make", cmd.Extend().Category)
	s.Len(cmd.Extend().Flags, 4)
}

func (s *ModuleMakeCommandTestSuite) TestDefault() {
	s.Require().NoError(file.PutContent("tests/test_case.go", "package tests\n"))

	ctx := s.context("Blog", false, false, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.ElementsMatch([]string{
		"module.go",
		"providers/service_provider.go",
		"routes/web.go",
		"routes/api.go",
		"http/controllers/blog_controller.go",
		"resources/views/index.tmpl",
		"lang/en/messages.json",
		"database/migrations/.gitkeep",
		"database/seeders/.gitkeep",
		"models/.gitkeep",
		"tests/feature/blog_test.go",
	}, s.files("modules/blog"))

	module, err := file.GetContent("modules/blog/module.go")
	s.NoError(err)
	s.Contains(module, "package blog\n")
	s.Contains(module, "//go:embed resources/views\nvar views embed.FS")
	s.Contains(module, "//go:embed lang\nvar lang embed.FS")
	s.Contains(module, `"github.com/goravel/framework/modules/blog/providers"`)
	s.Contains(module, "return \"blog\"\n")
	for _, method := range []string{"ServiceProviders()", "Routes(router route.Router)", "ApiRoutes(router route.Router)", "Config()", "Views()", "Translations()", "ViteEntries()"} {
		s.Contains(module, "func (m *Module) "+method)
	}
	s.NotContains(module, "Order()")

	// No app/facades here, so the controller uses the framework facades.
	controller, err := file.GetContent("modules/blog/http/controllers/blog_controller.go")
	s.NoError(err)
	s.Contains(controller, `"github.com/goravel/framework/facades"`)
	s.Contains(controller, `Make("blog/index.tmpl"`)
	s.Contains(controller, `facades.Lang(ctx).Get("blog/messages.title")`)

	s.True(file.Contain("modules/blog/routes/web.go", `router.Get("/blog", controllers.NewBlogController().Index)`))
	s.True(file.Contain("modules/blog/routes/api.go", `router.Get("/blog", controllers.NewBlogController().Show)`))
	s.True(file.Contain("modules/blog/lang/en/messages.json", `"title": "Blog"`))
	s.False(file.Contain("modules/blog/resources/views/index.tmpl", "define"))
	s.True(file.Contain("modules/blog/tests/feature/blog_test.go", `"github.com/goravel/framework/tests"`))
	s.True(file.Contain("modules/blog/tests/feature/blog_test.go", "tests.TestCase"))

	s.True(file.Contain("bootstrap/app.go", "WithModules(Modules)"))
	s.True(file.Contain("bootstrap/modules.go", `"github.com/goravel/framework/modules/blog"`))
	s.True(file.Contain("bootstrap/modules.go", "&blog.Module{},"))
}

func (s *ModuleMakeCommandTestSuite) TestApi() {
	ctx := s.context("billing", true, false, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/billing").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.ElementsMatch([]string{
		"module.go",
		"providers/service_provider.go",
		"routes/api.go",
		"http/controllers/billing_controller.go",
		"lang/en/messages.json",
		"database/migrations/.gitkeep",
		"database/seeders/.gitkeep",
		"models/.gitkeep",
	}, s.files("modules/billing"))

	module, err := file.GetContent("modules/billing/module.go")
	s.NoError(err)
	s.Contains(module, "func (m *Module) ApiRoutes(router route.Router)")
	s.Contains(module, "//go:embed lang\nvar lang embed.FS")
	s.Contains(module, "func (m *Module) Translations() fs.FS")
	s.True(file.Contain("modules/billing/lang/en/messages.json", `"title": "Billing"`))
	for _, absent := range []string{"resources/views", "Routes(router route.Router) {\n\troutes.Web", "Views()", "ViteEntries()"} {
		s.NotContains(module, absent)
	}

	controller, err := file.GetContent("modules/billing/http/controllers/billing_controller.go")
	s.NoError(err)
	s.NotContains(controller, "Index")
	s.NotContains(controller, "facades")
	s.Contains(controller, `"module": "billing"`)
}

func (s *ModuleMakeCommandTestSuite) TestPlainWithOrder() {
	ctx := s.context("Blog", false, true, false, 5)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.Equal([]string{"module.go"}, s.files("modules/blog"))
	module, err := file.GetContent("modules/blog/module.go")
	s.NoError(err)
	s.Equal(`package blog

type Module struct{}

func (m *Module) Name() string {
	return "blog"
}

func (m *Module) Order() int {
	return 5
}
`, module)
	s.True(file.Contain("bootstrap/modules.go", "&blog.Module{},"))
}

func (s *ModuleMakeCommandTestSuite) TestKebabNameAndConfiguredPath() {
	config := mocksconfig.NewConfig(s.T())
	config.EXPECT().GetString("modular.path", "modules").Return("src/modules").Once()

	ctx := s.context("BlogPosts", false, false, false, 0)
	ctx.EXPECT().Success("Module created successfully: src/modules/blog-posts").Once()

	s.NoError(NewModuleMakeCommand(config).Handle(ctx))

	s.Contains(s.files("src/modules/blog-posts"), "http/controllers/blog_posts_controller.go")
	module, err := file.GetContent("src/modules/blog-posts/module.go")
	s.NoError(err)
	s.Contains(module, "package blogposts\n")
	s.Contains(module, "return \"blog-posts\"\n")
	s.True(file.Contain("src/modules/blog-posts/http/controllers/blog_posts_controller.go", "type BlogPostsController struct{}"))
	s.True(file.Contain("src/modules/blog-posts/lang/en/messages.json", `"title": "Blog Posts"`))
	s.True(file.Contain("bootstrap/modules.go", `blogposts "github.com/goravel/framework/src/modules/blog-posts"`))
	s.True(file.Contain("bootstrap/modules.go", "&blogposts.Module{},"))
}

func (s *ModuleMakeCommandTestSuite) TestAppFacadesWithoutLang() {
	s.Require().NoError(file.PutContent("app/facades/config.go", "package facades\n"))

	ctx := s.context("Blog", false, false, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	controller, err := file.GetContent("modules/blog/http/controllers/blog_controller.go")
	s.NoError(err)
	s.NotContains(controller, "facades")
	s.Contains(controller, `"title": "Blog",`)
	s.files("modules/blog")

	// With the app's Lang facade, the controller uses it.
	s.Require().NoError(file.PutContent("app/facades/lang.go", "package facades\n"))
	ctx = s.context("Blog", false, false, true, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.True(file.Contain("modules/blog/http/controllers/blog_controller.go", `"github.com/goravel/framework/app/facades"`))
	s.files("modules/blog")
}

func (s *ModuleMakeCommandTestSuite) TestExistingModulesFileWithoutWithModules() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", assignedApp))
	s.Require().NoError(file.PutContent("bootstrap/modules.go", modify.ModulesStub()))

	warning := "Add .WithModules(Modules) to foundation.Setup() in bootstrap/app.go"
	ctx := s.context("Blog", false, true, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()
	ctx.EXPECT().Warning(warning).Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.True(file.Contain("bootstrap/modules.go", "&blog.Module{},"))
	app, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(assignedApp, app)
}

// An assigned chain cannot be edited: modules.go is written and the user is told to wire it.
func (s *ModuleMakeCommandTestSuite) TestAssignedChainWithoutModulesFile() {
	s.Require().NoError(file.PutContent("bootstrap/app.go", assignedApp))

	ctx := s.context("Blog", false, true, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()
	ctx.EXPECT().Warning("Add .WithModules(Modules) to foundation.Setup() in bootstrap/app.go").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.True(file.Contain("bootstrap/modules.go", "&blog.Module{},"))
	app, err := file.GetContent("bootstrap/app.go")
	s.NoError(err)
	s.Equal(assignedApp, app)
}

// --force removes the stubs the new variant does not write, keeps other files and registers once.
func (s *ModuleMakeCommandTestSuite) TestForceRemovesStaleStubs() {
	ctx := s.context("Blog", false, false, false, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()
	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
	s.Require().NoError(file.PutContent("modules/blog/resources/js/app.js", ""))
	s.Require().NoError(file.PutContent("modules/blog/models/post.go", "package models\n"))

	ctx = s.context("Blog", true, false, true, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()
	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.ElementsMatch([]string{
		"module.go",
		"providers/service_provider.go",
		"routes/api.go",
		"http/controllers/blog_controller.go",
		"lang/en/messages.json",
		"database/migrations/.gitkeep",
		"database/seeders/.gitkeep",
		"models/.gitkeep",
		"models/post.go",
		"resources/js/app.js",
	}, s.files("modules/blog"))
	s.False(file.Exists("modules/blog/resources/views"))

	ctx = s.context("Blog", false, true, true, 0)
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()
	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

	s.ElementsMatch([]string{"module.go", "models/post.go", "resources/js/app.js"}, s.files("modules/blog"))

	modules, err := file.GetContent("bootstrap/modules.go")
	s.NoError(err)
	s.Equal(1, strings.Count(modules, "&blog.Module{},"))
}

func (s *ModuleMakeCommandTestSuite) TestErrors() {
	s.Run("exists without force", func() {
		s.Require().NoError(file.PutContent("modules/blog/module.go", "package blog\n"))

		ctx := s.context("Blog", false, false, false, 0)
		ctx.EXPECT().Error(errors.ModularModuleExists.Args("blog", "modules/blog").Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		s.Equal([]string{"module.go"}, s.files("modules/blog"))
		s.False(file.Exists("bootstrap/modules.go"))
	})

	s.Run("api and plain", func() {
		ctx := s.context("Shop", true, true, false, 0)
		ctx.EXPECT().Error(errors.ModularOptionsConflict.Args("api", "plain").Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		s.False(file.Exists("modules/shop"))
	})

	for name, pkg := range map[string]string{"type": "type", "Go": "go", "main": "main", "Init": "init", "internal": "internal", "vendor": "vendor"} {
		s.Run("reserved name "+name, func() {
			ctx := s.context(name, false, false, false, 0)
			ctx.EXPECT().Error(errors.ModularModuleNameForbidden.Args(name, pkg).Error()).Once()

			s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
			s.False(file.Exists("modules/" + pkg))
		})
	}

	s.Run("package name of the contracts import", func() {
		ctx := s.context("Modular", false, false, false, 0)
		ctx.EXPECT().Error(errors.ModularModuleImportConflict.Args("modular", "modular", "github.com/goravel/framework/contracts/modular").Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		s.False(file.Exists("modules/modular"))
		s.False(file.Exists("bootstrap/modules.go"))
	})

	s.Run("package name of another module", func() {
		ctx := s.context("BlogPosts", false, true, false, 0)
		ctx.EXPECT().Success("Module created successfully: modules/blog-posts").Once()
		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))

		ctx = s.context("Blogposts", false, true, false, 0)
		ctx.EXPECT().Error(errors.ModularModuleImportConflict.Args("blogposts", "blogposts", "github.com/goravel/framework/modules/blog-posts").Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		s.False(file.Exists("modules/blogposts"))
		s.Equal(1, strings.Count(mustContent(s.T(), "bootstrap/modules.go"), "blogposts"+" "))
		removeScaffold(s.T())
		s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	})

	for _, name := range []string{"1blog", "blog/posts", "blog.posts", "-"} {
		s.Run("invalid name "+name, func() {
			ctx := s.context(name, false, false, false, 0)
			ctx.EXPECT().Error(errors.ModularModuleNameInvalid.Args(name).Error()).Once()

			s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		})
	}

	s.Run("empty name", func() {
		ctx := s.context("", false, false, false, 0)
		ctx.EXPECT().Ask("Enter the module name", mock.Anything).Return("", errors.ModularModuleNameEmpty).Once()
		ctx.EXPECT().Error(errors.ModularModuleNameEmpty.Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
	})

	s.Run("legacy bootstrap", func() {
		s.Require().NoError(file.PutContent("bootstrap/app.go", "package bootstrap\n"))

		ctx := s.context("Shop", false, false, false, 0)
		ctx.EXPECT().Error(errors.ModularRequiresBootstrapSetup.Args("make:module").Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
		s.False(file.Exists("modules/shop"))
	})
}

func mustContent(t *testing.T, path string) string {
	content, err := file.GetContent(path)
	if err != nil {
		t.Fatal(err)
	}

	return content
}

func (s *ModuleMakeCommandTestSuite) TestAsksForName() {
	ctx := s.context("", false, true, false, 0)
	ctx.EXPECT().Ask("Enter the module name", mock.Anything).RunAndReturn(func(_ string, options ...console.AskOption) (string, error) {
		s.Require().Len(options, 1)
		s.Equal(errors.ModularModuleNameEmpty, options[0].Validate(""))
		s.NoError(options[0].Validate("Blog"))

		return "Blog", nil
	}).Once()
	ctx.EXPECT().Success("Module created successfully: modules/blog").Once()

	s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
	s.Equal([]string{"module.go"}, s.files("modules/blog"))
}

func (s *ModuleMakeCommandTestSuite) TestFailures() {
	// failed runs make:module and expects one error, of any text.
	failed := func(name string, plain, force bool) {
		ctx := s.context(name, false, plain, force, 0)
		ctx.EXPECT().Error(mock.Anything).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
	}
	reset := func() {
		removeScaffold(s.T())
		s.Require().NoError(file.PutContent("bootstrap/app.go", chainedApp))
	}

	s.Run("blank name", func() {
		ctx := s.context("  ", false, false, false, 0)
		ctx.EXPECT().Error(errors.ModularModuleNameEmpty.Error()).Once()

		s.NoError(NewModuleMakeCommand(nil).Handle(ctx))
	})

	s.Run("bootstrap/modules.go is a directory", func() {
		s.Require().NoError(file.PutContent("bootstrap/modules.go/keep", ""))

		failed("Blog", false, false)
		s.False(file.Exists("modules/blog"))
		reset()
	})

	s.Run("bootstrap/modules.go is not Go", func() {
		s.Require().NoError(file.PutContent("bootstrap/modules.go", "not go\n"))

		failed("Blog", false, false)
		s.False(file.Exists("modules/blog"))
		reset()
	})

	s.Run("bootstrap/modules.go cannot be edited", func() {
		// The imports parse, the rest does not.
		s.Require().NoError(file.PutContent("bootstrap/modules.go", "package bootstrap\n\nfunc Modules( {\n"))

		failed("Blog", true, false)
		s.True(file.Exists("modules/blog/module.go"))
		reset()
	})

	s.Run("a stale stub cannot be removed", func() {
		// --force --plain removes routes/api.go, here a directory with content.
		s.Require().NoError(file.PutContent("modules/blog/routes/api.go/keep", ""))

		failed("Blog", true, true)
		s.False(file.Exists("modules/blog/module.go"))
		reset()
	})

	s.Run("a stub cannot be written", func() {
		s.Require().NoError(file.PutContent("modules/blog/module.go/keep", ""))

		failed("Blog", true, true)
		s.False(file.Exists("bootstrap/modules.go"))
		reset()
	})
}
