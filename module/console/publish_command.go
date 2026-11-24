package console

import (
	"path/filepath"
	"runtime/debug"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	frameworkmodule "github.com/goravel/framework/module"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
)

type PublishCommand struct {
	registry contractsmodule.Registry
	// buildInfo names the Go module and version of third-party modules.
	buildInfo func() (*debug.BuildInfo, bool)
}

func NewPublishCommand(registry contractsmodule.Registry) *PublishCommand {
	return &PublishCommand{registry: registry, buildInfo: debug.ReadBuildInfo}
}

// Signature The name and signature of the console command.
func (r *PublishCommand) Signature() string {
	return "module:publish"
}

// Description The console command description.
func (r *PublishCommand) Description() string {
	return "Copy the sources of third-party modules into the project"
}

// Extend The console command extend.
func (r *PublishCommand) Extend() command.Extend {
	return command.Extend{
		Category: "module",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "module",
				Usage: "Publish only the given module",
			},
		},
	}
}

// Handle Execute the console command. It publishes every module and fails when one could not be.
func (r *PublishCommand) Handle(ctx console.Context) error {
	if r.registry == nil {
		return errors.ModuleRegistryNotSet
	}

	modules := r.registry.All()
	if name := ctx.Option("module"); name != "" {
		// As given, then in kebab case, like the --module option of make:*.
		module, ok := r.registry.Find(name)
		if !ok {
			module, ok = r.registry.Find(str.Of(name).Kebab().String())
		}
		if !ok {
			return errors.ModuleNotFound.Args(name)
		}
		modules = []contractsmodule.Module{module}
	}

	info, _ := r.buildInfo()

	failed := 0
	for _, module := range modules {
		status, err := r.publish(module, info)
		if err != nil {
			status = statusFailed
			ctx.Error(err.Error())
			failed++
		}
		ctx.TwoColumnDetail(module.Name(), status)
	}
	if failed > 0 {
		return errors.ModulePublishFailed.Args(failed, len(modules))
	}

	return nil
}

func (r *PublishCommand) publish(module contractsmodule.Module, info *debug.BuildInfo) (string, error) {
	dependency, ok := frameworkmodule.DependencyOf(module, info)
	if !ok {
		return statusSkippedLocal, nil
	}

	dir, _ := r.registry.Path(module.Name())
	changed, err := frameworkmodule.Publish(module, dependency, path.Base(filepath.FromSlash(dir)))
	switch {
	case errors.Is(err, errors.ModulePublishTargetNotOwned):
		// Name the directory as the other messages do, relative to the project.
		return "", errors.ModulePublishTargetNotOwned.Args(module.Name(), dir, dependency.Path)
	case err != nil:
		return "", err
	case changed:
		return statusPublished, nil
	case frameworkmodule.PublishState(module, dependency, path.Base(filepath.FromSlash(dir))) == frameworkmodule.PublishNone:
		return statusSkippedNone, nil
	default:
		return statusUnchanged, nil
	}
}
