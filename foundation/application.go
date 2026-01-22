package foundation

import (
	"context"
	"flag"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/goravel/framework/config"
	frameworkconsole "github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

var App foundation.Application
var _ = flag.String("env", support.EnvFilePath, "custom .env path")

func init() {
	setEnv()
	setRootPath()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	app := &Application{
		Container:     NewContainer(),
		ctx:           ctx,
		cancel:        cancel,
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
		runnerWg:      sync.WaitGroup{},
	}

	app.providerRepository = NewProviderRepository()
	App = app

	baseProviders := app.getBaseServiceProviders()
	app.providerRepository.Add(baseProviders)
	app.providerRepository.Register(app)

	app.SetJson(json.New())
}

type Application struct {
	*Container
	ctx                context.Context
	cancel             context.CancelFunc
	providerRepository foundation.ProviderRepository
	publishes          map[string]map[string]string
	publishGroups      map[string]map[string]string
	json               foundation.Json
	bootedRunners      []string
	runnerWg           sync.WaitGroup
}

func NewApplication() foundation.Application {
	return App
}

func (r *Application) AddServiceProviders(providers []foundation.ServiceProvider) {
	r.providerRepository.Add(providers)
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

	r.registerCommands([]contractsconsole.Command{
		console.NewAboutCommand(r),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewProviderMakeCommand(),
		console.NewPackageInstallCommand(binding.Bindings, r.GetJson()),
		console.NewPackageUninstallCommand(binding.Bindings, r.GetJson()),
		console.NewVendorPublishCommand(r.publishes, r.publishGroups),
	})
	r.bootArtisan()
}

func (r *Application) BootServiceProviders() {
	r.providerRepository.Boot(r)

	r.registerCommands([]contractsconsole.Command{
		console.NewAboutCommand(r),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewProviderMakeCommand(),
		console.NewPackageInstallCommand(binding.Bindings, r.GetJson()),
		console.NewPackageUninstallCommand(binding.Bindings, r.GetJson()),
		console.NewVendorPublishCommand(r.publishes, r.publishGroups),
	})
	r.bootArtisan()
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
	r.Fresh()
	r.providerRepository.Reset()
	r.Boot()
}

func (r *Application) RegisterServiceProviders() {
	r.providerRepository.LoadFromConfig(r.MakeConfig())
	clear(r.publishes)
	clear(r.publishGroups)

	r.setTimezone()

	r.providerRepository.Register(r)
}

func (r *Application) Start(runners ...foundation.Runner) foundation.Application {
	type RunnerWithInfo struct {
		signature string
		runner    foundation.Runner
		running   bool
	}

	var runnersToRun []*RunnerWithInfo

	for _, serviceProvider := range r.providerRepository.GetBooted() {
		if serviceProviderWithRunners, ok := serviceProvider.(foundation.ServiceProviderWithRunners); ok {
			for _, runner := range serviceProviderWithRunners.Runners(r) {
				signature := runner.Signature()
				if slices.Contains(r.bootedRunners, signature) {
					continue
				}

				r.bootedRunners = append(r.bootedRunners, signature)

				if runner.ShouldRun() {
					runnersToRun = append(runnersToRun, &RunnerWithInfo{signature: signature, runner: runner, running: false})
				}
			}
		}
	}

	for _, runner := range runners {
		signature := runner.Signature()
		if slices.Contains(r.bootedRunners, signature) {
			continue
		}

		r.bootedRunners = append(r.bootedRunners, signature)

		if runner.ShouldRun() {
			runnersToRun = append(runnersToRun, &RunnerWithInfo{signature: signature, runner: runner, running: false})
		}
	}

	run := func(runner *RunnerWithInfo) {
		r.runnerWg.Add(1)

		go func() {
			runner.running = true
			if err := runner.runner.Run(); err != nil {
				r.runnerWg.Done()
				runner.running = false
				color.Errorf("failed to run %s: %v\n", runner.signature, err)
			}
		}()

		go func() {
			<-r.ctx.Done()
			if !runner.running {
				return
			}

			if err := runner.runner.Shutdown(); err != nil {
				color.Errorf("failed to shutdown %s: %v\n", runner.signature, err)
			}

			r.runnerWg.Done()
		}()
	}

	for _, runner := range runnersToRun {
		run(runner)
	}

	return r
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

func (r *Application) Shutdown() {
	r.cancel()
}

func (r *Application) Version() string {
	return support.Version
}

func (r *Application) Wait() {
	r.runnerWg.Wait()
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

func (r *Application) getBaseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
		&frameworkconsole.ServiceProvider{},
	}
}

func (r *Application) registerCommands(commands []contractsconsole.Command) {
	artisanFacade := r.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln(errors.ConsoleFacadeNotSet.Error())
		return
	}

	artisanFacade.Register(commands)
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

	if strings.HasSuffix(args[0], ".test") ||
		strings.HasSuffix(args[0], ".test.exe") ||
		strings.Contains(args[0], "__debug") {
		support.RuntimeMode = support.RuntimeTest
		support.DontVerifyEnvFileExists = true
	} else {
		if len(args) >= 2 {
			for _, arg := range args[1:] {
				if arg == "artisan" {
					support.RuntimeMode = support.RuntimeArtisan
				}
				support.DontVerifyEnvFileExists = support.DontVerifyEnvFileExists || slices.Contains(support.DontVerifyEnvFileWhitelist, arg)
			}
		}
	}

	envFilePath := getEnvFilePath()
	if support.RuntimeMode == support.RuntimeTest {
		var (
			relativePath string
			envExist     bool
			testEnv      = envFilePath
		)

		for range 50 {
			if _, err := os.Stat(testEnv); err == nil {
				envExist = true

				break
			} else {
				testEnv = filepath.Join("..", testEnv)
				relativePath = filepath.Join("..", relativePath)
			}
		}

		if envExist {
			envFilePath = testEnv
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
			if len(args) >= index+1 && !strings.HasPrefix(args[index+1], "-") {
				envFilePath = args[index+1]
				break
			}
		}
	}

	return envFilePath
}
