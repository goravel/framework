package foundation

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing/fstest"

	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksevent "github.com/goravel/framework/mocks/event"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksroute "github.com/goravel/framework/mocks/route"
	mocksschedule "github.com/goravel/framework/mocks/schedule"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
)

// testModule has only a name; tests embed it to add capabilities.
type testModule struct{ name string }

func (m *testModule) Name() string { return m.name }

type dependentTestModule struct {
	testModule
	dependencies []string
}

func (m *dependentTestModule) Dependencies() []string { return m.dependencies }

func moduleNames(modules []contractsmodule.Module) []string {
	var names []string
	for _, module := range modules {
		names = append(names, module.Name())
	}

	return names
}

func (s *ApplicationTestSuite) TestConfigureModules() {
	s.Run("without WithModules an empty registry is bound", func() {
		s.SetupTest()
		s.app.builder = NewApplicationBuilder(s.app)

		s.app.configureModules()

		registry := s.app.MakeModule()
		s.Require().NotNil(registry)
		s.Empty(registry.All())
		s.Empty(s.app.modules)
	})

	s.Run("modules are registered and kept in dependency order", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{
				&dependentTestModule{testModule{"comments"}, []string{"blog"}},
				&testModule{"blog"},
				&testModule{"shop"},
			}
		}
		s.app.builder = builder

		s.app.configureModules()

		s.Equal([]string{"blog", "comments", "shop"}, moduleNames(s.app.modules))
		s.True(s.app.MakeModule().Has("blog"))
		instance, err := s.app.Make(binding.Module)
		s.NoError(err)
		s.Same(s.app.MakeModule(), instance)
	})

	s.Run("module directories are resolved, third-party ones under module.path", func() {
		s.SetupTest()
		mockConfig := mocksconfig.NewConfig(s.T())
		mockConfig.EXPECT().GetString("module.path", "modules").Return("app/modules").Once()
		s.app.Instance(binding.Config, mockConfig)
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{&testModule{"blog"}}
		}
		s.app.builder = builder

		s.app.configureModules()

		// Under go test the main module is the framework.
		dir, ok := s.app.MakeModule().Path("blog")
		s.True(ok)
		s.Equal("foundation", dir)
	})

	s.Run("a duplicate name stops the boot", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{&testModule{"blog"}, &testModule{"blog"}, &testModule{"shop"}}
		}
		s.app.builder = builder

		s.PanicsWithError(`[module] module "blog" is registered more than once`, func() {
			s.app.configureModules()
		})
	})

	s.Run("an unresolvable dependency stops the boot", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{
				&dependentTestModule{testModule{"comments"}, []string{"blog"}},
				&testModule{"shop"},
			}
		}
		s.app.builder = builder

		s.PanicsWithError("[module] module comments depends on module blog, which is not registered", func() {
			s.app.configureModules()
		})
	})

	s.Run("a dependency cycle stops the boot", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{
				&dependentTestModule{testModule{"blog"}, []string{"comments"}},
				&dependentTestModule{testModule{"comments"}, []string{"blog"}},
			}
		}
		s.app.builder = builder

		s.PanicsWithError("[module] modules depend on each other: blog -> comments -> blog", func() {
			s.app.configureModules()
		})
	})

	s.Run("a rebuild binds a fresh registry", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodule.Module {
			return []contractsmodule.Module{&testModule{"blog"}}
		}
		s.app.builder = builder

		s.app.configureModules()
		first := s.app.MakeModule()
		s.app.configureModules()

		s.NotSame(first, s.app.MakeModule())
		s.Len(s.app.MakeModule().All(), 1)
	})

	s.Run("without Build there is no registry", func() {
		s.SetupTest()

		s.Nil(s.app.MakeModule())
	})
}

