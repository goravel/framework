package foundation

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/config"
	consolecontract "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
)

var (
	App foundation.Application
)

var _ = flag.String("env", ".env", "custom .env path")

func init() {
	setEnv()
	setRootPath()

	app := &Application{
		Container:     NewContainer(),
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}
	app.registerBaseServiceProviders()
	app.bootBaseServiceProviders()
	app.SetJson(json.NewJson())
	App = app
}

type Application struct {
	foundation.Container
	publishes     map[string]map[string]string
	publishGroups map[string]map[string]string
	json          foundation.Json
}

func NewApplication() foundation.Application {
	return App
}

// Boot Register and bootstrap configured service providers.
func (app *Application) Boot() {
	app.registerConfiguredServiceProviders()
	app.bootConfiguredServiceProviders()
	app.registerCommands([]consolecontract.Command{
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewVendorPublishCommand(app.publishes, app.publishGroups),
	})
	app.setTimezone()
	app.bootArtisan()
}

func (app *Application) Commands(commands []consolecontract.Command) {
	app.registerCommands(commands)
}

func (app *Application) Path(path ...string) string {
	path = append([]string{"app"}, path...)
	return filepath.Join(path...)
}

func (app *Application) BasePath(path ...string) string {
	return filepath.Join(path...)
}

func (app *Application) ConfigPath(path ...string) string {
	path = append([]string{"config"}, path...)
	return filepath.Join(path...)
}

func (app *Application) DatabasePath(path ...string) string {
	path = append([]string{"database"}, path...)
	return filepath.Join(path...)
}

func (app *Application) StoragePath(path ...string) string {
	path = append([]string{"storage"}, path...)
	return filepath.Join(path...)
}

func (app *Application) LangPath(path ...string) string {
	defaultPath := "lang"
	if configFacade := app.MakeConfig(); configFacade != nil {
		defaultPath = configFacade.GetString("app.lang_path", defaultPath)
	}

	path = append([]string{defaultPath}, path...)
	return filepath.Join(path...)
}

func (app *Application) PublicPath(path ...string) string {
	path = append([]string{"public"}, path...)
	return filepath.Join(path...)
}

func (app *Application) ExecutablePath(path ...string) string {
	path = append([]string{support.RootPath}, path...)
	return filepath.Join(path...)
}

func (app *Application) Publishes(packageName string, paths map[string]string, groups ...string) {
	app.ensurePublishArrayInitialized(packageName)

	for key, value := range paths {
		app.publishes[packageName][key] = value
	}

	for _, group := range groups {
		app.addPublishGroup(group, paths)
	}
}

func (app *Application) Version() string {
	return support.Version
}

func (app *Application) CurrentLocale(ctx context.Context) string {
	lang := app.MakeLang(ctx)
	if lang == nil {
		color.Red().Println("Error: Lang facade not initialized.")
		return ""
	}

	return lang.CurrentLocale()
}

func (app *Application) SetLocale(ctx context.Context, locale string) context.Context {
	lang := app.MakeLang(ctx)
	if lang == nil {
		color.Red().Println("Error: Lang facade not initialized.")
		return ctx
	}

	return lang.SetLocale(locale)
}

func (app *Application) SetJson(j foundation.Json) {
	if j != nil {
		app.json = j
	}
}

func (app *Application) GetJson() foundation.Json {
	return app.json
}

func (app *Application) IsLocale(ctx context.Context, locale string) bool {
	return app.CurrentLocale(ctx) == locale
}

func (app *Application) ensurePublishArrayInitialized(packageName string) {
	if _, exist := app.publishes[packageName]; !exist {
		app.publishes[packageName] = make(map[string]string)
	}
}

func (app *Application) addPublishGroup(group string, paths map[string]string) {
	if _, exist := app.publishGroups[group]; !exist {
		app.publishGroups[group] = make(map[string]string)
	}

	for key, value := range paths {
		app.publishGroups[group][key] = value
	}
}

// bootArtisan Boot artisan command.
func (app *Application) bootArtisan() {
	if artisanFacade := app.MakeArtisan(); artisanFacade != nil {
		artisanFacade.Run(os.Args, true)
	}
}

// getBaseServiceProviders Get base service providers.
func (app *Application) getBaseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
	}
}

// getConfiguredServiceProviders Get configured service providers.
func (app *Application) getConfiguredServiceProviders() []foundation.ServiceProvider {
	if configFacade := app.MakeConfig(); configFacade != nil {
		return configFacade.Get("app.providers").([]foundation.ServiceProvider)
	}

	return []foundation.ServiceProvider{}
}

// registerBaseServiceProviders Register base service providers.
func (app *Application) registerBaseServiceProviders() {
	app.registerServiceProviders(app.getBaseServiceProviders())
}

// bootBaseServiceProviders Bootstrap base service providers.
func (app *Application) bootBaseServiceProviders() {
	app.bootServiceProviders(app.getBaseServiceProviders())
}

// registerConfiguredServiceProviders Register configured service providers.
func (app *Application) registerConfiguredServiceProviders() {
	app.registerServiceProviders(app.getConfiguredServiceProviders())
}

// bootConfiguredServiceProviders Bootstrap configured service providers.
func (app *Application) bootConfiguredServiceProviders() {
	app.bootServiceProviders(app.getConfiguredServiceProviders())
}

// registerServiceProviders Register service providers.
func (app *Application) registerServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register(app)
	}
}

// bootServiceProviders Bootstrap service providers.
func (app *Application) bootServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot(app)
	}
}

func (app *Application) registerCommands(commands []consolecontract.Command) {
	if artisanFacade := app.MakeArtisan(); artisanFacade != nil {
		artisanFacade.Register(commands)
	}
}

func (app *Application) setTimezone() {
	if configFacade := app.MakeConfig(); configFacade != nil {
		carbon.SetTimezone(configFacade.GetString("app.timezone", carbon.UTC))
	}
}

func setEnv() {
	args := os.Args
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		support.Env = support.EnvTest
	}
	if len(args) >= 2 {
		for _, arg := range args[1:] {
			if arg == "artisan" {
				support.Env = support.EnvArtisan
			}
			if arg == "key:generate" {
				support.IsKeyGenerateCommand = true
			}
		}
	}

	env := getEnvPath()
	if support.Env == support.EnvTest {
		var (
			relativePath string
			envExist     bool
			testEnv      = env
		)

		for i := 0; i < 50; i++ {
			if _, err := os.Stat(testEnv); err == nil {
				envExist = true

				break
			} else {
				testEnv = filepath.Join("../", testEnv)
				relativePath = filepath.Join("../", relativePath)
			}
		}

		if envExist {
			env = testEnv
			support.RelativePath = relativePath
		}
	}

	support.EnvPath = env
}

func setRootPath() {
	support.RootPath = getCurrentAbsolutePath()
}

func getEnvPath() string {
	envPath := ".env"
	args := os.Args
	for index, arg := range args {
		if strings.HasPrefix(arg, "--env=") {
			if path := strings.TrimPrefix(arg, "--env="); path != "" {
				envPath = path
				break
			}
		}
		if strings.HasPrefix(arg, "-env=") {
			if path := strings.TrimPrefix(arg, "-env="); path != "" {
				envPath = path
				break
			}
		}
		if strings.HasPrefix(arg, "-e=") {
			if path := strings.TrimPrefix(arg, "-e="); path != "" {
				envPath = path
				break
			}
		}
		if arg == "--env" || arg == "-env" || arg == "-e" {
			if len(args) >= index+1 && !strings.HasPrefix(args[index+1], "-") {
				envPath = args[index+1]
				break
			}
		}
	}

	return envPath
}
