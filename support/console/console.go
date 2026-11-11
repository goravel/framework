package console

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/pterm/pterm"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type Make struct {
	name string
	root string
	// moduleDir is the directory of the --module target, "" for the application.
	moduleDir string
}

// ModuleFlag is the --module option of the make:* commands.
func ModuleFlag() command.Flag {
	return &command.StringFlag{
		Name:  "module",
		Usage: "The module to create the file in",
	}
}

// ModuleDir returns the directory of the module given with --module, relative to
// the project root, or "" without the option. Third-party modules are refused.
func ModuleDir(ctx console.Context) (string, error) {
	name := ctx.Option("module")
	if name == "" {
		return "", nil
	}

	registry := modular.Default()
	if registry == nil {
		return "", errors.ModularRegistryNotSet
	}

	module, ok := registry.Find(name)
	if !ok {
		module, ok = registry.Find(str.Of(name).Kebab().String())
	}
	if !ok {
		return "", errors.ModularModuleNotFound.Args(name)
	}
	if !modular.Local(module) {
		return "", errors.ModularModuleNotLocal.Args(module.Name())
	}

	dir, _ := registry.Path(module.Name())

	return dir, nil
}

// ModuleRoot moves root into the module directory, dropping the application
// directory: "app/http/controllers" becomes "<dir>/http/controllers".
func ModuleRoot(dir, root string) string {
	root = strings.Trim(filepath.ToSlash(root), "/")
	app := strings.Trim(filepath.ToSlash(support.Config.Paths.App), "/")
	if root == app {
		root = ""
	} else if app != "" {
		root = strings.TrimPrefix(root, app+"/")
	}

	return path.Join(dir, root)
}

// MakeNames returns every name passed to a make:* command. When no positional
// argument is provided it returns a single empty name so that NewMake falls back
// to its interactive prompt, preserving the original single-name behavior.
func MakeNames(ctx console.Context) []string {
	names := ctx.Arguments()
	if len(names) == 0 {
		return []string{""}
	}

	return names
}

// HandleEach runs makeOne for every name passed to a make:* command. A failure
// for one name is reported and does not stop the remaining names.
func HandleEach(ctx console.Context, makeOne func(ctx console.Context, name string) error) {
	for _, name := range MakeNames(ctx) {
		if err := makeOne(ctx, name); err != nil {
			ctx.Error(err.Error())
		}
	}
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

	moduleDir, err := ModuleDir(ctx)
	if err != nil {
		return nil, err
	}
	if moduleDir != "" {
		root = ModuleRoot(moduleDir, root)
	}

	m := &Make{
		name:      name,
		root:      root,
		moduleDir: moduleDir,
	}

	if !ctx.OptionBool("force") && file.Exists(m.GetFilePath()) {
		return nil, errors.ConsoleFileAlreadyExists.Args(ttype)
	}

	return m, nil
}

func (m *Make) GetName() string {
	return m.name
}

// GetModuleDir returns the directory of the --module target, "" for the application.
func (m *Make) GetModuleDir() string {
	return m.moduleDir
}

func (m *Make) GetFilePath() string {
	root := strings.Split(m.root, "/")
	filePaths := append(root, m.GetFolderPath(), str.Of(m.GetStructName()).Snake().String()+".go")

	return paths.Abs(filePaths...)
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
	// Normalize path separators to forward slashes for cross-platform compatibility
	// Replace backslashes with forward slashes to handle Windows paths
	normalizedRoot := strings.ReplaceAll(m.root, "\\", "/")
	packageName := str.Of(normalizedRoot).Trim("/").AfterLast("/").String()

	if len(segments) > 1 {
		packageName = segments[len(segments)-2]
	}

	return packageName
}

func (m *Make) GetModuleName() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Path
	}

	// fallback to default
	return "goravel"
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

func TwoColumnDetail(first, second string, filler ...rune) string {
	margin := func(s string, left, right int) string {
		var builder strings.Builder
		if left > 0 {
			builder.WriteString(strings.Repeat(" ", left))
		}
		builder.WriteString(s)
		if right > 0 {
			builder.WriteString(strings.Repeat(" ", right))
		}
		return builder.String()
	}
	width := func(s string) int {
		return runewidth.StringWidth(color.StripColor(s))
	}
	first = margin(first, 2, 1)
	if w := width(second); w > 0 {
		second = margin(second, 1, 2)
	} else {
		second = margin(second, 0, 2)
	}
	fillingText := ""
	if w := pterm.GetTerminalWidth() - width(first) - width(second); w > 0 {
		fillingText = color.Gray().Sprint(strings.Repeat(string(append(filler, '.')[0]), w))
	}

	return first + fillingText + second
}