// richModule implements every capability with test-set values.
type richModule struct {
	testModule
	config      map[string]any
	translation fs.FS
	views       any
	middleware  []contractshttp.Middleware
	providers   []foundation.ServiceProvider
	commands    []console.Command
	migrations  []schema.Migration
	seeders     []seeder.Seeder
	jobs        []queue.Job
	schedule    []schedule.Event
	events      map[event.Event][]event.Listener
	routed      *[]string
}

func (m *richModule) Config() map[string]any              { return m.config }
func (m *richModule) Translations() contractsmodule.Files { return filesOf(m.translation) }
func (m *richModule) Views() contractsmodule.Files        { return filesOf(m.views) }

// filesOf turns a test value into files: an fs.FS, a directory without one, or nothing.
func filesOf(value any) contractsmodule.Files {
	switch value := value.(type) {
	case fs.FS:
		return contractsmodule.Files{FS: value, Dir: "."}
	case string:
		return contractsmodule.Files{Dir: value}
	}

	return contractsmodule.Files{}
}
func (m *richModule) GlobalMiddleware() []contractshttp.Middleware   { return m.middleware }
func (m *richModule) ServiceProviders() []foundation.ServiceProvider { return m.providers }
func (m *richModule) Commands() []console.Command                    { return m.commands }
func (m *richModule) Migrations() []schema.Migration                 { return m.migrations }
func (m *richModule) Seeders() []seeder.Seeder                       { return m.seeders }
func (m *richModule) Jobs() []queue.Job                              { return m.jobs }
func (m *richModule) Schedule() []schedule.Event                     { return m.schedule }
func (m *richModule) Events() map[event.Event][]event.Listener       { return m.events }
func (m *richModule) Routes()                                        { *m.routed = append(*m.routed, m.name+":web") }

// withModules points the builder at modules and runs configureModules.
func (s *ApplicationTestSuite) withModules(builder *ApplicationBuilder, modules ...contractsmodule.Module) {
	builder.modules = func() []contractsmodule.Module { return modules }
	s.app.builder = builder
	s.app.configureModules()
}

type fakeMiddleware struct{ signature string }

func (m *fakeMiddleware) Handle(contractshttp.Context) {}
func (m *fakeMiddleware) Signature() string            { return m.signature }

func middlewareSignatures(middleware []contractshttp.Middleware) []string {
	var out []string
	for _, m := range middleware {
		out = append(out, m.Signature())
	}

	return out
}

func (s *ApplicationTestSuite) TestConfigureModuleConfig() {
	s.Run("module config and translations, app lang FS kept as fallback", func() {
		s.SetupTest()
		appLang := fstest.MapFS{"en/app.json": {Data: []byte(`{"name":"App"}`)}}
		blogLang := fstest.MapFS{"en/messages.json": {Data: []byte(`{"title":"Blog"}`)}}

		mockConfig := mocksconfig.NewConfig(s.T())
		mockConfig.EXPECT().GetString("module.path", "modules").Return("modules").Once()
		mockConfig.EXPECT().Add("modules.blog", map[string]any{"enabled": true}).Return().Once()
		mockConfig.EXPECT().Add("modules.shop", map[string]any{"currency": "EUR"}).Return().Once()
		mockConfig.EXPECT().Get("app.lang_fs").Return(appLang).Once()
		mockConfig.EXPECT().Add("app.lang_fs", mock.MatchedBy(func(fsys fs.FS) bool {
			module, err := fs.ReadFile(fsys, "en/blog/messages.json")
			app, appErr := fs.ReadFile(fsys, "en/app.json")

			return err == nil && appErr == nil && string(module) == `{"title":"Blog"}` && string(app) == `{"name":"App"}`
		})).Return().Once()
		s.app.Instance(binding.Config, mockConfig)

		s.withModules(NewApplicationBuilder(s.app),
			&richModule{testModule: testModule{"blog"}, config: map[string]any{"enabled": true}, translation: blogLang},
			&richModule{testModule: testModule{"shop"}, config: map[string]any{"currency": "EUR"}},
		)
		s.app.configureModuleConfig()
	})

	s.Run("without modules the config facade is not touched", func() {
		s.SetupTest()
		s.withModules(NewApplicationBuilder(s.app), &testModule{"plain"})
		s.app.configureModuleConfig() // no config bound: must not be needed
	})
}

