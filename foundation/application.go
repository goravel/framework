package foundation

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"os/signal"
	stdpath "path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/stats"

	"github.com/goravel/framework/config"
	frameworkconsole "github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/contracts/foundation"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/contracts/validation"
	contractsview "github.com/goravel/framework/contracts/view"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/configuration"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/modular"
	modularconsole "github.com/goravel/framework/modular/console"
	"github.com/goravel/framework/process"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

var App foundation.Application
var _ = flag.String("env", support.EnvFilePath, "custom .env path")

type RunnerWithInfo struct {
	signature string
	runner    foundation.Runner
	running   atomic.Bool
	doneOnce  sync.Once
}

func init() {
	setEnv()
	setRootPath()

	app := &Application{
		Container:     NewContainer(),
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
		runnerWg:      sync.WaitGroup{},
	}

	app.providerRepository = NewProviderRepository()
	App = app

	app.RegisterBaseServiceProviders()
	app.SetJson(json.New())
}

type Application struct {
	*Container
	ctx                context.Context
	cancel             context.CancelFunc
	builder            *ApplicationBuilder
	providerRepository foundation.ProviderRepository
	publishes          map[string]map[string]string
	publishGroups      map[string]map[string]string
	json               foundation.Json
	bootedRunners      []string
	commandsFilter     []string
	commandHooks       map[string][]contractsconsole.CommandHook
	modules            []contractsmodular.Module
	runnerWg           sync.WaitGroup
	runnersToRun       []*RunnerWithInfo
}

func NewApplication() foundation.Application {
	return App
}

func (r *Application) About(section string, items []foundation.AboutItem) {
	console.AddAboutInformation(section, items...)
}

func (r *Application) Boot() {
	r.providerRepository.LoadFromConfig(r.MakeConfig())
	clear(r.publishes)
	clear(r.publishGroups)

	r.setTimezone()

	r.providerRepository.Register(r)
	r.providerRepository.Boot(r)

	r.registerCommands(r.defaultCommands())
	r.bootArtisan()
}

func (r *Application) Build() foundation.Application {
	clear(r.publishes)
	clear(r.publishGroups)

	r.bootedRunners = nil
	r.runnersToRun = nil
	r.ctx, r.cancel = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	r.setTimezone()
	r.configurePaths()
	r.configureCustomConfig()
	r.configureModules()
	r.configureModuleConfig()
	if r.builder.commandsFilter != nil {
		r.commandsFilter = r.builder.commandsFilter()
	}
	if r.builder.commandHooks != nil {
		r.commandHooks = r.builder.commandHooks()
	}
	r.configureServiceProviders()
	r.providerRepository.Register(r)
	r.configureMiddleware()
	r.providerRepository.Boot(r)
	r.registerModuleResources()
	r.configureViews()
	r.configureEventListeners()
	r.configureCommands()
	r.configureSchedule()
	r.configureMigrations()
	r.configureSeeders()
	r.configureGrpc()
	r.configureJobs()
	r.configureValidation()
	r.configureRoutes()
	r.configureRunners()
	r.registerCommands(r.defaultCommands())
	r.configureCallback()
	r.bootArtisan()

	return r
}

func (r *Application) Commands(commands []contractsconsole.Command) {
	r.registerCommands(commands)
}

func (r *Application) Context() context.Context {
	return r.ctx
}

// GetJson get the JSON implementation.
// DEPRECATED, use Json instead.
func (r *Application) GetJson() foundation.Json {
	return r.json
}

func (r *Application) IsLocale(ctx context.Context, locale string) bool {
	return r.CurrentLocale(ctx) == locale
}

func (r *Application) Json() foundation.Json {
	return r.json
}

func (r *Application) Publishes(packageName string, paths map[string]string, groups ...string) {
	if _, exist := r.publishes[packageName]; !exist {
		r.publishes[packageName] = make(map[string]string)
	}
	maps.Copy(r.publishes[packageName], paths)
	for _, group := range groups {
		r.addPublishGroup(group, paths)
	}
}

