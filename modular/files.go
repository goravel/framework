package modular

import (
	"io/fs"
	"os"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/path"
)

// Default directories of a module.Files with an FS, per capability.
const (
	ViewsDirectory        = "resources/views"
	TranslationsDirectory = "lang"
)

// Files resolves what a capability (e.g. "Views()") returned: files.Dir inside files.FS, or
// on disk under root, the module directory. ok is false for the zero value. An FS without
// a Dir uses defaultDir, or its root when there is none.
func Files(name, capability, root, defaultDir string, files contractsmodule.Files) (fsys fs.FS, ok bool, err error) {
	if files.FS == nil && files.Dir == "" {
		return nil, false, nil
	}

	given := files.Dir
	if files.FS != nil && given == "" {
		if defaultDir == "" {
			return files.FS, true, nil
		}
		given = defaultDir
	}

	dir, inside := path.CleanRelative(given)
	if !inside {
		return nil, false, errors.ModularFilesDirInvalid.Args(name, capability, given)
	}

	if files.FS != nil {
		fsys, err = fs.Sub(files.FS, dir)
	} else {
		base, rooted := path.CleanRelative(root)
		if !rooted {
			return nil, false, errors.ModularFilesDirInvalid.Args(name, capability, given)
		}
		fsys = os.DirFS(path.Base(base, dir))
	}

	if err == nil {
		var info fs.FileInfo
		if info, err = fs.Stat(fsys, "."); err == nil && !info.IsDir() {
			err = fs.ErrInvalid
		}
	}
	if err != nil {
		return nil, false, errors.ModularFilesDirUnreadable.Args(name, capability, given, err)
	}

	return fsys, true, nil
}
