package console

import (
	"errors"
	"go/build"
	"os"
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

	for sourcePath, targetValue := range paths {
		targetValue = strings.TrimPrefix(strings.TrimPrefix(targetValue, "/"), "./")
		packagePath := filepath.Join(packageDir, sourcePath)

		res, err := receiver.publish(packagePath, targetValue, ctx.OptionBool("existing"), ctx.OptionBool("force"))
		if err != nil {
			return err
		}

		if len(res) > 0 {
			for sourceFile, targetFile := range res {
				color.Greenp("Copied Directory ")
				color.Yellowf("[%s]", sourceFile)
				color.Greenp(" To ")
				color.Yellowf("%s\n", targetFile)
			}
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

func (receiver *VendorPublishCommand) publish(sourcePath, targetPath string, existing, force bool) (map[string]string, error) {
	result := make(map[string]string)
	sourcePathStat, err := os.Stat(sourcePath)
	if err != nil {
		return nil, err
	}

	var sourceFiles []string
	if sourcePathStat.IsDir() {
		fileInfos, err := os.ReadDir(sourcePath)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range fileInfos {
			sourceFiles = append(sourceFiles, filepath.Join(sourcePath, fileInfo.Name()))
		}
	} else {
		sourceFiles = append(sourceFiles, sourcePath)
	}

	for _, sourceFile := range sourceFiles {
		targetFile := targetPath
		if filepath.Ext(targetFile) == "" {
			targetFile = filepath.Join(targetFile, filepath.Base(sourceFile))
		}

		success, err := receiver.publishFile(sourceFile, targetFile, existing, force)
		if err != nil {
			return nil, err
		}
		if success {
			result[sourceFile] = targetFile
		}
	}

	return result, nil
}

func (receiver *VendorPublishCommand) publishFile(sourceFile, targetFile string, existing, force bool) (bool, error) {
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return false, err
	}

	if !file.Exists(targetFile) && existing {
		return false, nil
	}
	if file.Exists(targetFile) && !force && !existing {
		return false, nil
	}

	if err := file.Create(targetFile, string(content)); err != nil {
		return false, err
	}

	return true, nil
}
