package foundation

import (
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/goravel/framework/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
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
	configuredServiceProviders []foundation.ServiceProvider
	publishes                  map[string]map[string]string
	publishGroups              map[string]map[string]string
	json                       foundation.Json
}

func NewApplication() foundation.Application {
	return App
}

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
	}
}

func (r *Application) getConfiguredServiceProviders() []foundation.ServiceProvider {
	if len(r.configuredServiceProviders) > 0 {
		return r.configuredServiceProviders
	}

	configFacade := r.MakeConfig()
	if configFacade == nil {
		color.Warningln(errors.ConfigFacadeNotSet.Error())
		return []foundation.ServiceProvider{}
	}

	providers, ok := configFacade.Get("app.providers").([]foundation.ServiceProvider)
	if !ok {
		color.Warningln(errors.ConsoleProvidersNotArray.Error())
		return []foundation.ServiceProvider{}
	}

	r.configuredServiceProviders = sortConfiguredServiceProviders(providers)

	return r.configuredServiceProviders
}

func (r *Application) registerBaseServiceProviders() {
	r.registerServiceProviders(r.getBaseServiceProviders())
}

func (r *Application) bootBaseServiceProviders() {
	r.bootServiceProviders(r.getBaseServiceProviders())
}

func (r *Application) registerConfiguredServiceProviders() {
	r.registerServiceProviders(r.getConfiguredServiceProviders())
}

func (r *Application) bootConfiguredServiceProviders() {
	r.bootServiceProviders(r.getConfiguredServiceProviders())
}

func (r *Application) registerServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Register(r)
	}
}

func (r *Application) bootServiceProviders(serviceProviders []foundation.ServiceProvider) {
	for _, serviceProvider := range serviceProviders {
		serviceProvider.Boot(r)
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

		for range 50 {
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

func sortConfiguredServiceProviders(providers []foundation.ServiceProvider) []foundation.ServiceProvider {
	if len(providers) == 0 {
		return providers
	}

	// Helper function to get binding names from a provider
	getBindings := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Bindings() []string }); ok {
			return p.Bindings()
		}
		return []string{}
	}

	// Helper function to get dependencies from a provider
	getDependencies := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Dependencies() []string }); ok {
			return p.Dependencies()
		}
		return []string{}
	}

	// Helper function to get provide-for bindings from a provider
	getProvideFor := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ ProvideFor() []string }); ok {
			return p.ProvideFor()
		}
		return []string{}
	}

	// Create a map for quick lookup of providers by their binding names
	providerMap := make(map[string]foundation.ServiceProvider)
	bindingToProvider := make(map[string]foundation.ServiceProvider)
	for _, provider := range providers {
		for _, binding := range getBindings(provider) {
			providerMap[binding] = provider
			bindingToProvider[binding] = provider
		}
	}

	// Create adjacency list for dependency graph
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize inDegree for all providers
	for _, provider := range providers {
		for _, binding := range getBindings(provider) {
			inDegree[binding] = 0
		}
	}

	// Build the dependency graph using both Dependencies and ProvideFor
	for _, provider := range providers {
		for _, binding := range getBindings(provider) {
			// Add dependencies (this provider depends on others)
			for _, dep := range getDependencies(provider) {
				if _, exists := providerMap[dep]; exists {
					graph[dep] = append(graph[dep], binding)
					inDegree[binding]++
				}
			}

			// Add provide-for relationships (others depend on this provider)
			for _, provideFor := range getProvideFor(provider) {
				if _, exists := providerMap[provideFor]; exists {
					graph[binding] = append(graph[binding], provideFor)
					inDegree[provideFor]++
				}
			}
		}
	}

	// Topological sort using Kahn's algorithm
	var queue []string
	var result []string

	// Add all nodes with in-degree 0 to queue
	for binding, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, binding)
		}
	}

	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree of all neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// If we couldn't process all nodes, there's a cycle
	if len(result) != len(inDegree) {
		// Detect and report the cycle
		cycle := detectCycle(graph, bindingToProvider)
		if len(cycle) > 0 {
			panic(errors.ServiceProviderCycle.Args(strings.Join(cycle, " -> ")))
		}
	}

	// Convert back to service providers in sorted order
	sortedProviders := make([]foundation.ServiceProvider, 0, len(providers))
	used := make(map[foundation.ServiceProvider]bool)

	for _, binding := range result {
		provider := providerMap[binding]
		if !used[provider] {
			sortedProviders = append(sortedProviders, provider)
			used[provider] = true
		}
	}

	// Add any remaining providers that weren't in the dependency graph
	for _, provider := range providers {
		if !used[provider] {
			sortedProviders = append(sortedProviders, provider)
		}
	}

	return sortedProviders
}

// detectCycle detects a cycle in the dependency graph and returns a descriptive error message
func detectCycle(graph map[string][]string, bindingToProvider map[string]foundation.ServiceProvider) []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)
	cycle := make([]string, 0)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found a cycle, collect the cycle path
				cycleStart := -1
				for i, p := range path {
					if p == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle = append(cycle, path[cycleStart:]...)
					cycle = append(cycle, neighbor)
				}
				return true
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
		return false
	}

	// Find cycles starting from each unvisited node
	// Sort nodes to ensure consistent behavior when multiple cycles exist
	var nodes []string
	for node := range graph {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	for _, node := range nodes {
		if !visited[node] {
			if dfs(node) {
				break
			}
		}
	}

	// Build error message with provider names
	if len(cycle) > 0 {
		var cycleProviders []string
		providerSet := make(map[string]struct{})

		for _, binding := range cycle {
			if provider, exists := bindingToProvider[binding]; exists {
				providerName := fmt.Sprintf("%T", provider)
				cycleProviders = append(cycleProviders, providerName)
				providerSet[providerName] = struct{}{}
			}
		}

		// If the cycle is a self-loop (A -> A), only show as 'A -> A' if there are two unique providers, otherwise just 'A'
		if len(cycleProviders) == 2 && cycleProviders[0] == cycleProviders[1] {
			if len(providerSet) == 1 && len(cycle) > 2 {
				// This is a missing mapping case, only one provider in the cycle
				return cycleProviders[0:1]
			}
		}

		return cycleProviders
	}

	return nil
}
