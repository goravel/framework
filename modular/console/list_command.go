package console

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/path"
)

type ListCommand struct {
	registry  contractsmodular.Registry
	buildInfo func() (*debug.BuildInfo, bool)
	// out receives --json as is: ctx.Line would wrap it in colour codes.
	out io.Writer
}

func NewListCommand(registry contractsmodular.Registry) *ListCommand {
	return &ListCommand{registry: registry, buildInfo: debug.ReadBuildInfo, out: os.Stdout}
}

// moduleInfo is one module in the --json output.
type moduleInfo struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Order        int               `json:"order"`
	Publish      string            `json:"publish"`
	Capabilities map[string]bool   `json:"capabilities"`
	ViteEntries  []string          `json:"vite_entries"`
	ViteAliases  map[string]string `json:"vite_aliases"`
	Static       []string          `json:"static"`
}

// Signature The name and signature of the console command.
func (r *ListCommand) Signature() string {
	return "modular:list"
}

// Description The console command description.
func (r *ListCommand) Description() string {
	return "List the registered modules"
}

// Extend The console command extend.
func (r *ListCommand) Extend() command.Extend {
	return command.Extend{
		Category: "modular",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:  "json",
				Usage: "Output the modules as JSON",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ListCommand) Handle(ctx console.Context) error {
	if r.registry == nil {
		return errors.ModularRegistryNotSet
	}

	info, _ := r.buildInfo()
	modules := make([]moduleInfo, 0)
	for _, module := range r.registry.All() {
		modules = append(modules, r.describe(module, info))
	}

	if ctx.OptionBool("json") {
		content, err := json.Marshal(modules)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(r.out, string(content))

		return err
	}

	if len(modules) == 0 {
		ctx.Warning("No modules are registered. List them in bootstrap/modules.go.")

		return nil
	}

	for _, module := range modules {
		ctx.TwoColumnDetail(
			fmt.Sprintf("<fg=green;op=bold>%s</> %s", module.Name, module.Path),
			fmt.Sprintf("%s, order %d", module.Publish, module.Order),
		)
		for _, detail := range module.details() {
			ctx.Line("    <fg=gray>" + detail + "</>")
		}
	}

	return nil
}

func (r *ListCommand) describe(module contractsmodular.Module, info *debug.BuildInfo) moduleInfo {
	dir, _ := r.registry.Path(module.Name())
	described := moduleInfo{
		Name:         module.Name(),
		Path:         dir,
		Publish:      modular.PublishLocal,
		Capabilities: r.registry.Capabilities(module),
		ViteEntries:  []string{},
		ViteAliases:  map[string]string{},
		Static:       []string{},
	}

	if dependency, ok := modular.DependencyOf(module, info); ok {
		described.Publish = modular.PublishState(module, dependency, path.Base(filepath.FromSlash(dir)))
	}
	if p, ok := module.(contractsmodular.Ordered); ok {
		described.Order = p.Order()
	}
	if p, ok := module.(contractsmodular.ProvidesViteEntries); ok {
		described.ViteEntries = append(described.ViteEntries, p.ViteEntries()...)
	}
	if p, ok := module.(contractsmodular.ProvidesViteAlias); ok {
		maps.Copy(described.ViteAliases, p.ViteAliases())
	}
	if p, ok := module.(contractsmodular.ProvidesStatic); ok {
		described.Static = slices.Sorted(maps.Keys(p.Static()))
	}

	return described
}

// details are the lines under a module row: its capabilities and what they declare.
func (r moduleInfo) details() []string {
	var capabilities []string
	for _, label := range slices.Sorted(maps.Keys(r.Capabilities)) {
		if r.Capabilities[label] {
			capabilities = append(capabilities, label)
		}
	}

	var details []string
	if len(capabilities) > 0 {
		details = append(details, strings.Join(capabilities, ", "))
	}
	if len(r.ViteEntries) > 0 {
		details = append(details, "vite entries: "+strings.Join(r.ViteEntries, ", "))
	}
	if len(r.ViteAliases) > 0 {
		var aliases []string
		for _, alias := range slices.Sorted(maps.Keys(r.ViteAliases)) {
			aliases = append(aliases, alias+" -> "+r.ViteAliases[alias])
		}
		details = append(details, "vite aliases: "+strings.Join(aliases, ", "))
	}
	if len(r.Static) > 0 {
		details = append(details, "static: "+strings.Join(r.Static, ", "))
	}

	return details
}
