package view

import "io/fs"

type View interface {
	// Exists checks if a view with the specified name exists.
	Exists(view string) bool
	// First returns a template for the first view in the list that exists. If none of them
	// exist, rendering the returned template fails.
	First(views []string, data ...any) Template
	// LoadViewsFrom registers a package view directory for template fallback.
	// Templates from registered directories are loaded after app views; if a
	// template name is already defined by an app view or an earlier package, it is skipped.
	LoadViewsFrom(path string)
	// LoadViewsFromFS registers an fs.FS (for example an embed.FS) as a package view source.
	// Templates are resolved relative to root within fsys; pass "." to use the whole filesystem.
	// Root may use either slash style ("views/admin" or `views\admin`) and leading "./" or "/"
	// are ignored. Filesystem sources are searched after app views and after directories
	// registered with LoadViewsFrom, in registration order.
	//
	// Unlike LoadViewsFrom, which accepts a directory that may only exist later, an fs.FS is
	// fixed at build time, so it panics if fsys is nil or root is not an existing directory
	// within fsys (including roots that escape it, such as "../views").
	//
	//	//go:embed views
	//	var views embed.FS
	//
	LoadViewsFromFS(fsys fs.FS, root string)
	// RegisteredViews returns the absolute paths of all registered package view directories.
	RegisteredViews() []string
	// RegisteredViewFS returns all package view filesystems registered with LoadViewsFromFS,
	// in registration order. Each returned filesystem is already rooted at the root passed to
	// LoadViewsFromFS, so template paths are relative to it (for example "layouts/app.tmpl").
	RegisteredViewFS() []fs.FS
	// Make returns a template for the view. The data may be a map or a struct; it is merged
	// over the shared data when the template is rendered.
	Make(view string, data ...any) Template
	// Share associates a key-value pair, where the key is a string and the value is of any type,
	// with the current view context. This shared data can be accessed by other parts of the application.
	Share(key string, value any)
	// Shared retrieves the value associated with the given key from the current view context's shared data.
	// If the key does not exist, it returns the optional default value (if provided).
	Shared(key string, def ...any) any
	// GetShared returns a map containing all the shared data associated with the current view context.
	GetShared() map[string]any
}

type Template interface {
	// Data returns the data passed to the template, without the shared data.
	Data() map[string]any
	// Name returns the name of the view.
	Name() string
	// Render renders the view and returns the resulting HTML. Templates are resolved the same
	// way as for HTTP responses: application views first, then directories registered with
	// LoadViewsFrom, then filesystems registered with LoadViewsFromFS.
	//
	// Rendering happens outside a request, so request-bound values such as csrf_token are
	// not available.
	Render() (string, error)
	// With adds a key-value pair to the template data.
	With(key string, value any) Template
}
