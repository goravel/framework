package view

type View interface {
	// Exists checks if a view with the specified name exists.
	Exists(view string) bool
	// LoadViewsFrom registers a package view directory for template fallback.
	// Templates from registered directories are loaded after app views; if a
	// template name is already defined by an app view or an earlier package, it is skipped.
	LoadViewsFrom(path string)
	// RegisteredViews returns the absolute paths of all registered package view directories.
	RegisteredViews() []string
	// Share associates a key-value pair, where the key is a string and the value is of any type,
	// with the current view context. This shared data can be accessed by other parts of the application.
	Share(key string, value any)
	// Shared retrieves the value associated with the given key from the current view context's shared data.
	// If the key does not exist, it returns the optional default value (if provided).
	Shared(key string, def ...any) any
	// GetShared returns a map containing all the shared data associated with the current view context.
	GetShared() map[string]any
}
