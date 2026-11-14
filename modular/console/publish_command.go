package console

import (
	"path/filepath"
	"runtime/debug"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
)

type PublishCommand struct {
	registry contractsmodular.Registry
	// buildInfo names the Go module and version of third-party modules.
	buildInfo func() (*debug.BuildInfo, bool)
}

func NewPublishCommand(registry contractsmodular.Registry) *PublishCommand {
	return &PublishCommand{registry: registry, buildInfo: debug.ReadBuildInfo}
}

// Signature The name and signature of the console command.
func (r *PublishCommand) Signature() string {
	return "modular:publish"
}

// Description The console command description.
func (r *PublishCommand) Description() string {
	return "Copy the sources of third-party modules into the project"
}

// Extend The console command extend.
func (r *PublishCommand) Extend() command.Extend {
	return command.Extend{
		Category: "modular",
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
		return errors.ModularRegistryNotSet
	}

	modules := r.registry.All()
	if name := ctx.Option("module"); name != "" {
		// As given, then in kebab case, like the --module option of make:*.
		module, ok := r.registry.Find(name)
		if !ok {
			module, ok = r.registry.Find(str.Of(name).Kebab().String())
		}
		if !ok {
			return errors.ModularModuleNotFound.Args(name)
		}
		modules = []contractsmodular.Module{module}
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
		return errors.ModularPublishFailed.Args(failed, len(modules))
	}

	return nil
}

func (r *PublishCommand) publish(module contractsmodular.Module, info *debug.BuildInfo) (string, error) {
	dependency, ok := modular.DependencyOf(module, info)
	if !ok {
		return statusSkippedLocal, nil
	}

	dir, _ := r.registry.Path(module.Name())
	changed, err := modular.Publish(module, dependency, path.Base(filepath.FromSlash(dir)))
	switch {
	case errors.Is(err, errors.ModularPublishTargetNotOwned):
		// Name the directory as the other messages do, relative to the project.
		return "", errors.ModularPublishTargetNotOwned.Args(module.Name(), dir, dependency.Path)
	case err != nil:
		return "", err
	case changed:
		return statusPublished, nil
	case modular.PublishState(module, dependency, path.Base(filepath.FromSlash(dir))) == modular.PublishNone:
		return statusSkippedNone, nil
	default:
		return statusUnchanged, nil
	}
}
