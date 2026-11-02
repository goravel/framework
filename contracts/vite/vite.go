package vite

import "html/template"

// Vite renders the <script> and <link> tags for assets built by Vite with
// goravel-vite-plugin, like Laravel's @vite directive. While the dev server runs
// (the hot file exists) every tag points at it; otherwise tags are resolved
// through the manifest in the build directory.
type Vite interface {
	// Asset returns the URL of a built asset by its source path, e.g. "resources/images/logo.svg".
	Asset(path string) (string, error)
	// BuildDirectory returns the build directory inside the public directory.
	BuildDirectory() string
	// Content returns a built asset's content for inlining; always read from the build.
	Content(path string) (string, error)
	// HotFile returns the absolute path of the hot file.
	HotFile() string
	// IsRunningHot reports whether the hot file exists.
	IsRunningHot() bool
	// ManifestFilename returns the manifest file name inside the build directory.
	ManifestFilename() string
	// ReactRefresh returns the @vitejs/plugin-react preamble while the dev server runs.
	ReactRefresh() template.HTML
	// Tags returns the stylesheet, modulepreload and module script tags for the
	// entries; while the dev server runs, the Vite client plus one tag per entry.
	Tags(entries ...string) (template.HTML, error)
	// UseBuildDirectory sets the build directory. Default "build".
	UseBuildDirectory(directory string) Vite
	// UseHotFile sets the hot file; a relative path resolves against the project root. Default "public/hot".
	UseHotFile(path string) Vite
	// UseManifestFilename sets the manifest file name. Default ".vite/manifest.json".
	UseManifestFilename(filename string) Vite
}
