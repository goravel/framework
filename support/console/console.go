package console

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type Make struct {
	name string
	root string
}

func NewMake(ctx console.Context, ttype, name, root string) (*Make, error) {
	if name == "" {
		var err error
		name, err = ctx.Ask(fmt.Sprintf("Enter the %s name", ttype), console.AskOption{
			Validate: func(s string) error {
				if s == "" {
					return errors.ConsoleEmptyFieldValue.Args(ttype)
				}

				return nil
			},
		})
		if err != nil {
			return nil, err
		}
	}

	m := &Make{
		name: name,
		root: root,
	}

	if !ctx.OptionBool("force") && file.Exists(m.GetFilePath()) {
		return nil, errors.ConsoleFileAlreadyExists.Args(ttype)
	}

	return m, nil
}

func (m *Make) GetFilePath() string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, m.root, m.GetFolderPath(), str.Of(m.GetStructName()).Snake().String()+".go")
}

func (m *Make) GetSignature() string {
	return str.Of(filepath.Join(m.GetFolderPath(), m.GetStructName())).
		Replace(string(filepath.Separator), "_").Studly().String()
}

func (m *Make) GetStructName() string {
	name := strings.TrimSuffix(m.name, ".go")
	segments := strings.Split(name, "/")

	return str.Of(segments[len(segments)-1]).Studly().String()
}

func (m *Make) GetPackageImportPath() string {
	var paths []string
	if info, ok := debug.ReadBuildInfo(); ok {
		paths = append(paths, info.Main.Path)
	}

	if len(m.root) > 0 {
		paths = append(paths, strings.Split(m.root, string(filepath.Separator))...)
	}

	if folders := m.GetFolderPath(); len(folders) > 0 {
		paths = append(paths, strings.Split(folders, string(filepath.Separator))...)
	}

	return strings.Join(paths, "/")
}

func (m *Make) GetPackageName() string {
	name := strings.TrimSuffix(m.name, ".go")
	segments := strings.Split(name, "/")
	packageName := str.Of(m.root).Trim(string(filepath.Separator)).AfterLast(string(filepath.Separator)).String()

	if len(segments) > 1 {
		packageName = segments[len(segments)-2]
	}

	return packageName
}

func (m *Make) GetFolderPath() string {
	name := strings.TrimSuffix(m.name, ".go")
	segments := strings.Split(name, "/")

	var folderPath string
	if len(segments) > 1 {
		folderPath = filepath.Join(segments[:len(segments)-1]...)
	}

	return folderPath
}

func ConfirmToProceed(ctx console.Context, env string) bool {
	if env != "production" {
		return true
	}
	if ctx.OptionBool("force") {
		return true
	}

	return ctx.Confirm("Are you sure you want to run this command?")
}
