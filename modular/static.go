package modular

import (
	"io/fs"
	"os"
	stdpath "path"
	"strings"

	"github.com/goravel/framework/contracts/modular"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/path"
)

// StaticFS returns p.Dir inside p.FS, or on disk under root (the module directory)
// when p.FS is nil. Directories are not listable, so drivers answer 404.
func StaticFS(name, root string, p modular.Path) (fs.FS, error) {
	dir, ok := path.CleanRelative(p.Dir)
	if !ok {
		return nil, errors.ModularStaticDirInvalid.Args(name, p.Dir)
	}

	var fsys fs.FS
	if p.FS != nil {
		sub, err := fs.Sub(p.FS, dir)
		if err != nil {
			return nil, errors.ModularStaticDirUnreadable.Args(name, p.Dir, err)
		}
		fsys = sub
	} else {
		base, ok := path.CleanRelative(root)
		if !ok {
			return nil, errors.ModularStaticDirInvalid.Args(name, root)
		}
		fsys = os.DirFS(path.Base(base, dir))
	}

	info, err := fs.Stat(fsys, ".")
	if err != nil {
		return nil, errors.ModularStaticDirUnreadable.Args(name, p.Dir, err)
	}
	if !info.IsDir() {
		return nil, errors.ModularStaticDirUnreadable.Args(name, p.Dir, fs.ErrInvalid)
	}

	return noListingFS{fsys}, nil
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