func (s *ApplicationTestSuite) TestConfigureServiceProviders_Modules() {
	s.SetupTest()
	s.app.providerRepository = NewProviderRepository()
	appProvider := &AServiceProvider{}
	moduleProvider := &BServiceProvider{}

	builder := NewApplicationBuilder(s.app)
	builder.configuredServiceProviders = func() []foundation.ServiceProvider { return []foundation.ServiceProvider{appProvider} }
	s.withModules(builder, &richModule{testModule: testModule{"blog"}, providers: []foundation.ServiceProvider{moduleProvider}})
	s.app.configureServiceProviders()

	repository := s.app.providerRepository.(*ProviderRepository)
	s.Equal([]foundation.ServiceProvider{appProvider, moduleProvider}, repository.providers)
}

func (s *ApplicationTestSuite) TestConfigureMiddleware_Modules() {
	driver := &fakeMiddleware{"driver"}
	module := &fakeMiddleware{"module"}
	app := &fakeMiddleware{"app"}

	s.Run("module middleware without WithMiddleware", func() {
		s.SetupTest()
		mockRoute := mocksroute.NewRoute(s.T())
		mockRoute.EXPECT().GetGlobalMiddleware().Return([]contractshttp.Middleware{driver}).Once()
		mockRoute.EXPECT().SetGlobalMiddleware(mock.MatchedBy(func(m []contractshttp.Middleware) bool {
			return slices.Equal(middlewareSignatures(m), []string{"driver", "module"})
		})).Return().Once()
		s.app.Instance(binding.Route, mockRoute)

		s.withModules(NewApplicationBuilder(s.app), &richModule{testModule: testModule{"blog"}, middleware: []contractshttp.Middleware{module}})
		s.app.configureMiddleware()
	})

	s.Run("WithMiddleware sees module middleware and runs last", func() {
		s.SetupTest()
		mockRoute := mocksroute.NewRoute(s.T())
		mockRoute.EXPECT().GetGlobalMiddleware().Return([]contractshttp.Middleware{driver}).Once()
		mockRoute.EXPECT().SetGlobalMiddleware(mock.MatchedBy(func(m []contractshttp.Middleware) bool {
			return slices.Equal(middlewareSignatures(m), []string{"driver", "module", "app"})
		})).Return().Once()
		s.app.Instance(binding.Route, mockRoute)

		builder := NewApplicationBuilder(s.app)
		builder.middleware = func(handler contractsconfiguration.Middleware) {
			s.Len(handler.GetGlobalMiddleware(), 2)
			handler.Append(app)
		}
		s.withModules(builder, &richModule{testModule: testModule{"blog"}, middleware: []contractshttp.Middleware{module}})
		s.app.configureMiddleware()
	})
}

func (s *ApplicationTestSuite) TestRegisterModuleResources() {
	s.Run("views are registered under the module name", func() {
		s.SetupTest()
		mockView := mocksview.NewView(s.T())
		mockView.EXPECT().LoadViewsFromFS(mock.MatchedBy(func(fsys fs.FS) bool {
			_, err := fs.Stat(fsys, "blog/index.tmpl")

			return err == nil
		}), ".").Return().Once()
		s.app.Instance(binding.View, mockView)

		s.withModules(NewApplicationBuilder(s.app),
			&richModule{testModule: testModule{"blog"}, views: fstest.MapFS{"index.tmpl": {Data: []byte("x")}}},
			&richModule{testModule: testModule{"empty"}}, // nil views: skipped
			&testModule{"plain"},
		)
		s.app.registerModuleResources()
	})

	s.Run("a views directory without an FS is not read from disk", func() {
		s.SetupTest()
		relativePath := support.RelativePath
		defer func() { support.RelativePath = relativePath }()
		support.RelativePath = s.T().TempDir()

		// The view mock has no expectations: nothing is registered.
		s.app.Instance(binding.View, mocksview.NewView(s.T()))

		s.withModules(NewApplicationBuilder(s.app),
			&richModule{testModule: testModule{"blog"}, views: "resources/views"},
		)
		root, _ := s.app.MakeModule().Path("blog")
		dir := filepath.Join(support.RelativePath, root, "resources", "views")
		s.Require().NoError(os.MkdirAll(dir, 0o755))
		s.Require().NoError(os.WriteFile(filepath.Join(dir, "index.tmpl"), []byte("x"), 0o644))
		s.app.registerModuleResources()
	})

	s.Run("without view modules the view facade is not needed", func() {
		s.SetupTest()
		s.withModules(NewApplicationBuilder(s.app), &testModule{"plain"})
		s.app.registerModuleResources()
	})
}

