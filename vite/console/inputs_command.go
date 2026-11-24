package console

import (
	"encoding/json"
	"io/fs"
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
	contractsmodule "github.com/goravel/framework/contracts/module"
	contractsview "github.com/goravel/framework/contracts/view"
	contractsvite "github.com/goravel/framework/contracts/vite"
	"github.com/goravel/framework/errors"
	frameworkmodule "github.com/goravel/framework/module"
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
	modules   contractsmodule.Registry
	buildInfo func() (*debug.BuildInfo, bool)
}

// NewInputsCommand takes the module registry, nil when the application has none.
func NewInputsCommand(views contractsview.View, vite contractsvite.Vite, config contractsconfig.Config, modules contractsmodule.Registry) *InputsCommand {
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
		if p, ok := module.(contractsmodule.ProvidesFrontend); ok {
			frontend := p.Frontend()
			_, thirdParty := frameworkmodule.DependencyOf(module, info)
			root, inSources, err := frontendFiles(module, dir, frontend, thirdParty)
			if err != nil {
				return nil, err
			}

			for _, entry := range frontend.Entries {
				// Checked like aliases: Vite would build a file outside the module otherwise.
				cleaned, inside := path.CleanRelative(entry)
				if !inside || cleaned == "." {
					return nil, errors.ModuleViteEntryInvalid.Args(name, entry)
				}
				if inSources != nil && !inSources(cleaned) {
					return nil, errors.ModuleViteEntryMissing.Args(name, entry)
				}
				input.Entries = append(input.Entries, stdpath.Join(root, cleaned))
			}

			aliases := frontend.Aliases
			for _, key := range slices.Sorted(maps.Keys(aliases)) {
				target, inside := path.CleanRelative(aliases[key])
				if !aliasKey.MatchString(key) || !inside {
					return nil, errors.ModuleViteAliasInvalid.Args(name, key, aliases[key])
				}
				if owner, taken := owners[key]; taken && owner != name {
					return nil, errors.ModuleViteAliasDuplicate.Args(key, name, owner)
				}

				owners[key] = name
				input.Aliases[key] = stdpath.Join(root, target)
			}
		}

		inputs = append(inputs, input)
	}

	return inputs, nil
}

// frontendFiles returns the directory entries and aliases are relative to, and a check
// that an entry is a file of the sources: an embedded Files.FS, or what module:publish
// copies of a third-party module. The check is nil for files of the project.
func frontendFiles(module contractsmodule.Module, dir string, frontend contractsmodule.Frontend, thirdParty bool) (string, func(entry string) bool, error) {
	files := frontend.Files
	sub, inside := path.CleanRelative(files.Dir)
	if !inside {
		return "", nil, errors.ModuleFilesDirInvalid.Args(module.Name(), "Frontend()", files.Dir)
	}
	root := stdpath.Join(dir, sub)
	if files.FS == nil && !thirdParty {
		return root, nil, nil
	}

	sources, prefix, err := frameworkmodule.FrontendSources(module, frontend)
	if err != nil || sources == nil {
		return root, nil, err
	}

	return root, func(entry string) bool {
		// The sources start at prefix, which is below sub for the default directory.
		name := stdpath.Join(sub, entry)
		if prefix != "." {
			var within bool
			if name, within = strings.CutPrefix(name, prefix+"/"); !within {
				return false
			}
		}
		stat, err := fs.Stat(sources, name)

		return err == nil && !stat.IsDir()
	}, nil
}

// checkPublished fails for a third-party module whose files are not in dir as
// module:publish writes them: Vite would build without them, or with old ones.
func checkPublished(module contractsmodule.Module, info *debug.BuildInfo, dir string) error {
	dependency, ok := frameworkmodule.DependencyOf(module, info)
	if !ok {
		return nil
	}

	source := dependency.Path
	if dependency.Version != "" {
		source += " " + dependency.Version
	}

	switch frameworkmodule.PublishState(module, dependency, path.Base(filepath.FromSlash(dir))) {
	case frameworkmodule.PublishMissing:
		return errors.ModuleNotPublished.Args(module.Name(), source, dir)
	case frameworkmodule.PublishStale:
		return errors.ModulePublishStale.Args(module.Name(), dir, source)
	case frameworkmodule.PublishConflict:
		return errors.ModulePublishTargetNotOwned.Args(module.Name(), dir, dependency.Path)
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
