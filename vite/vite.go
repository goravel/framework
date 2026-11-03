// Package vite resolves assets built by Vite with goravel-vite-plugin into the
// tags that load them, and registers the "vite", "viteAsset" and
// "viteReactRefresh" template functions so views can write
//
//	{{ vite "resources/css/app.css" "resources/js/app.js" }}
package vite

import (
	"encoding/json"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goravel/framework/contracts/config"
	contractsvite "github.com/goravel/framework/contracts/vite"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/packages/paths"
	"github.com/goravel/framework/support"
)

const (
	// DefaultBuildDirectory is where goravel-vite-plugin writes the build, inside the public directory.
	DefaultBuildDirectory = "build"
	// DefaultHotFile is where goravel-vite-plugin announces the dev server, relative to the project root.
	DefaultHotFile = "public/hot"
	// DefaultManifestFilename is where Vite writes the manifest, relative to the build directory.
	DefaultManifestFilename = ".vite/manifest.json"
)

var _ contractsvite.Vite = (*Vite)(nil)

// chunk is one entry of Vite's manifest.json.
type chunk struct {
	File           string   `json:"file"`
	Src            string   `json:"src"`
	IsEntry        bool     `json:"isEntry"`
	CSS            []string `json:"css"`
	Imports        []string `json:"imports"`
	DynamicImports []string `json:"dynamicImports"`
}

// manifest is a parsed manifest.json with the size and mtime of its file, so a
// rebuild is picked up without a restart.
type manifest struct {
	chunks  map[string]chunk
	size    int64
	modTime int64
}

type Vite struct {
	config config.Config

	mu               sync.RWMutex
	hotFile          string
	buildDirectory   string
	manifestFilename string

	manifestMu sync.Mutex
	manifests  map[string]*manifest
}

// NewVite reads vite.hot_file, vite.build_directory and vite.manifest from the
// configuration, falling back to the defaults. goravel-vite-plugin reads the same
// values through `artisan vite:inputs`, so config/vite.go is the single source of truth.
func NewVite(config config.Config) *Vite {
	v := &Vite{
		config:           config,
		hotFile:          DefaultHotFile,
		buildDirectory:   DefaultBuildDirectory,
		manifestFilename: DefaultManifestFilename,
		manifests:        make(map[string]*manifest),
	}

	if config != nil {
		v.hotFile = config.GetString("vite.hot_file", DefaultHotFile)
		v.buildDirectory = config.GetString("vite.build_directory", DefaultBuildDirectory)
		v.manifestFilename = config.GetString("vite.manifest", DefaultManifestFilename)
	}

	return v
}

// Funcs returns the template functions the service provider registers, for
// applications that build their own template set.
func Funcs(v contractsvite.Vite) template.FuncMap {
	return template.FuncMap{
		"vite":             v.Tags,
		"viteAsset":        v.Asset,
		"viteReactRefresh": v.ReactRefresh,
	}
}

func (r *Vite) Asset(asset string) (string, error) {
	asset = normalizeEntry(asset)
	if url, ok := r.hotURL(); ok {
		return url + "/" + asset, nil
	}

	m, err := r.manifest()
	if err != nil {
		return "", err
	}

	c, ok := m.chunks[asset]
	if !ok {
		return "", errors.ViteEntryNotFound.Args(asset)
	}

	return r.assetURL(c.File), nil
}

func (r *Vite) Content(asset string) (string, error) {
	asset = normalizeEntry(asset)
	m, err := r.manifest()
	if err != nil {
		return "", err
	}

	c, ok := m.chunks[asset]
	if !ok {
		return "", errors.ViteEntryNotFound.Args(asset)
	}

	content, err := os.ReadFile(filepath.Join(r.buildPath(), filepath.FromSlash(c.File)))
	if err != nil {
		return "", errors.ViteAssetReadFailed.Args(c.File, err)
	}

	return string(content), nil
}

func (r *Vite) BuildDirectory() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.buildDirectory
}

func (r *Vite) HotFile() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return paths.Abs(r.hotFile)
}

func (r *Vite) ManifestFilename() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.manifestFilename
}

func (r *Vite) IsRunningHot() bool {
	info, err := os.Stat(r.HotFile())

	return err == nil && !info.IsDir()
}

func (r *Vite) ReactRefresh() template.HTML {
	url, ok := r.hotURL()
	if !ok {
		return ""
	}

	// The URL is placed inside a JavaScript string, so it is escaped as one.
	return template.HTML(`<script type="module">import RefreshRuntime from '` + template.JSEscapeString(url) + `/@react-refresh';` +
		`RefreshRuntime.injectIntoGlobalHook(window);` +
		`window.$RefreshReg$ = () => {};` +
		`window.$RefreshSig$ = () => (type) => type;` +
		`window.__vite_plugin_react_preamble_installed__ = true;</script>`)
}

func (r *Vite) Tags(entries ...string) (template.HTML, error) {
	normalized := make([]string, len(entries))
	for i, entry := range entries {
		normalized[i] = normalizeEntry(entry)
	}
	entries = normalized

	if url, ok := r.hotURL(); ok {
		return r.hotTags(url, entries), nil
	}

	m, err := r.manifest()
	if err != nil {
		return "", err
	}

	return r.buildTags(m, entries)
}

func (r *Vite) UseBuildDirectory(directory string) contractsvite.Vite {
	r.mu.Lock()
	r.buildDirectory = directory
	r.mu.Unlock()

	return r
}

func (r *Vite) UseHotFile(path string) contractsvite.Vite {
	r.mu.Lock()
	r.hotFile = path
	r.mu.Unlock()

	return r
}