func (s *ApplicationTestSuite) TestConfigureRegistrations_Modules() {
	s.SetupTest()
	appCommand, moduleCommand := &fakeCommand{signature: "app:run"}, &fakeCommand{signature: "blog:run"}
	appMigration, moduleMigration := mocksschema.NewMigration(s.T()), mocksschema.NewMigration(s.T())
	appSeeder, moduleSeeder := mocksseeder.NewSeeder(s.T()), mocksseeder.NewSeeder(s.T())
	appJob, moduleJob := mocksqueue.NewJob(s.T()), mocksqueue.NewJob(s.T())
	appEvent, moduleEvent := mocksschedule.NewEvent(s.T()), mocksschedule.NewEvent(s.T())

	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Register([]console.Command{appCommand, moduleCommand}).Return().Once()
	s.app.Instance(binding.Artisan, mockArtisan)
	mockSchema := mocksschema.NewSchema(s.T())
	mockSchema.EXPECT().Register([]schema.Migration{appMigration, moduleMigration}).Return().Once()
	s.app.Instance(binding.Schema, mockSchema)
	mockSeeder := mocksseeder.NewFacade(s.T())
	mockSeeder.EXPECT().Register([]seeder.Seeder{appSeeder, moduleSeeder}).Return().Once()
	s.app.Instance(binding.Seeder, mockSeeder)
	mockQueue := mocksqueue.NewQueue(s.T())
	mockQueue.EXPECT().Register([]queue.Job{appJob, moduleJob}).Return().Once()
	s.app.Instance(binding.Queue, mockQueue)
	mockSchedule := mocksschedule.NewSchedule(s.T())
	mockSchedule.EXPECT().Register([]schedule.Event{appEvent, moduleEvent}).Return().Once()
	s.app.Instance(binding.Schedule, mockSchedule)

	builder := NewApplicationBuilder(s.app)
	builder.commands = func() []console.Command { return []console.Command{appCommand} }
	builder.migrations = func() []schema.Migration { return []schema.Migration{appMigration} }
	builder.seeders = func() []seeder.Seeder { return []seeder.Seeder{appSeeder} }
	builder.jobs = func() []queue.Job { return []queue.Job{appJob} }
	builder.schedule = func() []schedule.Event { return []schedule.Event{appEvent} }
	s.withModules(builder, &richModule{
		testModule: testModule{"blog"},
		commands:   []console.Command{moduleCommand},
		migrations: []schema.Migration{moduleMigration},
		seeders:    []seeder.Seeder{moduleSeeder},
		jobs:       []queue.Job{moduleJob},
		schedule:   []schedule.Event{moduleEvent},
	})

	s.app.configureCommands()
	s.app.configureMigrations()
	s.app.configureSeeders()
	s.app.configureJobs()
	s.app.configureSchedule()
}

