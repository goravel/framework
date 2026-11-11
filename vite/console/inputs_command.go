package console

import (
	"encoding/json"
	"maps"
	stdpath "path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	contractsview "github.com/goravel/framework/contracts/view"
	contractsvite "github.com/goravel/framework/contracts/vite"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/view"
)

// Template functions whose string arguments name Vite entries and assets, as vite.Funcs registers them.
const (
	tagsFunction  = "vite"
	assetFunction = "viteAsset"
)

// Inputs is what goravel-vite-plugin reads: the entries passed to vite and the
// files passed to viteAsset, each once and sorted, the calls with computed
// arguments, the configured paths and the modules. The entries and assets include
// vite.inputs and vite.assets from config, so a build tool need not reach into config itself.
type Inputs struct {
	Inputs  []string       `json:"inputs"`
	Assets  []string       `json:"assets"`
	Dynamic []DynamicCall  `json:"dynamic"`
	Paths   Paths          `json:"paths"`
	Modules []ModuleInputs `json:"modules"`
}

// ModuleInputs is a module as the plugin needs it: its directory, entries and
// alias targets, all relative to the project root.
type ModuleInputs struct {
	Name    string            `json:"name"`
	Path    string            `json:"path"`
	Entries []string          `json:"entries"`
	Aliases map[string]string `json:"aliases"`
}

// Paths is the vite configuration as the plugin needs it: the public directory
// and hot file relative to the project root (the hot file absolute when outside
// it), the build directory relative to the public directory and the manifest
// relative to the build directory.
type Paths struct {
	PublicDirectory string `json:"public_directory"`
	BuildDirectory  string `json:"build_directory"`
	HotFile         string `json:"hot_file"`
	Manifest        string `json:"manifest"`
}

// DynamicCall names a template and the function it calls with a computed argument, e.g. {{ vite .entry }}.
type DynamicCall struct {
	Template string `json:"template"`
	Function string `json:"function"`
}

// InputsCommand prints the entries, assets and build paths the views and config
// declare, so a build tool need not repeat them.
type InputsCommand struct {
	view      contractsview.View
	vite      contractsvite.Vite
	config    contractsconfig.Config
	modules   contractsmodular.Registry
	buildInfo func() (*debug.BuildInfo, bool)
}

// NewInputsCommand takes the module registry, nil when the application has none.
func NewInputsCommand(views contractsview.View, vite contractsvite.Vite, config contractsconfig.Config, modules contractsmodular.Registry) *InputsCommand {
	return &InputsCommand{view: views, vite: vite, config: config, modules: modules, buildInfo: debug.ReadBuildInfo}
}

// Signature The name and signature of the console command.
func (r *InputsCommand) Signature() string {
	return "vite:inputs"
}

// Description The console command description.
func (r *InputsCommand) Description() string {
	return "Print the Vite entry points and assets the views and config declare, and the build paths, as JSON"
}

// Extend The console command extend.
func (r *InputsCommand) Extend() command.Extend {
	return command.Extend{
		Category: "vite",
	}
}

// Handle compiles the view set as the view registry does and prints one JSON object on a single line.
func (r *InputsCommand) Handle(ctx console.Context) error {
	if r.view == nil {
		return errors.ViewFacadeNotSet.SetModule(errors.ModuleVite)
	}
	if r.vite == nil {
		return errors.ViteFacadeNotSet
	}

	modules, err := r.moduleInputs()
	if err != nil {
		return err
	}

	tmpl, err := view.Compile(view.CompileOptions{
		FuncMap: r.view.Funcs(),
		Paths:   r.view.RegisteredViews(),
		FS:      r.view.RegisteredViewFS(),
	})
	if err != nil {
		return errors.ViewCompileFailed.Args(err).SetModule(errors.ModuleVite)
	}

	inputs := Inputs{
		Inputs:  r.configuredEntries("vite.inputs"),
		Assets:  r.configuredEntries("vite.assets"),
		Dynamic: []DynamicCall{},
		Paths:   r.paths(),
		Modules: modules,
	}
	collect(view.Calls(tmpl, tagsFunction), tagsFunction, &inputs.Inputs, &inputs.Dynamic)
	collect(view.Calls(tmpl, assetFunction), assetFunction, &inputs.Assets, &inputs.Dynamic)

	output, err := json.Marshal(inputs)
	if err != nil {
		return err
	}

	ctx.Line(string(output))

	return nil
}

