package module

import (
	"io/fs"

	contractsmodule "github.com/goravel/framework/contracts/module"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/path"
)

// Default directories of a module.Files, per capability.
const (
	ViewsDirectory        = "resources/views"
	TranslationsDirectory = "lang"
	// FrontendDirectory is what module:publish copies from a package directory on disk.
	FrontendDirectory = "resources"
)

// Files resolves what a capability (e.g. "Views()") returned: files.Dir inside files.FS.
// ok is false for the zero value, a Dir without an FS is an error. Without a Dir the
// files start at defaultDir, or at the root of the FS when there is none.
func Files(name, capability, defaultDir string, files contractsmodule.Files) (fsys fs.FS, ok bool, err error) {
	if files.FS == nil {
		if files.Dir == "" {
			return nil, false, nil
		}

		return nil, false, errors.ModuleFilesFSRequired.Args(name, capability, files.Dir)
	}

	given := files.Dir
	if given == "" {
		if defaultDir == "" {
			return files.FS, true, nil
		}
		given = defaultDir
	}

	dir, inside := path.CleanRelative(given)
	if !inside {
		return nil, false, errors.ModuleFilesDirInvalid.Args(name, capability, given)
	}

	fsys, err = fs.Sub(files.FS, dir)
	if err == nil {
		var info fs.FileInfo
		if info, err = fs.Stat(fsys, "."); err == nil && !info.IsDir() {
			err = fs.ErrInvalid
		}
	}
	if err != nil {
		return nil, false, errors.ModuleFilesDirUnreadable.Args(name, capability, given, err)
	}

	return fsys, true, nil
}
