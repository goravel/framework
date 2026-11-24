package modular

import (
	"io/fs"
	stdpath "path"
	"strings"

	contractsmodule "github.com/goravel/framework/contracts/module"
)

// StaticFS resolves a Static() value with Files. Directories are not listable, so
// drivers answer 404.
func StaticFS(name, root string, files contractsmodule.Files) (fs.FS, bool, error) {
	fsys, ok, err := Files(name, "Static()", root, "", files)
	if !ok {
		return nil, false, err
	}

	return noListingFS{fsys}, true, nil
}

// CleanPrefix normalises a URL prefix to "/a/b"; "/" is refused.
func CleanPrefix(prefix string) (string, bool) {
	cleaned := stdpath.Clean("/" + strings.TrimSpace(prefix))
	if cleaned == "/" {
		return "", false
	}

	return cleaned, true
}

// noListingFS refuses to open directories.
type noListingFS struct {
	fs.FS
}

func (r noListingFS) Open(name string) (fs.File, error) {
	file, err := r.FS.Open(name)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, err
	}
	if info.IsDir() {
		_ = file.Close()

		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return file, nil
}
