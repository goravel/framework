package modular

import (
	"errors"
	"io/fs"
	"strings"
)

var _ fs.ReadFileFS = (*langFS)(nil)

// langFS maps "en/blog/messages.json" to the blog module's "en/messages.json".
// The app's own lang FS (fallback) is asked first, so it can override modules.
type langFS struct {
	modules  map[string]fs.FS
	fallback fs.FS
}

// NewLangFS combines module translations, keyed by name, with an optional fallback.
// A langFS fallback is unwrapped: its modules are replaced and its own fallback kept.
func NewLangFS(modules map[string]fs.FS, fallback fs.FS) fs.FS {
	if prev, ok := fallback.(*langFS); ok {
		fallback = prev.fallback
	}

	return &langFS{modules: modules, fallback: fallback}
}

func (r *langFS) Open(name string) (fs.File, error) {
	return open(r, "open", name, func(fsys fs.FS, name string) (fs.File, error) {
		return fsys.Open(name)
	})
}

func (r *langFS) ReadFile(name string) ([]byte, error) {
	return open(r, "readfile", name, fs.ReadFile)
}

// open asks the fallback, then the owning module.
func open[T any](r *langFS, op, name string, read func(fs.FS, string) (T, error)) (T, error) {
	var zero T
	if !fs.ValidPath(name) {
		return zero, &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}

	module, inner, owned := r.owner(name)
	if r.fallback != nil {
		result, err := read(r.fallback, name)
		if err == nil || !owned || !errors.Is(err, fs.ErrNotExist) {
			return result, err
		}
	}
	if !owned {
		return zero, &fs.PathError{Op: op, Path: name, Err: fs.ErrNotExist}
	}

	return read(module, inner)
}

// owner splits "<locale>/<module>/<rest>".
func (r *langFS) owner(name string) (fs.FS, string, bool) {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) < 3 {
		return nil, "", false
	}

	module, ok := r.modules[parts[1]]
	if !ok || module == nil {
		return nil, "", false
	}

	return module, parts[0] + "/" + parts[2], true
}
