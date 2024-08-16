package console

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
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
					return fmt.Errorf("the %s name cannot be empty", ttype)
				}

				return nil
			},
		})
		if err != nil {
			return nil, err
		}
	}

	make := &Make{
		name: name,
		root: root,
	}

	if !ctx.OptionBool("force") && file.Exists(make.GetFilePath()) {
		return nil, fmt.Errorf("the %s already exists. Use the --force or -f flag to overwrite", ttype)
	}

	return make, nil
}

func (m *Make) GetFilePath() string {
	pwd, _ := os.Getwd()

	return filepath.Join(pwd, m.root, m.GetFolderPath(), str.Camel2Case(m.GetStructName())+".go")
}

func (m *Make) GetStructName() string {
	name := strings.TrimSuffix(m.name, ".go")
	segments := strings.Split(name, "/")

	return str.Case2Camel(segments[len(segments)-1])
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
