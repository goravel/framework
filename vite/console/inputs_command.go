package console

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"

	contractsconfig "github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsview "github.com/goravel/framework/contracts/view"
	contractsvite "github.com/goravel/framework/contracts/vite"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/view"
)

// Template functions whose string arguments name Vite entries and assets, as vite.Funcs registers them.
const (
	tagsFunction  = "vite"
	assetFunction = "viteAsset"
)

// Inputs is what goravel-vite-plugin reads: the entries passed to vite and the
// files passed to viteAsset, each once and sorted, the calls with computed
// arguments, and the configured paths. The entries and assets include vite.inputs
// and vite.assets from config, so a build tool need not reach into config itself.
type Inputs struct {
	Inputs  []string      `json:"inputs"`
	Assets  []string      `json:"assets"`
	Dynamic []DynamicCall `json:"dynamic"`
	Paths   Paths         `json:"paths"`
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
	view   contractsview.View
	vite   contractsvite.Vite
	config contractsconfig.Config
}

func NewInputsCommand(views contractsview.View, vite contractsvite.Vite, config contractsconfig.Config) *InputsCommand {
	return &InputsCommand{view: views, vite: vite, config: config}
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

	tmpl, err := view.Compile(view.CompileOptions{
		FuncMap: r.view.Funcs(),
		Paths:   r.view.RegisteredViews(),
		FS:      r.view.RegisteredViewFS(),
	})
	if err != nil {
		return errors.ViewCompileFailed.Args(err).SetModule(errors.ModuleVite)
	}

	inputs := Inputs{
		Inputs:  r.configured("vite.inputs"),
		Assets:  r.configured("vite.assets"),
		Dynamic: []DynamicCall{},
		Paths:   r.paths(),
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

// configured returns vite.inputs or vite.assets as []string; empty when the key is
// unset or no config is bound, so the JSON is always [] and never null.
func (r *InputsCommand) configured(key string) []string {
	if r.config == nil {
		return []string{}
	}

	values := r.config.GetStringSlice(key, []string{})
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