func (r *Application) Refresh() {
	// clear Container.instances except the config facade to keep the config values
	r.Fresh()

	// reset provider repository
	r.providerRepository.Reset()

	// re-register base service providers
	r.RegisterBaseServiceProviders()

	// rebuild the application
	r.Build()
}

func (r *Application) RegisterBaseServiceProviders() {
	baseProviders := r.baseServiceProviders()
	r.providerRepository.Add(baseProviders)
	r.providerRepository.Register(r)
}

func (r *Application) Restart() error {
	if err := r.Shutdown(); err != nil {
		return err
	}

	r.Refresh()

	go r.Start()

	i := 0
	for {
		failed := ""
		for _, runner := range r.runnersToRun {
			if !runner.running.Load() {
				failed = runner.signature
				break
			}
		}

		if failed == "" {
			break
		}
		if i > 100 {
			return fmt.Errorf("timeout waiting for runner %s to run", failed)
		}

		i++

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (r *Application) Start() {
	var (
		errsMu sync.Mutex
		errs   []error
	)

	// Restart replaces r.cancel; an erroring runner must cancel the context it was started with.
	cancel := r.cancel

	run := func(runner *RunnerWithInfo) {
		r.runnerWg.Add(1)

		go func() {
			runner.running.Store(true)
			if err := runner.runner.Run(); err != nil {
				runner.running.Store(false)
				errsMu.Lock()
				defer errsMu.Unlock()

				errs = append(errs, fmt.Errorf("failed to run %s: %w", runner.signature, err))
				if log := r.MakeLog(); log != nil {
					log.Errorf("failed to run %s: %v\n", runner.signature, err)
				}

				cancel()
			}
			// Run may be a blocking call, so don't write anything after it.
		}()
	}

	for _, runner := range r.runnersToRun {
		run(runner)
	}

	if len(r.runnersToRun) > 0 {
		go r.shutdownRunners()
	}

	r.runnerWg.Wait()

	// Run may error after its shutdown completed, so reading errs still needs the lock.
	errsMu.Lock()
	defer errsMu.Unlock()

	if len(errs) > 0 {
		panic(errors.Join(errs...))
	}
}

func (r *Application) shutdownRunners() {
	<-r.ctx.Done()

	for _, group := range r.runnersByShutdownPriority() {
		var wg sync.WaitGroup

		for _, runner := range group {
			wg.Add(1)

			go func(runner *RunnerWithInfo) {
				defer wg.Done()
				defer runner.doneOnce.Do(func() {
					r.runnerWg.Done()
				})

				// Only call Shutdown if the runner is still running (Run didn't error)
				if !runner.running.Load() {
					return
				}

				if err := runner.runner.Shutdown(); err != nil {
					if log := r.MakeLog(); log != nil {
						log.Errorf("failed to shutdown %s: %v\n", runner.signature, err)
					}
				}

				runner.running.Store(false)
			}(runner)
		}

		wg.Wait()
	}
}

func (r *Application) runnersByShutdownPriority() [][]*RunnerWithInfo {
	groups := map[int][]*RunnerWithInfo{}
	for _, runner := range r.runnersToRun {
		priority := 0
		if p, ok := runner.runner.(foundation.RunnerWithShutdownPriority); ok {
			priority = p.ShutdownPriority()
		}
		groups[priority] = append(groups[priority], runner)
	}

	ordered := make([][]*RunnerWithInfo, 0, len(groups))
	for _, priority := range slices.Sorted(maps.Keys(groups)) {
		ordered = append(ordered, groups[priority])
	}

	return ordered
}

func (r *Application) SetBuilder(builder foundation.ApplicationBuilder) foundation.Application {
	r.builder = builder.(*ApplicationBuilder)

	return r
}

// CommandsFilter returns the positive-list of command signatures to keep.
// nil means no filter is applied. Used by the console service provider so
// its own command batch goes through the same gate as the framework's
// defaultCommands batch.
func (r *Application) CommandsFilter() []string {
	return r.commandsFilter
}

// CommandHooks returns the hooks that run around Artisan commands, keyed by
// command signature or '*' pattern. nil means no hooks. Used by the console
// service provider, which passes the map to the Artisan application when it
// constructs it, the same way CommandsFilter reaches its command batch.
func (r *Application) CommandHooks() map[string][]contractsconsole.CommandHook {
	return r.commandHooks
}

func (r *Application) SetJson(j foundation.Json) {
	if j != nil {
		r.json = j
	}
}

func (r *Application) SetLocale(ctx context.Context, locale string) context.Context {
	lang := r.MakeLang(ctx)
	if lang == nil {
		color.Errorln("Lang facade not initialized.")
		return ctx
	}

	return lang.SetLocale(locale)
}

func (r *Application) Shutdown() error {
	r.cancel()

	i := 0
	for {
		running := ""
		for _, runner := range r.runnersToRun {
			if runner.running.Load() {
				running = runner.signature
				break
			}
		}

		if running == "" {
			break
		}
		if i > 100 {
			return fmt.Errorf("timeout waiting for runner %s to stop", running)
		}

		i++

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (r *Application) Version() string {
	return support.Version
}

func (r *Application) BasePath(paths ...string) string {
	return path.Base(paths...)
}

func (r *Application) BootstrapPath(paths ...string) string {
	return path.Bootstrap(paths...)
}

func (r *Application) ConfigPath(paths ...string) string {
	return path.Config(paths...)
}

func (r *Application) ModelPath(paths ...string) string {
	return path.Model(paths...)
}

func (r *Application) DatabasePath(paths ...string) string {
	return path.Database(paths...)
}

func (r *Application) CurrentLocale(ctx context.Context) string {
	lang := r.MakeLang(ctx)
	if lang == nil {
		color.Errorln("Lang facade not initialized.")
		return ""
	}

	return lang.CurrentLocale()
}

func (r *Application) ExecutablePath(paths ...string) string {
	return path.Executable(paths...)
}

func (r *Application) FacadesPath(paths ...string) string {
	return path.Facade(paths...)
}

func (r *Application) LangPath(paths ...string) string {
	return path.Lang(paths...)
}

func (r *Application) Path(paths ...string) string {
	return path.App(paths...)
}

func (r *Application) PublicPath(paths ...string) string {
	return path.Public(paths...)
}

func (r *Application) ResourcePath(paths ...string) string {
	return path.Resource(paths...)
}

func (r *Application) StoragePath(paths ...string) string {
	return path.Storage(paths...)
}

func (r *Application) addPublishGroup(group string, paths map[string]string) {
	if _, exist := r.publishGroups[group]; !exist {
		r.publishGroups[group] = make(map[string]string)
	}

	maps.Copy(r.publishGroups[group], paths)
}

func (r *Application) bootArtisan() {
	artisanFacade := r.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln(errors.ConsoleFacadeNotSet.Error())
		return
	}

	_ = artisanFacade.Run(os.Args, true)
}

func (r *Application) configureCallback() {
	if r.builder.callback != nil {
		r.builder.callback()
	}
}

func (r *Application) configureCommands() {
	var commands []contractsconsole.Command
	if r.builder.commands != nil {
		commands = r.builder.commands()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesCommands](r.modules) {
		commands = append(commands, module.Commands()...)
	}

	if len(commands) > 0 {
		r.registerCommands(commands)
	}
}

func (r *Application) configureCustomConfig() {
	if r.builder.config != nil {
		r.builder.config()
	}
}

// configureModuleConfig sets "modules.<name>" (replacing WithConfig's) and wraps
// app.lang_fs with module translations.
func (r *Application) configureModuleConfig() {
	translations := make(map[string]fs.FS)
	var configs []contractsmodular.Module
	for _, module := range r.modules {
		if _, ok := module.(contractsmodular.ProvidesConfig); ok {
			configs = append(configs, module)
		}
		if p, ok := module.(contractsmodular.ProvidesTranslations); ok {
			if fsys := p.Translations(); fsys != nil {
				translations[module.Name()] = fsys
			}
		}
	}
	if len(configs) == 0 && len(translations) == 0 {
		return
	}

	configFacade := r.MakeConfig()
	if configFacade == nil {
		color.Errorln(errors.ConfigFacadeNotSet.SetModule(errors.ModuleModular))
		return
	}

	for _, module := range configs {
		configFacade.Add("modules."+module.Name(), module.(contractsmodular.ProvidesConfig).Config())
	}

	if len(translations) > 0 {
		fallback, _ := configFacade.Get("app.lang_fs").(fs.FS)
		configFacade.Add("app.lang_fs", modular.NewLangFS(translations, fallback))
	}
}

func (r *Application) configureEventListeners() {
	var eventToListeners map[event.Event][]event.Listener
	if r.builder.eventToListeners != nil {
		eventToListeners = r.builder.eventToListeners()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesEvents](r.modules) {
		if eventToListeners == nil {
			eventToListeners = make(map[event.Event][]event.Listener)
		}
		// Append, so shared events keep every listener.
		for e, listeners := range module.Events() {
			eventToListeners[e] = append(eventToListeners[e], listeners...)
		}
	}

	if len(eventToListeners) > 0 {
		eventFacade := r.MakeEvent()
		if eventFacade == nil {
			color.Errorln("Event facade not found, please install it first: ./artisan package:install Event")
		} else {
			// WithEvents carries the deprecated event.Listener, Register is
			// the only entry accepting it.
			//nolint:staticcheck
			eventFacade.Register(eventToListeners)
		}
	}
}

func (r *Application) configureGrpc() {
	var (
		grpcClientCredentials   map[string]credentials.TransportCredentials
		grpcClientInterceptors  map[string][]grpc.UnaryClientInterceptor
		grpcServerCredentials   credentials.TransportCredentials
		grpcServerInterceptors  []grpc.UnaryServerInterceptor
		grpcClientStatsHandlers map[string][]stats.Handler
		grpcServerStatsHandlers []stats.Handler
	)

	if r.builder.grpcClientCredentials != nil {
		grpcClientCredentials = r.builder.grpcClientCredentials()
	}

	if r.builder.grpcClientInterceptors != nil {
		grpcClientInterceptors = r.builder.grpcClientInterceptors()
	}

	if r.builder.grpcServerCredentials != nil {
		grpcServerCredentials = r.builder.grpcServerCredentials()
	}

	if r.builder.grpcServerInterceptors != nil {
		grpcServerInterceptors = r.builder.grpcServerInterceptors()
	}

	if r.builder.grpcClientStatsHandlers != nil {
		grpcClientStatsHandlers = r.builder.grpcClientStatsHandlers()
	}

	if r.builder.grpcServerStatsHandlers != nil {
		grpcServerStatsHandlers = r.builder.grpcServerStatsHandlers()
	}

	if len(grpcClientCredentials) > 0 || len(grpcClientInterceptors) > 0 ||
		grpcServerCredentials != nil || len(grpcServerInterceptors) > 0 ||
		len(grpcClientStatsHandlers) > 0 || len(grpcServerStatsHandlers) > 0 {
		grpcFacade := r.MakeGrpc()
		if grpcFacade == nil {
			color.Errorln("gRPC facade not found, please install it first: ./artisan package:install Grpc")
		} else {
			if len(grpcClientCredentials) > 0 {
				grpcFacade.ClientCredentials(grpcClientCredentials)
			}
			if len(grpcClientInterceptors) > 0 {
				grpcFacade.UnaryClientInterceptorGroups(grpcClientInterceptors)
			}
			if len(grpcServerInterceptors) > 0 {
				grpcFacade.UnaryServerInterceptors(grpcServerInterceptors)
			}
			if grpcServerCredentials != nil {
				grpcFacade.ServerCredentials(grpcServerCredentials)
			}
			if len(grpcClientStatsHandlers) > 0 {
				grpcFacade.ClientStatsHandlerGroups(grpcClientStatsHandlers)
			}
			if len(grpcServerStatsHandlers) > 0 {
				grpcFacade.ServerStatsHandlers(grpcServerStatsHandlers)
			}
		}
	}
}

func (r *Application) configureJobs() {
	var jobs []queue.Job
	if r.builder.jobs != nil {
		jobs = r.builder.jobs()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesJobs](r.modules) {
		jobs = append(jobs, module.Jobs()...)
	}

	if len(jobs) > 0 {
		queueFacade := r.MakeQueue()
		if queueFacade == nil {
			color.Errorln("Queue facade not found, please install it first: ./artisan package:install Queue")
		} else {
			queueFacade.Register(jobs)
		}
	}
}

// configureMiddleware: driver defaults, then modules, then WithMiddleware (final say).
func (r *Application) configureMiddleware() {
	modules := modular.Providing[contractsmodular.ProvidesGlobalMiddleware](r.modules)
	if r.builder.middleware != nil || len(modules) > 0 {
		routeFacade := r.MakeRoute()
		if routeFacade == nil {
			color.Errorln("Route facade not found, please install it first: ./artisan package:install Route")
		} else {
			defaultGlobalMiddleware := routeFacade.GetGlobalMiddleware()
			middleware := configuration.NewMiddleware(defaultGlobalMiddleware)
			for _, module := range modules {
				middleware.Append(module.GlobalMiddleware()...)
			}
			if r.builder.middleware != nil {
				r.builder.middleware(middleware)
			}
			routeFacade.SetGlobalMiddleware(middleware.GetGlobalMiddleware())

			if recoveryHandler := middleware.GetRecover(); recoveryHandler != nil {
				routeFacade.Recover(recoveryHandler)
			}
		}
	}
}

func (r *Application) configureMigrations() {
	var migrations []schema.Migration
	if r.builder.migrations != nil {
		migrations = r.builder.migrations()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesMigrations](r.modules) {
		migrations = append(migrations, module.Migrations()...)
	}

	if len(migrations) > 0 {
		schemaFacade := r.MakeSchema()
		if schemaFacade == nil {
			color.Errorln("Schema facade not found, please install it first: ./artisan package:install Schema")
		} else {
			schemaFacade.Register(migrations)
		}
	}
}

// configureModules binds a fresh registry and keeps the sorted modules. It runs
// after WithConfig (modular.path) and before providers. Errors are reported.
func (r *Application) configureModules() {
	directory := modular.DefaultDirectory
	var modules []contractsmodular.Module
	if r.builder.modules != nil {
		modules = r.builder.modules()
	}
	if len(modules) > 0 {
		if config := r.MakeConfig(); config != nil {
			directory = config.GetString("modular.path", modular.DefaultDirectory)
		}
	}

	registry := modular.NewRegistry(directory)
	if err := registry.Register(modules...); err != nil {
		color.Errorln(err)
	}

	// Make caches instances; drop the previous registry on a rebuild.
	r.Fresh(binding.Modular)
	r.Instance(binding.Modular, registry)
	r.modules = registry.All()
}

func (r *Application) configurePaths() {
	if r.builder.paths != nil {
		paths := configuration.NewPaths()
		r.builder.paths(paths)
	}
}

// configureRoutes runs WithRouting, then module static files on the root router
// (no group middleware), then module routes in the modular.routes.* groups.
func (r *Application) configureRoutes() {
	if r.builder.routes != nil {
		r.builder.routes()
	}

	statics := modular.Providing[contractsmodular.ProvidesStatic](r.modules)
	web := modular.Providing[contractsmodular.ProvidesRoutes](r.modules)
	api := modular.Providing[contractsmodular.ProvidesApiRoutes](r.modules)
	if len(statics) == 0 && len(web) == 0 && len(api) == 0 {
		return
	}

	routeFacade := r.MakeRoute()
	if routeFacade == nil {
		color.Errorln("Route facade not found, please install it first: ./artisan package:install Route")
		return
	}

	r.registerModuleStatic(routeFacade)

	configFacade := r.MakeConfig()
	webPrefix, webMiddleware := moduleRouteGroup(configFacade, "modular.routes.web", "")
	apiPrefix, apiMiddleware := moduleRouteGroup(configFacade, "modular.routes.api", "api")
	for _, module := range web {
		routeFacade.Prefix(webPrefix).Middleware(webMiddleware...).Group(module.Routes)
	}
	for _, module := range api {
		routeFacade.Prefix(apiPrefix).Middleware(apiMiddleware...).Group(module.ApiRoutes)
	}
}

// registerModuleStatic serves static paths in module, then prefix, order. Duplicate
// or invalid ones are reported and skipped.
func (r *Application) registerModuleStatic(routeFacade route.Route) {
	registry := r.MakeModular()
	served := make(map[string]string)
	for _, module := range r.modules {
		p, ok := module.(contractsmodular.ProvidesStatic)
		if !ok {
			continue
		}

		paths := p.Static()
		for _, prefix := range slices.Sorted(maps.Keys(paths)) {
			cleaned, ok := modular.CleanPrefix(prefix)
			if !ok {
				color.Errorln(errors.ModularStaticPrefixInvalid.Args(module.Name(), prefix))
				continue
			}
			if owner, ok := served[cleaned]; ok {
				color.Errorln(errors.ModularStaticDuplicate.Args(cleaned, module.Name(), owner))
				continue
			}

			root, _ := registry.Path(module.Name())
			fsys, err := modular.StaticFS(module.Name(), root, paths[prefix])
			if err != nil {
				color.Errorln(err)
				continue
			}

			served[cleaned] = module.Name()
			routeFacade.StaticFS(cleaned, http.FS(fsys))
		}
	}
}

// moduleRouteGroup reads a route group's prefix and middleware.
func moduleRouteGroup(configFacade contractsconfig.Config, key, defaultPrefix string) (string, []contractshttp.Middleware) {
	if configFacade == nil {
		return defaultPrefix, nil
	}

	middleware, _ := configFacade.Get(key + ".middleware").([]contractshttp.Middleware)

	return configFacade.GetString(key+".prefix", defaultPrefix), middleware
}

func (r *Application) configureRunners() {
	disabledRunners := []string{}
	if config := r.MakeConfig(); config != nil {
		disabledRunners = config.GetStringSlice("app.disabled_runners", []string{})
	}

	for _, serviceProvider := range r.providerRepository.GetBooted() {
		if serviceProviderWithRunners, ok := serviceProvider.(foundation.ServiceProviderWithRunners); ok {
			for _, runner := range serviceProviderWithRunners.Runners(r) {
				signature := runner.Signature()
				if slices.Contains(r.bootedRunners, signature) {
					continue
				}

				r.bootedRunners = append(r.bootedRunners, signature)

				if !r.isRunnerDisabled(signature, disabledRunners) && runner.ShouldRun() {
					r.runnersToRun = append(r.runnersToRun, &RunnerWithInfo{signature: signature, runner: runner})
				}
			}
		}
	}

	if r.builder.runners != nil {
		for _, runner := range r.builder.runners() {
			signature := runner.Signature()
			if slices.Contains(r.bootedRunners, signature) {
				continue
			}

			r.bootedRunners = append(r.bootedRunners, signature)

			if !r.isRunnerDisabled(signature, disabledRunners) && runner.ShouldRun() {
				r.runnersToRun = append(r.runnersToRun, &RunnerWithInfo{signature: signature, runner: runner})
			}
		}
	}
}

func (r *Application) isRunnerDisabled(signature string, disabledRunners []string) bool {
	for _, pattern := range disabledRunners {
		matched, err := stdpath.Match(pattern, signature)
		if err != nil {
			if log := r.MakeLog(); log != nil {
				log.Warning(errors.AppInvalidDisabledRunnersPattern.Args(pattern, err).Error())
			}
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (r *Application) configureSchedule() {
	var events []schedule.Event
	if r.builder.schedule != nil {
		events = r.builder.schedule()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesSchedule](r.modules) {
		events = append(events, module.Schedule()...)
	}

	if len(events) > 0 {
		scheduleFacade := r.MakeSchedule()
		if scheduleFacade == nil {
			color.Errorln("Schedule facade not found, please install it first: ./artisan package:install Schedule")
		} else {
			scheduleFacade.Register(events)
		}
	}
}

func (r *Application) configureSeeders() {
	var seeders []seeder.Seeder
	if r.builder.seeders != nil {
		seeders = r.builder.seeders()
	}
	for _, module := range modular.Providing[contractsmodular.ProvidesSeeders](r.modules) {
		seeders = append(seeders, module.Seeders()...)
	}

	if len(seeders) > 0 {
		seederFacade := r.MakeSeeder()
		if seederFacade == nil {
			color.Errorln("Seeder facade not found, please install it first: ./artisan package:install Seeder")
		} else {
			seederFacade.Register(seeders)
		}
	}
}

func (r *Application) configureServiceProviders() {
	var providers []foundation.ServiceProvider
	if r.builder.configuredServiceProviders != nil {
		providers = r.builder.configuredServiceProviders()
	}
	for _, module := range modular.Providing[foundation.ProvidesServices](r.modules) {
		providers = append(providers, module.ServiceProviders()...)
	}

	if len(providers) > 0 {
		r.providerRepository.Add(providers)
	}
}

// registerModuleResources registers module views under "<name>/", after providers
// boot and before configureViews compiles them.
func (r *Application) registerModuleResources() {
	modules := modular.Providing[contractsmodular.ProvidesViews](r.modules)
	if len(modules) == 0 {
		return
	}

	viewFacade := r.MakeView()
	if viewFacade == nil {
		color.Errorln("View facade not found, please install it first: ./artisan package:install View")
		return
	}

	for _, module := range r.modules {
		p, ok := module.(contractsmodular.ProvidesViews)
		if !ok {
			continue
		}

		fsys := p.Views()
		if fsys == nil {
			// LoadViewsFromFS panics on a nil filesystem.
			color.Errorln(errors.ModularViewsFSRequired.Args(module.Name()))
			continue
		}

		viewFacade.LoadViewsFromFS(modular.NewPrefixFS(module.Name(), fsys), ".")
	}
}

// configureViews hands the application's template functions to the view registry
// and asks the route driver to compile the view set. It runs after every service
// provider has booted, so views and functions registered in a provider's Boot
// are part of the compiled set regardless of provider order. A compile error is
// reported here but does not stop Build(): Artisan commands build the
// application too. The driver keeps the error and returns it from its first
// serve call, so an HTTP process still fails before it serves without a renderer.
func (r *Application) configureViews() {
	if r.builder.viewFuncs != nil {
		if funcs := r.builder.viewFuncs(); len(funcs) > 0 {
			viewFacade := r.MakeView()
			if viewFacade == nil {
				color.Errorln("View facade not found, please install it first: ./artisan package:install View")
			} else {
				viewFacade.AddFuncs(funcs)
			}
		}
	}

	routeFacade := r.MakeRoute()
	if routeFacade == nil {
		return
	}

	if compiler, ok := routeFacade.(contractsview.Compiler); ok {
		if err := compiler.CompileViews(); err != nil {
			color.Errorln(errors.ViewCompileFailed.Args(err).SetModule(errors.ModuleView).Error())
		}
	}
}

func (r *Application) configureValidation() {
	var (
		rules   []validation.Rule
		filters []validation.Filter
	)

	if r.builder.rules != nil {
		rules = r.builder.rules()
	}

	if r.builder.filters != nil {
		filters = r.builder.filters()
	}

	if len(rules) > 0 || len(filters) > 0 {
		validationFacade := r.MakeValidation()
		if validationFacade == nil {
			color.Errorln("Validation facade not found, please install it first: ./artisan package:install Validation")
		} else {
			if len(rules) > 0 {
				if err := validationFacade.AddRules(rules); err != nil {
					color.Errorf("add validation rules error: %+v\n", err)
				}
			}
			if len(filters) > 0 {
				if err := validationFacade.AddFilters(filters); err != nil {
					color.Errorf("add validation filters error: %+v\n", err)
				}
			}
		}
	}
}

func (r *Application) defaultCommands() []contractsconsole.Command {
	commands := []contractsconsole.Command{
		console.NewAboutCommand(r),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewProviderMakeCommand(),
		console.NewPackageInstallCommand(binding.Bindings, r.MakeProcess(), r.Json()),
		console.NewPackageUninstallCommand(binding.Bindings, r.MakeProcess(), r.Json()),
		console.NewVendorPublishCommand(r.publishes, r.publishGroups),
		modularconsole.NewInitCommand(r.MakeConfig()),
	}

	if r.MakeRoute() != nil {
		storage := r.MakeStorage()
		view := r.MakeView()
		hash := r.MakeHash()
		cache := r.MakeCache()
		config := r.MakeConfig()
		maintenance := console.NewMaintenanceMode(config, cache, storage)
		commands = append(commands, console.NewUpCommand(maintenance), console.NewDownCommand(view, hash, maintenance))
	}

	return commands
}

func (r *Application) baseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
		&frameworkconsole.ServiceProvider{},
		&process.ServiceProvider{},
	}
}

func (r *Application) registerCommands(commands []contractsconsole.Command) {
	artisanFacade := r.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln(errors.ConsoleFacadeNotSet.Error())
		return
	}

	artisanFacade.Register(frameworkconsole.FilterCommandsByAllowlist(commands, r.commandsFilter))
}

func (r *Application) setTimezone() {
	configFacade := r.MakeConfig()
	if configFacade == nil {
		color.Warningln(errors.ConfigFacadeNotSet.Error())
		carbon.SetTimezone(carbon.UTC)
		return
	}

	carbon.SetTimezone(configFacade.GetString("app.timezone", carbon.UTC))
}

func setEnv() {
	args := os.Args
	isDebugBinary := strings.Contains(args[0], "__debug")
	isTestDebugBinary := isDebugBinary && slices.ContainsFunc(args[1:], func(arg string) bool {
		return strings.HasPrefix(arg, "-test.")
	})

	if strings.HasSuffix(args[0], ".test") ||
		strings.HasSuffix(args[0], ".test.exe") ||
		isTestDebugBinary {
		support.RuntimeMode = support.RuntimeTest
		support.DontVerifyAppKey = true
	} else {
		if len(args) >= 2 {
			for _, arg := range args[1:] {
				if arg == "artisan" {
					support.RuntimeMode = support.RuntimeArtisan

					if len(args) == 2 {
						// Run go run . artisan without any command
						support.DontVerifyAppKey = true
					}
				}

				support.DontVerifyAppKey = support.DontVerifyAppKey || slices.Contains(support.DontVerifyAppKeyWhitelist, arg)
			}
		}
	}

	envFilePath := getEnvFilePath()
	if support.RuntimeMode == support.RuntimeTest {
		var (
			relativePath string
			envExist     bool
			goModExist   bool
			testEnv      = envFilePath
			wg           sync.WaitGroup
		)

		wg.Add(2)
		go func() {
			defer wg.Done()

			for range 10 {
				if _, err := os.Stat(testEnv); err == nil {
					envExist = true
					break
				} else {
					testEnv = filepath.Join("..", testEnv)
				}
			}

		}()

		go func() {
			defer wg.Done()

			for range 10 {
				if _, err := os.Stat(filepath.Join(relativePath, "go.mod")); err == nil {
					goModExist = true
					break
				} else {
					relativePath = filepath.Join("..", relativePath)
				}
			}
		}()

		wg.Wait()

		if envExist {
			envFilePath = testEnv
		}
		if goModExist {
			support.RelativePath = relativePath
		}
	}

	support.EnvFilePath = envFilePath
}

func setRootPath() {
	support.RootPath = env.CurrentAbsolutePath()
}

func getEnvFilePath() string {
	envFilePath := ".env"
	args := os.Args
	for index, arg := range args {
		if path, ok := strings.CutPrefix(arg, "--env="); ok && len(path) > 0 {
			envFilePath = path
			break
		}

		if path, ok := strings.CutPrefix(arg, "-env="); ok && len(path) > 0 {
			envFilePath = path
			break
		}

		if path, ok := strings.CutPrefix(arg, "-e="); ok && len(path) > 0 {
			envFilePath = path
			break
		}

		if arg == "--env" || arg == "-env" || arg == "-e" {
			if index+1 < len(args) && !strings.HasPrefix(args[index+1], "-") {
				envFilePath = args[index+1]
				break
			}
		}
	}

	return envFilePath
}