// paths reports the configured directories, with the hot file relative to the
// project root when inside it, so the output is the same on every machine.
func (r *InputsCommand) paths() Paths {
	hotFile := r.vite.HotFile()
	if rel, err := filepath.Rel(paths.Abs(), hotFile); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		hotFile = rel
	}

	return Paths{
		PublicDirectory: filepath.ToSlash(support.Config.Paths.Public),
		BuildDirectory:  r.vite.BuildDirectory(),
		HotFile:         filepath.ToSlash(hotFile),
		Manifest:        r.vite.ManifestFilename(),
	}
}

// aliasKey is a Vite alias a module may declare: "@" and a name without slashes.
var aliasKey = regexp.MustCompile(`^@[A-Za-z0-9][A-Za-z0-9_-]*$`)

// moduleInputs lists every module with its entries and aliases made relative to the
// project root. It fails on an invalid or duplicate alias and on a third-party
// module whose published copy is missing or differs from the linked version.
func (r *InputsCommand) moduleInputs() ([]ModuleInputs, error) {
	inputs := []ModuleInputs{}
	if r.modules == nil {
		return inputs, nil
	}

	all := r.modules.All()
	info, _ := r.buildInfo()

	// "@<name>" belongs to its module, and every other key to the first module declaring it.
	owners := make(map[string]string, len(all))
	for _, module := range all {
		owners["@"+module.Name()] = module.Name()
	}

	for _, module := range all {
		name := module.Name()
		dir, _ := r.modules.Path(name)
		if err := checkPublished(module, info, dir); err != nil {
			return nil, err
		}

		input := ModuleInputs{Name: name, Path: dir, Entries: []string{}, Aliases: map[string]string{}}
		if p, ok := module.(contractsmodular.ProvidesViteEntries); ok {
			for _, entry := range p.ViteEntries() {
				input.Entries = append(input.Entries, stdpath.Join(dir, filepath.ToSlash(entry)))
			}
		}
		if p, ok := module.(contractsmodular.ProvidesViteAlias); ok {
			aliases := p.ViteAliases()
			for _, key := range slices.Sorted(maps.Keys(aliases)) {
				target, inside := path.CleanRelative(aliases[key])
				if !aliasKey.MatchString(key) || !inside {
					return nil, errors.ModularViteAliasInvalid.Args(name, key, aliases[key])
				}
				if owner, taken := owners[key]; taken && owner != name {
					return nil, errors.ModularViteAliasDuplicate.Args(key, name, owner)
				}

				owners[key] = name
				input.Aliases[key] = stdpath.Join(dir, target)
			}
		}

		inputs = append(inputs, input)
	}

	return inputs, nil
}

// checkPublished fails for a third-party module whose files are not in dir as
// modular:publish writes them: Vite would build without them, or with old ones.
func checkPublished(module contractsmodular.Module, info *debug.BuildInfo, dir string) error {
	dependency, ok := modular.DependencyOf(module, info)
	if !ok {
		return nil
	}

	source := dependency.Path
	if dependency.Version != "" {
		source += " " + dependency.Version
	}

	switch modular.PublishState(module, dependency, path.Base(filepath.FromSlash(dir))) {
	case modular.PublishMissing:
		return errors.ModularModuleNotPublished.Args(module.Name(), source, dir)
	case modular.PublishStale:
		return errors.ModularModulePublishStale.Args(module.Name(), dir, source)
	case modular.PublishConflict:
		return errors.ModularPublishTargetNotOwned.Args(module.Name(), dir, dependency.Path)
	}

	return nil
}

// configuredEntries returns vite.inputs or vite.assets as []string; empty when the
// key is unset or no config is bound, so the JSON is always [] and never null. The
// slice is cloned so sorting and compacting never rewrites the config's own value.
func (r *InputsCommand) configuredEntries(key string) []string {
	if r.config == nil {
		return []string{}
	}

	values := slices.Clone(r.config.GetStringSlice(key, []string{}))
	if values == nil {
		return []string{}
	}

	return values
}

// collect adds the calls' arguments to values, each once and sorted, and the dynamic
// calls to dynamic. A dynamic call's literal arguments are still inputs.
func collect(calls []view.Call, function string, values *[]string, dynamic *[]DynamicCall) {
	for _, call := range calls {
		if call.Dynamic {
			*dynamic = append(*dynamic, DynamicCall{Template: call.Template, Function: function})
		}
		*values = append(*values, call.Args...)
	}

	slices.Sort(*values)
	*values = slices.Compact(*values)
}