func (s *ApplicationTestSuite) TestConfigureRegistrations_ModulesOnly() {
	s.SetupTest()
	moduleCommand := &fakeCommand{signature: "blog:run"}
	mockArtisan := mocksconsole.NewArtisan(s.T())
	mockArtisan.EXPECT().Register([]console.Command{moduleCommand}).Return().Once()
	s.app.Instance(binding.Artisan, mockArtisan)

	s.withModules(NewApplicationBuilder(s.app), &richModule{testModule: testModule{"blog"}, commands: []console.Command{moduleCommand}})
	s.app.configureCommands()
}

func (s *ApplicationTestSuite) TestConfigureEventListeners_Modules() {
	s.SetupTest()
	shared, moduleOnly := mocksevent.NewEvent(s.T()), mocksevent.NewEvent(s.T())
	appListener, moduleListener, otherListener := mocksevent.NewListener(s.T()), mocksevent.NewListener(s.T()), mocksevent.NewListener(s.T())

	mockEvent := mocksevent.NewInstance(s.T())
	mockEvent.EXPECT().Register(map[event.Event][]event.Listener{
		shared:     {appListener, moduleListener},
		moduleOnly: {otherListener},
	}).Return().Twice()
	s.app.Instance(binding.Event, mockEvent)

	// One map for every build, its slice with spare capacity: neither may be written to.
	appListeners := make([]event.Listener, 1, 4)
	appListeners[0] = appListener
	appEvents := map[event.Event][]event.Listener{shared: appListeners}

	builder := NewApplicationBuilder(s.app)
	builder.eventToListeners = func() map[event.Event][]event.Listener {
		return appEvents
	}
	s.withModules(builder, &richModule{testModule: testModule{"blog"}, events: map[event.Event][]event.Listener{
		shared:     {moduleListener},
		moduleOnly: {otherListener},
	}})
	s.app.configureEventListeners()
	s.app.configureEventListeners()

	s.Equal(map[event.Event][]event.Listener{shared: {appListener}}, appEvents)
	s.Equal([]event.Listener{appListener, nil}, appListeners[:2])
}

func (s *ApplicationTestSuite) TestConfigureRoutes_Modules() {
	s.Run("module routes are registered in module order, after the app's", func() {
		s.SetupTest()
		var routed []string

		s.app.Instance(binding.Route, mocksroute.NewRoute(s.T()))

		builder := NewApplicationBuilder(s.app)
		builder.routes = func() { routed = append(routed, "app") }
		s.withModules(builder,
			&richModule{testModule: testModule{"blog"}, routed: &routed},
			&webOnlyModule{testModule{"shop"}, &routed},
		)
		s.app.configureRoutes()

		s.Equal([]string{"app", "blog:web", "shop:web"}, routed)
	})

	s.Run("late routes follow every module's routes, dependants first", func() {
		s.SetupTest()
		var routed []string

		s.app.Instance(binding.Route, mocksroute.NewRoute(s.T()))

		builder := NewApplicationBuilder(s.app)
		builder.routes = func() { routed = append(routed, "app") }
		s.withModules(builder,
			&lateRoutesModule{webOnlyModule{testModule{"admin"}, &routed}},
			&webOnlyModule{testModule{"shop"}, &routed},
			&lateRoutesModule{webOnlyModule{testModule{"reports"}, &routed}},
			&lateOnlyModule{dependentTestModule{testModule{"blog"}, []string{"admin"}}, &routed},
		)
		s.app.configureRoutes()

		s.Equal([]string{
			"app", "admin:web", "shop:web", "reports:web",
			"blog:late", "reports:late", "admin:late",
		}, routed)
	})

	s.Run("late routes alone need the route facade", func() {
		s.SetupTest()
		var routed []string

		s.app.Instance(binding.Route, mocksroute.NewRoute(s.T()))
		s.withModules(NewApplicationBuilder(s.app),
			&lateOnlyModule{dependentTestModule{testModule{"admin"}, nil}, &routed},
		)
		s.app.configureRoutes()

		s.Equal([]string{"admin:late"}, routed)
	})

	s.Run("without module routes or static files the route facade is not needed", func() {
		s.SetupTest()
		s.withModules(NewApplicationBuilder(s.app), &testModule{"plain"})
		s.app.configureRoutes()
	})
}

