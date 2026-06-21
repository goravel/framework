package console

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

type LangPublishCommand struct {
	langPath string
	langFS   fs.FS
}

func NewLangPublishCommand(langPath string, langFS fs.FS) *LangPublishCommand {
	return &LangPublishCommand{langPath: langPath, langFS: langFS}
}

func (r *LangPublishCommand) Signature() string {
	return "lang:publish"
}

func (r *LangPublishCommand) Description() string {
	return "Publish the validation language files to the application"
}

func (r *LangPublishCommand) Extend() command.Extend {
	return command.Extend{
		Category: "validation",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite any existing files",
			},
		},
	}
}

func (r *LangPublishCommand) Handle(ctx console.Context) error {
	force := ctx.OptionBool("force")

	err := fs.WalkDir(r.langFS, "lang", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel("lang", path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(r.langPath, relativePath)
		return r.publishFile(path, targetPath, force)
	})
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Success("Publishing complete")

	return nil
}

func (r *LangPublishCommand) publishFile(sourcePath, targetPath string, force bool) error {
	content, err := fs.ReadFile(r.langFS, sourcePath)
	if err != nil {
		return err
	}

	if file.Exists(targetPath) && !force {
		return nil
	}

	if err := file.PutContent(targetPath, string(content)); err != nil {
		return err
	}

	color.Green().Print("Copied ")
	color.Yellow().Printf("[%s]", strings.TrimPrefix(sourcePath, "lang/"))
	color.Green().Print(" To ")
	color.Yellow().Printf("%s\n", targetPath)

	return nil
}
