package foundation

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goravel/framework/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/console"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/env"
)

var (
	App foundation.Application
)

var _ = flag.String("env", support.EnvFilePath, "custom .env path")

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
	*Container
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
	app.registerCommands([]contractsconsole.Command{
		console.NewAboutCommand(app),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewPackageInstallCommand(),
		console.NewPackageUninstallCommand(),
		console.NewVendorPublishCommand(app.publishes, app.publishGroups),
	})
	app.setTimezone()
	app.bootArtisan()
}

func (app *Application) Commands(commands []contractsconsole.Command) {
	app.registerCommands(commands)
}

func (app *Application) Path(path ...string) string {
	path = append([]string{support.RelativePath, "app"}, path...)
	return app.absPath(path...)
}

func (app *Application) BasePath(path ...string) string {
	return app.absPath(path...)
}

func (app *Application) ConfigPath(path ...string) string {
	path = append([]string{support.RelativePath, "config"}, path...)
	return app.absPath(path...)
}

func (app *Application) DatabasePath(path ...string) string {
	path = append([]string{support.RelativePath, "database"}, path...)
	return app.absPath(path...)
}

func (app *Application) StoragePath(path ...string) string {
	path = append([]string{support.RelativePath, "storage"}, path...)
	return app.absPath(path...)
}

func (app *Application) ResourcePath(path ...string) string {
	path = append([]string{support.RelativePath, "resources"}, path...)
	return app.absPath(path...)
}

func (app *Application) LangPath(path ...string) string {
	defaultPath := "lang"
	if configFacade := app.MakeConfig(); configFacade != nil {
		defaultPath = configFacade.GetString("app.lang_path", defaultPath)
	}

	path = append([]string{support.RelativePath, defaultPath}, path...)
	return app.absPath(path...)
}

func (app *Application) PublicPath(path ...string) string {
	path = append([]string{support.RelativePath, "public"}, path...)
	return app.absPath(path...)
}

func (app *Application) ExecutablePath(path ...string) string {
	path = append([]string{support.RootPath}, path...)
	return app.absPath(path...)
}

func (app *Application) Publishes(packageName string, paths map[string]string, groups ...string) {
	if _, exist := app.publishes[packageName]; !exist {
		app.publishes[packageName] = make(map[string]string)
	}
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
		color.Errorln("Lang facade not initialized.")
		return ""
	}

	return lang.CurrentLocale()
}

func (app *Application) SetLocale(ctx context.Context, locale string) context.Context {
	lang := app.MakeLang(ctx)
	if lang == nil {
		color.Errorln("Lang facade not initialized.")
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

func (app *Application) About(section string, items []foundation.AboutItem) {
	console.AddAboutInformation(section, items...)
}

func (app *Application) IsLocale(ctx context.Context, locale string) bool {
	return app.CurrentLocale(ctx) == locale
}

// absPath ensures the returned path is absolute
func (app *Application) absPath(paths ...string) string {
	path := filepath.Join(paths...)
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
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
	artisanFacade := app.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping artisan command execution.")
		return
	}

	_ = artisanFacade.Run(os.Args, true)
}

// getBaseServiceProviders Get base service providers.
func (app *Application) getBaseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
	}
}

// getConfiguredServiceProviders Get configured service providers.
func (app *Application) getConfiguredServiceProviders() []foundation.ServiceProvider {
	configFacade := app.MakeConfig()
	if configFacade == nil {
		color.Warningln("config facade is not initialized. Skipping registering service providers.")
		return []foundation.ServiceProvider{}
	}

	providers, ok := configFacade.Get("app.providers").([]foundation.ServiceProvider)
	if !ok {
		color.Warningln("providers configuration is not of type []foundation.ServiceProvider. Skipping registering service providers.")
		return []foundation.ServiceProvider{}
	}
	return providers
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

func (app *Application) registerCommands(commands []contractsconsole.Command) {
	artisanFacade := app.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping command registration.")
		return
	}

	artisanFacade.Register(commands)
}

func (app *Application) setTimezone() {
	configFacade := app.MakeConfig()
	if configFacade == nil {
		color.Warningln("config facade is not initialized. Using default timezone UTC.")
		carbon.SetTimezone(carbon.UTC)
		return
	}

	carbon.SetTimezone(configFacade.GetString("app.timezone", carbon.UTC))
}

func setEnv() {
	args := os.Args
	if strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") {
		support.RuntimeMode = support.RuntimeTest
	}
	if len(args) >= 2 {
		for _, arg := range args[1:] {
			if arg == "artisan" {
				support.RuntimeMode = support.RuntimeArtisan
			}
			support.DontVerifyEnvFileExists = slices.Contains(support.DontVerifyEnvFileWhitelist, arg)
		}
	}

	envFilePath := getEnvFilePath()
	if support.RuntimeMode == support.RuntimeTest {
		var (
			relativePath string
			envExist     bool
			testEnv      = envFilePath
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
		if strings.HasPrefix(arg, "--env=") {
			if path := strings.TrimPrefix(arg, "--env="); path != "" {
				envFilePath = path
				break
			}
		}
		if strings.HasPrefix(arg, "-env=") {
			if path := strings.TrimPrefix(arg, "-env="); path != "" {
				envFilePath = path
				break
			}
		}
		if strings.HasPrefix(arg, "-e=") {
			if path := strings.TrimPrefix(arg, "-e="); path != "" {
				envFilePath = path
				break
			}
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