type webOnlyModule struct {
	testModule
	routed *[]string
}

func (m *webOnlyModule) Routes() { *m.routed = append(*m.routed, m.name+":web") }

type lateRoutesModule struct {
	webOnlyModule
}

func (m *lateRoutesModule) LateRoutes() { *m.routed = append(*m.routed, m.name+":late") }

type lateOnlyModule struct {
	dependentTestModule
	routed *[]string
}

func (m *lateOnlyModule) LateRoutes() { *m.routed = append(*m.routed, m.name+":late") }

type staticModule struct {
	testModule
	static map[string]contractsmodule.Files
}

func (m *staticModule) Static() map[string]contractsmodule.Files { return m.static }

func (s *ApplicationTestSuite) TestConfigureRoutes_ModuleStatic() {
	s.SetupTest()
	public := contractsmodule.Files{FS: fstest.MapFS{"public/robots.txt": {Data: []byte("x")}}, Dir: "public"}

	mockRoute := mocksroute.NewRoute(s.T())
	mockRoute.EXPECT().StaticFS("/blog/static", mock.MatchedBy(func(fsys http.FileSystem) bool {
		file, err := fsys.Open("/robots.txt")
		if err != nil {
			return false
		}
		_ = file.Close()
		_, err = fsys.Open("/")

		return err != nil // directories are not listed
	})).Return(mocksroute.NewAction(s.T())).Once()
	s.app.Instance(binding.Route, mockRoute)

	s.withModules(NewApplicationBuilder(s.app),
		&staticModule{testModule: testModule{"blog"}, static: map[string]contractsmodule.Files{
			"blog/static/": public,
			"/":            public,              // invalid prefix: reported
			"/blog/secret": {Dir: "../outside"}, // escapes the module: reported
			"/blog/empty":  {},                  // absent: skipped
		}},
		&staticModule{testModule: testModule{"shop"}, static: map[string]contractsmodule.Files{
			"/blog/static": public, // taken by blog: reported
		}},
	)
	s.app.configureRoutes()
}

// Capabilities whose facade is not installed are reported, not fatal.
func (s *ApplicationTestSuite) TestModules_FacadesNotInstalled() {
	s.SetupTest()
	routed := []string{}
	s.withModules(NewApplicationBuilder(s.app),
		&richModule{
			testModule:  testModule{"blog"},
			config:      map[string]any{"enabled": true},
			translation: fstest.MapFS{"en/messages.json": {Data: []byte(`{}`)}},
			views:       fstest.MapFS{"index.tmpl": {Data: []byte("x")}},
			migrations:  []schema.Migration{mocksschema.NewMigration(s.T())},
			seeders:     []seeder.Seeder{mocksseeder.NewSeeder(s.T())},
			jobs:        []queue.Job{mocksqueue.NewJob(s.T())},
			schedule:    []schedule.Event{mocksschedule.NewEvent(s.T())},
			events:      map[event.Event][]event.Listener{mocksevent.NewEvent(s.T()): {mocksevent.NewListener(s.T())}},
			routed:      &routed,
		},
	)

	output := color.CaptureOutput(func(io.Writer) {
		s.app.configureModuleConfig()
		s.app.configureEventListeners()
		s.app.configureJobs()
		s.app.configureMigrations()
		s.app.configureRoutes()
		s.app.configureSchedule()
		s.app.configureSeeders()
		s.app.registerModuleResources()
	})

	s.Contains(output, errors.ConfigFacadeNotSet.SetModule(errors.ModuleModule).Error())
	for _, facade := range []string{"Event", "Queue", "Schema", "Route", "Schedule", "Seeder", "View"} {
		s.Contains(output, facade+" facade not found")
	}
	s.Empty(routed)
}
