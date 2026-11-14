package modular

import (
	stdpath "path"
	"reflect"
	"strings"

	"github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/path"
)

// DefaultDirectory holds modules, including published third-party ones.
const DefaultDirectory = "modules"

// ResolvePath returns the module directory relative to the project root: Located,
// else the package path relative to mainModule, else "<directory>/<name>" for a
// package of another Go module.
func ResolvePath(module modular.Module, mainModule, directory string) (string, error) {
	name := module.Name()
	if located, ok := module.(modular.Located); ok {
		dir, ok := path.CleanRelative(located.Path())
		if !ok || strings.TrimSpace(located.Path()) == "" {
			return "", errors.ModularModulePathInvalid.Args(name, located.Path())
		}

		return dir, nil
	}

	return derivePath(name, packagePath(module), mainModule, directory)
}

func derivePath(name, pkg, mainModule, directory string) (string, error) {
	if pkg == "" || pkg == "main" || mainModule == "" {
		return "", errors.ModularModulePathUnknown.Args(name, pkg)
	}

	if pkg == mainModule {
		return ".", nil
	}
	if rel, ok := strings.CutPrefix(pkg, mainModule+"/"); ok {
		return rel, nil
	}

	// Another Go module: one element under directory.
	base, ok := path.CleanRelative(directory)
	if !ok {
		return "", errors.ModularModulePathInvalid.Args(name, directory)
	}
	dir := stdpath.Join(base, name)
	if stdpath.Dir(dir) != base || stdpath.Base(dir) != name {
		return "", errors.ModularModulePathInvalid.Args(name, dir)
	}

	return dir, nil
}

// packagePath is the import path of module's type, "" when unnamed.
func packagePath(module modular.Module) string {
	t := reflect.TypeOf(module)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.PkgPath()
}
