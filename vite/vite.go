// Package vite resolves assets built by Vite with goravel-vite-plugin into the
// tags that load them, and registers the "vite", "viteAsset" and
// "viteReactRefresh" template functions so views can write
//
//	{{ vite "resources/css/app.css" "resources/js/app.js" }}
package vite

import (
	"crypto/md5"
	"encoding/hex"
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
	defaultBuildSubdirectory = "build"
	defaultHotFilename       = "hot"
	// DefaultManifestFilename is where Vite writes the manifest, relative to the build directory.
	DefaultManifestFilename = ".vite/manifest.json"
)

// DefaultBuildDirectory returns the default build directory, inside the public
// directory.
func DefaultBuildDirectory() string {
	return defaultBuildSubdirectory
}

// DefaultHotFile returns the default hot file, inside the public directory.
func DefaultHotFile() string {
	return defaultHotFilename
}

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
	chunks map[string]chunk
	// importsOf is each chunk's transitive imports in preload order, computed once
	// when the manifest is parsed so a render does not re-walk the graph.
	importsOf    map[string][]string
	manifestHash string // md5 of the manifest file, so ManifestHash and the tags come from the same bytes
	size         int64
	modTime      int64
}

// newManifest indexes the parsed chunks.
func newManifest(chunks map[string]chunk, size, modTime int64, manifestHash string) *manifest {
	m := &manifest{chunks: chunks, importsOf: make(map[string][]string, len(chunks)), manifestHash: manifestHash, size: size, modTime: modTime}
	for name := range chunks {
		m.importsOf[name] = m.imports(name, nil)
	}

	return m
}

type Vite struct {
	config config.Config

	mu               sync.RWMutex
	hotFile          string
	hotPath          string // hotFile resolved to an absolute path, once, so a render does not call getwd
	buildDirectory   string
	buildRoot        string // buildDirectory resolved to an absolute path, once
	manifestFilename string

	manifestMu sync.Mutex
	manifests  map[string]*manifest
}

// NewVite reads vite.hot_file, vite.build_directory and vite.manifest from the
// configuration, falling back to the defaults. goravel-vite-plugin reads the
// same values through `artisan vite:inputs`, so config/vite.go is the single source of truth.
func NewVite(config config.Config) *Vite {
	defaultHot, defaultBuild := DefaultHotFile(), DefaultBuildDirectory()
	v := &Vite{
		config:           config,
		hotFile:          defaultHot,
		buildDirectory:   defaultBuild,
		manifestFilename: DefaultManifestFilename,
		manifests:        make(map[string]*manifest),
	}

	if config != nil {
		v.hotFile = config.GetString("vite.hot_file", defaultHot)
		v.buildDirectory = config.GetString("vite.build_directory", defaultBuild)
		v.manifestFilename = config.GetString("vite.manifest", DefaultManifestFilename)
	}

	v.hotPath = resolveHotPath(v.hotFile)
	v.buildDirectory = normalizeEntry(v.buildDirectory)
	v.buildRoot = resolveBuildRoot(v.buildDirectory)

	return v
}

// resolveHotPath is the absolute path of the hot file: an absolute value is used
// as is, a value that names the public directory relative to the project root
// ("public/hot") resolves against the project root, and anything else resolves
// against the public directory.
func resolveHotPath(hotFile string) string {
	if filepath.IsAbs(hotFile) {
		return paths.Abs(hotFile)
	}

	hotFile = normalizeEntry(hotFile)
	if namesPublicDirectory(hotFile) {
		return paths.Abs(hotFile)
	}

	return paths.Abs(support.Config.Paths.Public, hotFile)
}

// namesPublicDirectory reports whether hotFile names the public directory
// relative to the project root, the older "public/hot" form. Both sides are
// compared as cleaned slash paths, case-insensitively, and a public directory
// that is not project-relative never matches.
func namesPublicDirectory(hotFile string) bool {
	public := slashClean(support.Config.Paths.Public)
	if public == "." || path.IsAbs(public) || filepath.VolumeName(support.Config.Paths.Public) != "" {
		return false
	}

	value := slashClean(hotFile)
	if len(value) <= len(public) {
		return strings.EqualFold(value, public)
	}

	return strings.EqualFold(value[:len(public)], public) && value[len(public)] == '/'
}

