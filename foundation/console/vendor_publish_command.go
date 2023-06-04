package console

import (
	"errors"
	"go/build"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
)

type VendorPublishCommand struct {
	publishes     map[string]map[string]string
	publishGroups map[string]map[string]string
}

func NewVendorPublishCommand(publishes, publishGroups map[string]map[string]string) *VendorPublishCommand {
	return &VendorPublishCommand{
		publishes:     publishes,
		publishGroups: publishGroups,
	}
}

// Signature The name and signature of the console command.
func (receiver *VendorPublishCommand) Signature() string {
	return "vendor:publish"
}

// Description The console command description.
func (receiver *VendorPublishCommand) Description() string {
	return "Publish any publishable assets from vendor packages"
}

// Extend The console command extend.
func (receiver *VendorPublishCommand) Extend() command.Extend {
	return command.Extend{
		Category: "vendor",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "existing",
				Aliases: []string{"e"},
				Usage:   "Publish and overwrite only the files that have already been published",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite any existing files",
			},
			&command.StringFlag{
				Name:    "package",
				Aliases: []string{"p"},
				Usage:   "Package name to publish",
			},
			&command.StringFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "One tag that have assets you want to publish",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *VendorPublishCommand) Handle(ctx console.Context) error {
	packageName := ctx.Option("package")
	paths := receiver.pathsForPackageOrGroup(packageName, ctx.Option("tag"))
	if len(paths) == 0 {
		return errors.New("no vendor found")
	}

	packageDir, err := receiver.packageDir(packageName)
	if err != nil {
		return err
	}

	for key, value := range paths {
		value = strings.TrimPrefix(strings.TrimPrefix(value, "/"), "./")
		content, err := ioutil.ReadFile(filepath.Join(packageDir, key))
		if err != nil {
			return err
		}

		success, err := receiver.publish(value, string(content), ctx.OptionBool("existing"), ctx.OptionBool("force"))
		if err != nil {
			return err
		}

		if success {
			color.Greenp("Copied Directory ")
			color.Yellowf("[%s/%s]", strings.TrimSuffix(packageName, "/"), strings.TrimPrefix(key, "/"))
			color.Greenp(" To ")
			color.Yellowf("/%s\n", value)
		}
	}

	color.Greenln("Publishing complete")

	return nil
}

func (receiver *VendorPublishCommand) pathsForPackageOrGroup(packageName, group string) map[string]string {
	if packageName != "" && group != "" {
		return receiver.pathsForProviderAndGroup(packageName, group)
	} else if group != "" {
		if paths, exist := receiver.publishGroups[group]; exist {
			return paths
		}
	} else if packageName != "" {
		if paths, exist := receiver.publishes[packageName]; exist {
			return paths
		}
	}

	return nil
}

func (receiver *VendorPublishCommand) pathsForProviderAndGroup(packageName, group string) map[string]string {
	packagePaths, exist := receiver.publishes[packageName]
	if !exist {
		return nil
	}

	groupPaths, exist := receiver.publishGroups[group]
	if !exist {
		return nil
	}

	paths := make(map[string]string)
	for key, path := range packagePaths {
		if _, exist := groupPaths[key]; exist {
			paths[key] = path
		}
	}

	return paths
}

func (receiver *VendorPublishCommand) packageDir(packageName string) (string, error) {
	var srcDir string
	if build.IsLocalImport(packageName) {
		srcDir = "./"
	}

	pkg, err := build.Import(packageName, srcDir, build.FindOnly)
	if err != nil {
		return "", err
	}

	return pkg.Dir, nil
}

func (receiver *VendorPublishCommand) publish(path, content string, existing, force bool) (bool, error) {
	if !file.Exists(path) && existing {
		return false, nil
	}
	if file.Exists(path) && !force && !existing {
		return false, nil
	}

	if err := file.Create(path, content); err != nil {
		return false, err
	}

	return true, nil
}
