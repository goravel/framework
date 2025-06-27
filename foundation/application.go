package foundation

import (
	"context"
	"flag"
	"maps"
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
	app.SetJson(json.New())
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
func (r *Application) Boot() {
	r.setTimezone()
	r.registerConfiguredServiceProviders()
	r.bootConfiguredServiceProviders()
	r.registerCommands([]contractsconsole.Command{
		console.NewAboutCommand(r),
		console.NewEnvEncryptCommand(),
		console.NewEnvDecryptCommand(),
		console.NewTestMakeCommand(),
		console.NewPackageMakeCommand(),
		console.NewPackageInstallCommand(),
		console.NewPackageUninstallCommand(),
		console.NewVendorPublishCommand(r.publishes, r.publishGroups),
	})
	r.bootArtisan()
}

func (r *Application) Commands(commands []contractsconsole.Command) {
	r.registerCommands(commands)
}

func (r *Application) Path(path ...string) string {
	path = append([]string{support.RelativePath, "app"}, path...)
	return r.absPath(path...)
}

func (r *Application) BasePath(path ...string) string {
	return r.absPath(path...)
}

func (r *Application) ConfigPath(path ...string) string {
	path = append([]string{support.RelativePath, "config"}, path...)
	return r.absPath(path...)
}

func (r *Application) DatabasePath(path ...string) string {
	path = append([]string{support.RelativePath, "database"}, path...)
	return r.absPath(path...)
}

func (r *Application) StoragePath(path ...string) string {
	path = append([]string{support.RelativePath, "storage"}, path...)
	return r.absPath(path...)
}

func (r *Application) Refresh() {
	r.Fresh()
	r.Boot()
}

func (r *Application) ResourcePath(path ...string) string {
	path = append([]string{support.RelativePath, "resources"}, path...)
	return r.absPath(path...)
}

func (r *Application) LangPath(path ...string) string {
	defaultPath := "lang"
	if configFacade := r.MakeConfig(); configFacade != nil {
		defaultPath = configFacade.GetString("app.lang_path", defaultPath)
	}

	path = append([]string{support.RelativePath, defaultPath}, path...)
	return r.absPath(path...)
}

func (r *Application) PublicPath(path ...string) string {
	path = append([]string{support.RelativePath, "public"}, path...)
	return r.absPath(path...)
}

func (r *Application) ExecutablePath(path ...string) string {
	path = append([]string{support.RootPath}, path...)
	return r.absPath(path...)
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

func (r *Application) Version() string {
	return support.Version
}

func (r *Application) CurrentLocale(ctx context.Context) string {
	lang := r.MakeLang(ctx)
	if lang == nil {
		color.Errorln("Lang facade not initialized.")
		return ""
	}

	return lang.CurrentLocale()
}

func (r *Application) SetLocale(ctx context.Context, locale string) context.Context {
	lang := r.MakeLang(ctx)
	if lang == nil {
		color.Errorln("Lang facade not initialized.")
		return ctx
	}

	return lang.SetLocale(locale)
}

func (r *Application) SetJson(j foundation.Json) {
	if j != nil {
		r.json = j
	}
}

func (r *Application) GetJson() foundation.Json {
	return r.json
}

func (r *Application) About(section string, items []foundation.AboutItem) {
	console.AddAboutInformation(section, items...)
}

func (r *Application) IsLocale(ctx context.Context, locale string) bool {
	return r.CurrentLocale(ctx) == locale
}

// absPath ensures the returned path is absolute
func (r *Application) absPath(paths ...string) string {
	path := filepath.Join(paths...)
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func (r *Application) addPublishGroup(group string, paths map[string]string) {
	if _, exist := r.publishGroups[group]; !exist {
		r.publishGroups[group] = make(map[string]string)
	}

	maps.Copy(r.publishGroups[group], paths)
}

// bootArtisan Boot artisan command.
func (r *Application) bootArtisan() {
	artisanFacade := r.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping artisan command execution.")
		return
	}

	_ = artisanFacade.Run(os.Args, true)
}

// getBaseServiceProviders Get base service providers.
func (r *Application) getBaseServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&config.ServiceProvider{},
	}
}

// getConfiguredServiceProviders Get configured service providers.
func (r *Application) getConfiguredServiceProviders() []foundation.ServiceProvider {
	configFacade := r.MakeConfig()
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
func (r *Application) registerBaseServiceProviders() {
	r.registerServiceProviders(r.getBaseServiceProviders())
}

// bootBaseServiceProviders Bootstrap base service providers.
func (r *Application) bootBaseServiceProviders() {
	r.bootServiceProviders(r.getBaseServiceProviders())
}

// registerConfiguredServiceProviders Register configured service providers.
func (r *Application) registerConfiguredServiceProviders() {
	r.registerServiceProviders(r.getConfiguredServiceProviders())
}

// bootConfiguredServiceProviders Bootstrap configured service providers.
func (r *Application) bootConfiguredServiceProviders() {
	r.bootServiceProviders(r.getConfiguredServiceProviders())
}

// registerServiceProviders Register service providers.
func (r *Application) registerServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register(r)
	}
}

// bootServiceProviders Bootstrap service providers.
func (r *Application) bootServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot(r)
	}
}

func (r *Application) registerCommands(commands []contractsconsole.Command) {
	artisanFacade := r.MakeArtisan()
	if artisanFacade == nil {
		color.Warningln("Artisan Facade is not initialized. Skipping command registration.")
		return
	}

	artisanFacade.Register(commands)
}

func (r *Application) setTimezone() {
	configFacade := r.MakeConfig()
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
