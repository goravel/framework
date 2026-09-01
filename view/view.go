package view

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

type View struct {
	mu          sync.RWMutex
	paths       []string
	filesystems []fs.FS
	shared      sync.Map
}

func NewView() *View {
	return &View{}
}

func (r *View) Exists(view string) bool {
	if file.Exists(paths.Abs(support.Config.Paths.Resources, "views", view)) {
		return true
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.paths {
		if file.Exists(paths.Abs(p, view)) {
			return true
		}
	}

	name := normalizeFSPath(view)
	for _, fsys := range r.filesystems {
		if _, err := fs.Stat(fsys, name); err == nil {
			return true
		}
	}

	return false
}

func (r *View) LoadViewsFrom(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
}

func (r *View) LoadViewsFromFS(fsys fs.FS, root string) {
	if fsys == nil {
		panic(errors.ViewFSRequired)
	}

	// Root the filesystem once at registration so consumers can treat every
	// registered source uniformly, with template paths relative to ".".
	sub, err := fs.Sub(fsys, normalizeFSPath(root))
	if err != nil {
		panic(errors.ViewInvalidFSRoot.Args(root, err))
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.filesystems = append(r.filesystems, sub)
}

func (r *View) RegisteredViews() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.paths))
	copy(out, r.paths)
	return out
}

func (r *View) RegisteredViewFS() []fs.FS {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]fs.FS, len(r.filesystems))
	copy(out, r.filesystems)
	return out
}

func (r *View) Share(key string, value any) {
	r.shared.Store(key, value)
}

func (r *View) Shared(key string, def ...any) any {
	value, ok := r.shared.Load(key)
	if !ok {
		if len(def) > 0 {
			return def[0]
		}

		return nil
	}

	return value
}

func (r *View) GetShared() map[string]any {
	shared := make(map[string]any)
	r.shared.Range(func(key, value any) bool {
		shared[key.(string)] = value
		return true
	})

	return shared
}

// normalizeFSPath converts an OS-style or loosely written path ("", "./views",
// "/views/", `views\admin`) into the slash-separated, unrooted form required by io/fs.
func normalizeFSPath(p string) string {
	p = strings.TrimPrefix(path.Clean(filepath.ToSlash(p)), "/")
	if p == "" {
		return "."
	}

	return p
}