// slashClean rewrites a path to slash form and cleans it, so spellings that
// differ only by separators compare equal on every OS.
func slashClean(value string) string {
	return path.Clean(strings.ReplaceAll(value, `\`, "/"))
}

// resolveBuildRoot is the absolute path of the build directory, resolved against
// the public directory.
func resolveBuildRoot(directory string) string {
	return paths.Abs(support.Config.Paths.Public, normalizeEntry(directory))
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

	return assetURL(r.assetBase(), c.File), nil
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

	// The manifest is data, so a "file" such as "../../.env" must not reach outside the build.
	root := r.buildPath()
	file := filepath.Join(root, filepath.FromSlash(c.File))
	if rel, err := filepath.Rel(root, file); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.ViteAssetOutsideBuild.Args(c.File)
	}

	content, err := os.ReadFile(file)
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

	return r.hotPath
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

	return r.buildTags(m, r.assetBase(), entries)
}

// UseBuildDirectory sets the build directory, given relative to the public
// directory as in config/vite.go. Default "build".
// The setters live on *Vite, not the contract: the shared instance is configured
// from config/vite.go, which goravel-vite-plugin reads too, so a change at render
// time would desync the build. They exist for tests and for applications that
// construct their own Vite.
func (r *Vite) UseBuildDirectory(directory string) *Vite {
	directory = normalizeEntry(directory)
	r.mu.Lock()
	r.buildDirectory = directory
	r.buildRoot = resolveBuildRoot(directory)
	r.mu.Unlock()

	return r
}

// UseHotFile sets the hot file. Default "hot". A relative value is read under the
// public directory, a value that names the public directory relative to the
// project root ("public/hot") is read from the project root for older config
// files, and an absolute value is used as is.
func (r *Vite) UseHotFile(hotFile string) *Vite {
	r.mu.Lock()
	r.hotFile = hotFile
	r.hotPath = resolveHotPath(hotFile)
	r.mu.Unlock()

	return r
}

// UseManifestFilename sets the manifest file name inside the build directory. Default ".vite/manifest.json".
func (r *Vite) UseManifestFilename(filename string) *Vite {
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
func (r *Vite) buildTags(m *manifest, base string, entries []string) (template.HTML, error) {
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
				stylesheets.WriteString(stylesheetTag(assetURL(base, c.File)))
			}
			continue
		}

		for _, css := range c.CSS {
			if once(css) {
				stylesheets.WriteString(stylesheetTag(assetURL(base, css)))
			}
		}
		for _, imported := range m.importsOf[entry] {
			ic := m.chunks[imported]
			if once(ic.File) {
				preloads.WriteString(preloadTag(assetURL(base, ic.File)))
			}
			for _, css := range ic.CSS {
				if once(css) {
					stylesheets.WriteString(stylesheetTag(assetURL(base, css)))
				}
			}
		}
		if once(c.File) {
			scripts.WriteString(scriptTag(assetURL(base, c.File)))
		}
	}

	return template.HTML(preloads.String() + stylesheets.String() + scripts.String()), nil
}

// imports returns every chunk the named chunk statically imports, transitively,
// each once in discovery order, skipping cycles and chunks missing from the
// manifest. It runs once per chunk in newManifest, never per render.
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

// assetBase is the public URL of the build directory, prefixed with app.asset_url.
// It is read once per Tags or Asset call, not once per emitted tag.
func (r *Vite) assetBase() string {
	r.mu.RLock()
	directory := r.buildDirectory
	r.mu.RUnlock()

	prefix := ""
	if r.config != nil {
		prefix = strings.TrimRight(r.config.GetString("app.asset_url"), "/")
	}

	if directory = strings.Trim(directory, "/"); directory == "" {
		return prefix
	}

	return prefix + "/" + directory
}

// assetURL is the public URL of a file inside the build directory. The base is
// concatenated, not joined, so the "//" of an absolute asset URL survives.
func assetURL(base, file string) string {
	return base + "/" + strings.TrimLeft(file, "/")
}

// buildPath is the absolute path of the build directory.
func (r *Vite) buildPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.buildRoot
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

	sum := md5.Sum(content)
	m := newManifest(chunks, info.Size(), info.ModTime().UnixNano(), hex.EncodeToString(sum[:]))
	r.manifests[manifestPath] = m

	return m, nil
}

// ManifestHash returns an asset version derived from the build manifest — the md5
// of the manifest file — or "" while the dev server runs or there is no readable
// manifest. It is read through the same cache as Tags, so a rebuild is picked up
// by Tags and ManifestHash together.
func (r *Vite) ManifestHash() string {
	if r.IsRunningHot() {
		return ""
	}

	m, err := r.manifest()
	if err != nil {
		return ""
	}

	return m.manifestHash
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