func (r *Vite) UseManifestFilename(filename string) contractsvite.Vite {
	r.mu.Lock()
	r.manifestFilename = filename
	r.mu.Unlock()

	return r
}

// normalizeEntry strips a leading slash so "/resources/js/app.js" and
// "resources/js/app.js" name the same manifest key and dev-server path.
func normalizeEntry(entry string) string {
	return strings.TrimPrefix(entry, "/")
}

// hotTags loads the Vite client, then every entry from the dev server.
func (r *Vite) hotTags(url string, entries []string) template.HTML {
	var b strings.Builder
	b.WriteString(scriptTag(url + "/@vite/client"))
	for _, entry := range entries {
		src := url + "/" + entry
		if isCSS(entry) {
			b.WriteString(stylesheetTag(src))
		} else {
			b.WriteString(scriptTag(src))
		}
	}

	return template.HTML(b.String())
}

// buildTags resolves the entries through the manifest: modulepreloads for imported
// chunks, then stylesheets, then scripts, each file once.
func (r *Vite) buildTags(m *manifest, entries []string) (template.HTML, error) {
	var preloads, stylesheets, scripts strings.Builder
	seen := make(map[string]struct{})
	once := func(file string) bool {
		if _, ok := seen[file]; ok {
			return false
		}
		seen[file] = struct{}{}

		return true
	}

	for _, entry := range entries {
		c, ok := m.chunks[entry]
		if !ok {
			return "", errors.ViteEntryNotFound.Args(entry)
		}

		if isCSS(c.File) {
			if once(c.File) {
				stylesheets.WriteString(stylesheetTag(r.assetURL(c.File)))
			}
			continue
		}

		for _, css := range c.CSS {
			if once(css) {
				stylesheets.WriteString(stylesheetTag(r.assetURL(css)))
			}
		}
		for _, imported := range m.imports(entry, nil) {
			ic := m.chunks[imported]
			if once(ic.File) {
				preloads.WriteString(preloadTag(r.assetURL(ic.File)))
			}
			for _, css := range ic.CSS {
				if once(css) {
					stylesheets.WriteString(stylesheetTag(r.assetURL(css)))
				}
			}
		}
		if once(c.File) {
			scripts.WriteString(scriptTag(r.assetURL(c.File)))
		}
	}

	return template.HTML(preloads.String() + stylesheets.String() + scripts.String()), nil
}

// imports returns every chunk the named chunk statically imports, transitively, each once in discovery order.
func (m *manifest) imports(name string, visited map[string]struct{}) []string {
	if visited == nil {
		visited = map[string]struct{}{name: {}}
	}

	var out []string
	for _, imported := range m.chunks[name].Imports {
		if _, ok := visited[imported]; ok {
			continue
		}
		if _, ok := m.chunks[imported]; !ok {
			continue
		}
		visited[imported] = struct{}{}
		out = append(out, imported)
		out = append(out, m.imports(imported, visited)...)
	}

	return out
}

// hotURL returns the dev server URL while the hot file exists.
func (r *Vite) hotURL() (string, bool) {
	content, err := os.ReadFile(r.HotFile())
	if err != nil {
		return "", false
	}

	return strings.TrimRight(strings.TrimSpace(string(content)), "/"), true
}

// assetURL is the public URL of a file inside the build directory, prefixed with app.asset_url.
func (r *Vite) assetURL(file string) string {
	r.mu.RLock()
	directory := r.buildDirectory
	r.mu.RUnlock()

	prefix := ""
	if r.config != nil {
		prefix = strings.TrimRight(r.config.GetString("app.asset_url"), "/")
	}

	return prefix + "/" + path.Join(strings.Trim(directory, "/"), file)
}

// buildPath is the absolute path of the build directory.
func (r *Vite) buildPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return paths.Abs(support.Config.Paths.Public, r.buildDirectory)
}

// manifest returns the parsed manifest, re-reading it only when the file changed.
func (r *Vite) manifest() (*manifest, error) {
	r.mu.RLock()
	filename := r.manifestFilename
	r.mu.RUnlock()

	buildPath := r.buildPath()
	manifestPath := filepath.Join(buildPath, filepath.FromSlash(filename))
	info, err := os.Stat(manifestPath)
	if err != nil {
		return nil, errors.ViteManifestNotFound.Args(manifestPath)
	}

	r.manifestMu.Lock()
	defer r.manifestMu.Unlock()

	if m, ok := r.manifests[manifestPath]; ok && m.size == info.Size() && m.modTime == info.ModTime().UnixNano() {
		return m, nil
	}

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.ViteManifestReadFailed.Args(manifestPath, err)
	}

	chunks := make(map[string]chunk)
	if err := json.Unmarshal(content, &chunks); err != nil {
		return nil, errors.ViteManifestInvalid.Args(manifestPath, err)
	}

	m := &manifest{chunks: chunks, size: info.Size(), modTime: info.ModTime().UnixNano()}
	r.manifests[manifestPath] = m

	return m, nil
}

func scriptTag(src string) string {
	return `<script type="module" src="` + template.HTMLEscapeString(src) + `"></script>`
}

func stylesheetTag(href string) string {
	return `<link rel="stylesheet" href="` + template.HTMLEscapeString(href) + `">`
}

func preloadTag(href string) string {
	return `<link rel="modulepreload" href="` + template.HTMLEscapeString(href) + `">`
}

func isCSS(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".css", ".less", ".sass", ".scss", ".styl", ".stylus", ".pcss", ".postcss":
		return true
	}

	return false
}
