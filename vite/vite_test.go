package vite

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support"
)

const sampleManifest = `{
  "resources/js/app.js": {
    "file": "assets/app-4ed993c7.js",
    "src": "resources/js/app.js",
    "isEntry": true,
    "css": ["assets/app-9a1b2c3d.css"],
    "imports": ["_vendor-1f2e3d4c.js"]
  },
  "resources/js/admin.js": {
    "file": "assets/admin-5fe6a7b8.js",
    "src": "resources/js/admin.js",
    "isEntry": true,
    "imports": ["_vendor-1f2e3d4c.js", "_shared-9c8b7a6f.js"]
  },
  "_vendor-1f2e3d4c.js": {
    "file": "assets/vendor-1f2e3d4c.js",
    "css": ["assets/vendor-aabbccdd.css"],
    "imports": ["_shared-9c8b7a6f.js"]
  },
  "_shared-9c8b7a6f.js": {
    "file": "assets/shared-9c8b7a6f.js",
    "imports": ["_vendor-1f2e3d4c.js"]
  },
  "resources/css/site.css": {
    "file": "assets/site-0a0b0c0d.css",
    "src": "resources/css/site.css",
    "isEntry": true
  },
  "resources/images/logo.svg": {
    "file": "assets/logo-11223344.svg",
    "src": "resources/images/logo.svg"
  }
}`

// setupProject points the project root at a temporary directory and returns its public directory.
func setupProject(t *testing.T) string {
	t.Helper()

	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	public := filepath.Join(support.RelativePath, support.Config.Paths.Public)
	require.NoError(t, os.MkdirAll(public, 0o755))

	return public
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// newTestVite returns a Vite configured from a config mock that answers with the defaults.
func newTestVite(t *testing.T, assetURL string) *Vite {
	t.Helper()

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return(DefaultHotFile()).Once()
	config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return(DefaultBuildDirectory()).Once()
	config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return(DefaultManifestFilename).Once()
	config.EXPECT().GetString("app.asset_url").Return(assetURL).Maybe()

	return NewVite(config)
}

func TestNewVite_ReadsConfig(t *testing.T) {
	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return("storage/vite.hot").Once()
	config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return("dist").Once()
	config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return("my-manifest.json").Once()

	setupProject(t)
	vite := NewVite(config)
	assert.Equal(t, "storage/vite.hot", vite.hotFile)
	assert.Equal(t, "dist", vite.BuildDirectory())
	assert.Equal(t, "my-manifest.json", vite.ManifestFilename())

	vite = NewVite(nil)
	assert.Equal(t, DefaultHotFile(), vite.hotFile)
	assert.Equal(t, DefaultBuildDirectory(), vite.buildDirectory)
	assert.Equal(t, "build", vite.BuildDirectory()) // public-relative of public/build
	assert.Equal(t, DefaultManifestFilename, vite.manifestFilename)
	assert.False(t, vite.IsRunningHot())
}

func TestHotFile(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")

	assert.Equal(t, filepath.Join(public, "hot"), vite.HotFile())
	assert.False(t, vite.IsRunningHot())

	writeFile(t, filepath.Join(public, "hot"), "http://localhost:5173\n")
	assert.True(t, vite.IsRunningHot())

	// A relative hot file is resolved against the project root.
	vite.UseHotFile("storage/hot")
	assert.Equal(t, filepath.Join(support.RelativePath, "storage", "hot"), vite.HotFile())
	assert.False(t, vite.IsRunningHot())

	// A leading slash is trimmed, so an absolute-looking path is read under the
	// project root and cannot escape it.
	vite.UseHotFile("/srv/app/hot")
	assert.Equal(t, filepath.Join(support.RelativePath, "srv", "app", "hot"), vite.HotFile())
	assert.False(t, vite.IsRunningHot())

	// A genuine absolute path is not used as is either; it is read under the project root.
	absolute := filepath.Join(t.TempDir(), "hot")
	writeFile(t, absolute, "http://localhost:5173")
	vite.UseHotFile(absolute)
	assert.NotEqual(t, absolute, vite.HotFile())
	assert.Equal(t, filepath.Join(support.RelativePath, absolute), vite.HotFile())
	assert.False(t, vite.IsRunningHot())

	// HotFile is absolute, so feeding it back resolves under the project root again
	// rather than trusting the filesystem root.
	vite.UseHotFile(vite.HotFile())
	assert.True(t, strings.HasPrefix(vite.HotFile(), support.RelativePath+string(filepath.Separator)))

	// A directory is not a hot file.
	require.NoError(t, os.MkdirAll(filepath.Join(support.RelativePath, "storage"), 0o755))
	vite.UseHotFile("storage")
	assert.False(t, vite.IsRunningHot())
	vite.UseHotFile(filepath.Dir(absolute))
	assert.False(t, vite.IsRunningHot())
}

func TestPathsResolvedOnce(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")
	vite.UseHotFile("public/hot").UseBuildDirectory("dist")

	hot, build := vite.HotFile(), vite.buildPath()
	assert.Equal(t, filepath.Join(public, "hot"), hot)
	assert.Equal(t, filepath.Join(support.RelativePath, "dist"), build)

	// The paths are fixed when set, so a render never resolves them again.
	support.RelativePath = t.TempDir()
	assert.Equal(t, hot, vite.HotFile())
	assert.Equal(t, build, vite.buildPath())
}

func TestTags_Hot(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")
	writeFile(t, filepath.Join(public, "hot"), "http://[::1]:5173/\n")

	tags, err := vite.Tags("resources/css/app.css", "/resources/js/app.js")
	require.NoError(t, err)
	assert.Equal(t, template.HTML(
		`<script type="module" src="http://[::1]:5173/@vite/client"></script>`+
			`<link rel="stylesheet" href="http://[::1]:5173/resources/css/app.css">`+
			`<script type="module" src="http://[::1]:5173/resources/js/app.js"></script>`,
	), tags)

	// The manifest is not consulted while the dev server runs.
	url, err := vite.Asset("resources/images/logo.svg")
	require.NoError(t, err)
	assert.Equal(t, "http://[::1]:5173/resources/images/logo.svg", url)

	assert.Equal(t, template.HTML(
		`<script type="module">import RefreshRuntime from 'http://[::1]:5173/@react-refresh';`+
			`RefreshRuntime.injectIntoGlobalHook(window);`+
			`window.$RefreshReg$ = () => {};`+
			`window.$RefreshSig$ = () => (type) => type;`+
			`window.__vite_plugin_react_preamble_installed__ = true;</script>`,
	), vite.ReactRefresh())
}

func TestTags_HotEscapesURL(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")
	writeFile(t, filepath.Join(public, "hot"), `http://localhost:5173/"><script>alert(1)</script>`)

	tags, err := vite.Tags("resources/js/app.js")
	require.NoError(t, err)
	assert.NotContains(t, string(tags), "<script>alert(1)</script>")
	assert.Contains(t, string(tags), "&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt;")

	// The preamble places the URL inside a JavaScript string, so it is escaped as one.
	preamble := string(vite.ReactRefresh())
	assert.NotContains(t, preamble, `"><script>alert(1)</script>`)
	assert.Contains(t, preamble, `from 'http://localhost:5173/\"\u003E\u003Cscript\u003Ealert(1)\u003C/script\u003E/@react-refresh'`)
}

func TestTags_Build(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), sampleManifest)

	tests := []struct {
		name     string
		assetURL string
		entries  []string
		expected string
		err      error
	}{
		{
			name:    "entry with css and nested imports",
			entries: []string{"resources/js/app.js"},
			expected: `<link rel="modulepreload" href="/build/assets/vendor-1f2e3d4c.js">` +
				`<link rel="modulepreload" href="/build/assets/shared-9c8b7a6f.js">` +
				`<link rel="stylesheet" href="/build/assets/app-9a1b2c3d.css">` +
				`<link rel="stylesheet" href="/build/assets/vendor-aabbccdd.css">` +
				`<script type="module" src="/build/assets/app-4ed993c7.js"></script>`,
		},
		{
			name:     "css entry",
			entries:  []string{"resources/css/site.css"},
			expected: `<link rel="stylesheet" href="/build/assets/site-0a0b0c0d.css">`,
		},
		{
			name:     "leading slash resolves like the dev server path",
			entries:  []string{"/resources/css/site.css"},
			expected: `<link rel="stylesheet" href="/build/assets/site-0a0b0c0d.css">`,
		},
		{
			name:    "several entries share chunks once, stylesheets before scripts",
			entries: []string{"resources/css/site.css", "resources/js/app.js", "resources/js/admin.js"},
			expected: `<link rel="modulepreload" href="/build/assets/vendor-1f2e3d4c.js">` +
				`<link rel="modulepreload" href="/build/assets/shared-9c8b7a6f.js">` +
				`<link rel="stylesheet" href="/build/assets/site-0a0b0c0d.css">` +
				`<link rel="stylesheet" href="/build/assets/app-9a1b2c3d.css">` +
				`<link rel="stylesheet" href="/build/assets/vendor-aabbccdd.css">` +
				`<script type="module" src="/build/assets/app-4ed993c7.js"></script>` +
				`<script type="module" src="/build/assets/admin-5fe6a7b8.js"></script>`,
		},
		{
			name:     "asset url prefix",
			assetURL: "https://cdn.example.com/",
			entries:  []string{"resources/css/site.css"},
			expected: `<link rel="stylesheet" href="https://cdn.example.com/build/assets/site-0a0b0c0d.css">`,
		},
		{
			name:     "no entries",
			entries:  nil,
			expected: "",
		},
		{
			name:    "unknown entry",
			entries: []string{"resources/js/missing.js"},
			err:     errors.ViteEntryNotFound.Args("resources/js/missing.js"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vite := newTestVite(t, test.assetURL)

			tags, err := vite.Tags(test.entries...)
			if test.err != nil {
				assert.EqualError(t, err, test.err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, template.HTML(test.expected), tags)
		})
	}
}

func TestTags_ImportCycle(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), `{
		"a.js": {"file": "a.js", "isEntry": true, "imports": ["b.js", "missing.js"]},
		"b.js": {"file": "b.js", "imports": ["a.js"]}
	}`)
	vite := newTestVite(t, "")

	tags, err := vite.Tags("a.js")
	require.NoError(t, err)
	assert.Equal(t, template.HTML(
		`<link rel="modulepreload" href="/build/b.js">`+
			`<script type="module" src="/build/a.js"></script>`,
	), tags)
}

func TestAssetAndContent(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), sampleManifest)
	writeFile(t, filepath.Join(public, "build", "assets", "site-0a0b0c0d.css"), "body{margin:0}")
	vite := newTestVite(t, "")

	url, err := vite.Asset("resources/images/logo.svg")
	require.NoError(t, err)
	assert.Equal(t, "/build/assets/logo-11223344.svg", url)

	content, err := vite.Content("resources/css/site.css")
	require.NoError(t, err)
	assert.Equal(t, "body{margin:0}", content)

	// A leading slash is accepted everywhere, not only on the dev server.
	url, err = vite.Asset("/resources/images/logo.svg")
	require.NoError(t, err)
	assert.Equal(t, "/build/assets/logo-11223344.svg", url)

	content, err = vite.Content("/resources/css/site.css")
	require.NoError(t, err)
	assert.Equal(t, "body{margin:0}", content)

	_, err = vite.Asset("missing.svg")
	assert.EqualError(t, err, errors.ViteEntryNotFound.Args("missing.svg").Error())

	_, err = vite.Content("missing.svg")
	assert.EqualError(t, err, errors.ViteEntryNotFound.Args("missing.svg").Error())

	_, err = vite.Content("resources/images/logo.svg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read vite asset assets/logo-11223344.svg")

	// Content comes from the build even while the dev server runs.
	writeFile(t, filepath.Join(public, "hot"), "http://localhost:5173")
	content, err = vite.Content("resources/css/site.css")
	require.NoError(t, err)
	assert.Equal(t, "body{margin:0}", content)
}

func TestContent_RejectsPathOutsideBuild(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "secret.txt"), "secret")
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), `{
  "resources/js/escape.js": {"file": "../secret.txt", "isEntry": true},
  "resources/js/sibling.js": {"file": "assets/../../secret.txt", "isEntry": true}
}`)
	vite := newTestVite(t, "")

	for entry, file := range map[string]string{
		"resources/js/escape.js":  "../secret.txt",
		"resources/js/sibling.js": "assets/../../secret.txt",
	} {
		_, err := vite.Content(entry)
		assert.EqualError(t, err, errors.ViteAssetOutsideBuild.Args(file).Error(), entry)
	}
}

func TestManifest_Missing(t *testing.T) {
	public := setupProject(t)
	vite := newTestVite(t, "")

	_, err := vite.Tags("resources/js/app.js")
	assert.EqualError(t, err, errors.ViteManifestNotFound.Args(filepath.Join(public, "build", ".vite", "manifest.json")).Error())

	// The error names the configured build directory and manifest.
	vite.UseBuildDirectory("dist").UseManifestFilename("assets.json")
	_, err = vite.Tags("resources/js/app.js")
	assert.EqualError(t, err, errors.ViteManifestNotFound.Args(filepath.Join(support.RelativePath, "dist", "assets.json")).Error())

	// Asset and Content report the same error when there is no hot server to fall back on.
	_, err = vite.Asset("resources/images/logo.svg")
	assert.EqualError(t, err, errors.ViteManifestNotFound.Args(filepath.Join(support.RelativePath, "dist", "assets.json")).Error())
	_, err = vite.Content("resources/images/logo.svg")
	assert.EqualError(t, err, errors.ViteManifestNotFound.Args(filepath.Join(support.RelativePath, "dist", "assets.json")).Error())
}

func TestManifest_Unreadable(t *testing.T) {
	// A directory at the manifest path passes the existence check but cannot be read.
	public := setupProject(t)
	manifestPath := filepath.Join(public, "build", ".vite", "manifest.json")
	require.NoError(t, os.MkdirAll(manifestPath, 0o755))
	vite := newTestVite(t, "")

	_, err := vite.Tags("resources/js/app.js")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[vite] failed to read vite manifest "+manifestPath+": ")
}

func TestManifest_Invalid(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), "{not json")
	vite := newTestVite(t, "")

	_, err := vite.Tags("resources/js/app.js")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[vite] vite manifest "+filepath.Join(public, "build", ".vite", "manifest.json")+" is invalid")
}

func TestManifest_ReloadedWhenChanged(t *testing.T) {
	public := setupProject(t)
	manifestPath := filepath.Join(public, "build", ".vite", "manifest.json")
	writeFile(t, manifestPath, `{"a.js": {"file": "a-1.js"}}`)
	vite := newTestVite(t, "")

	url, err := vite.Asset("a.js")
	require.NoError(t, err)
	assert.Equal(t, "/build/a-1.js", url)

	// Same size and modification time: the cached manifest is kept.
	first, ok := vite.manifests[manifestPath]
	require.True(t, ok)
	_, err = vite.Asset("a.js")
	require.NoError(t, err)
	assert.Same(t, first, vite.manifests[manifestPath])

	// A rebuild changes the file and is picked up without a restart.
	writeFile(t, manifestPath, `{"a.js": {"file": "a-2.js"}, "b.js": {"file": "b-1.js"}}`)
	url, err = vite.Asset("a.js")
	require.NoError(t, err)
	assert.Equal(t, "/build/a-2.js", url)
	assert.NotSame(t, first, vite.manifests[manifestPath])
}

func TestManifestHash_FromManifest(t *testing.T) {
	public := setupProject(t)
	manifestPath := filepath.Join(public, "build", ".vite", "manifest.json")
	writeFile(t, manifestPath, sampleManifest)
	vite := newTestVite(t, "")

	sum := md5.Sum([]byte(sampleManifest))
	assert.Equal(t, hex.EncodeToString(sum[:]), vite.ManifestHash())

	// Tags reuses the manifest hash, so a render leaves ManifestHash unchanged.
	cached := vite.manifests[manifestPath]
	_, err := vite.Tags("resources/js/app.js")
	require.NoError(t, err)
	assert.Same(t, cached, vite.manifests[manifestPath])
	assert.Equal(t, hex.EncodeToString(sum[:]), vite.ManifestHash())
}

func TestManifestHash_EmptyWithoutManifest(t *testing.T) {
	setupProject(t)
	vite := newTestVite(t, "")

	assert.Equal(t, "", vite.ManifestHash())
}

func TestManifestHash_EmptyWhileRunningHot(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), sampleManifest)
	writeFile(t, filepath.Join(public, "hot"), "http://localhost:5173")
	vite := newTestVite(t, "")

	assert.True(t, vite.IsRunningHot())
	assert.Equal(t, "", vite.ManifestHash())
}

func TestManifestHash_InvalidManifest(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), "{not json")
	vite := newTestVite(t, "")

	assert.Equal(t, "", vite.ManifestHash())
}

func TestManifestHash_MovesWithRebuild(t *testing.T) {
	public := setupProject(t)
	manifestPath := filepath.Join(public, "build", ".vite", "manifest.json")
	writeFile(t, manifestPath, `{"a.js": {"file": "a-1.js"}}`)
	vite := newTestVite(t, "")

	firstHash := vite.ManifestHash()
	assert.NotEmpty(t, firstHash)
	tags, err := vite.Tags("a.js")
	require.NoError(t, err)
	assert.Contains(t, string(tags), "a-1.js")

	// A rebuild that only swaps the content keeps the size, so the mtime is what
	// invalidates the cache.
	sameLength := `{"a.js": {"file": "a-2.js"}}`
	writeFile(t, manifestPath, sameLength)
	info, err := os.Stat(manifestPath)
	require.NoError(t, err)
	future := info.ModTime().Add(time.Second)
	require.NoError(t, os.Chtimes(manifestPath, future, future))

	secondHash := vite.ManifestHash()
	sum := md5.Sum([]byte(sameLength))
	assert.Equal(t, hex.EncodeToString(sum[:]), secondHash)
	assert.NotEqual(t, firstHash, secondHash)
	tags, err = vite.Tags("a.js")
	require.NoError(t, err)
	assert.Contains(t, string(tags), "a-2.js")

	// A rebuild that changes the size is picked up as well.
	writeFile(t, manifestPath, `{"a.js": {"file": "a-3.js"}, "b.js": {"file": "b-1.js"}}`)
	assert.NotEqual(t, secondHash, vite.ManifestHash())
	tags, err = vite.Tags("a.js")
	require.NoError(t, err)
	assert.Contains(t, string(tags), "a-3.js")
}

func TestBuildDirectoryNormalized(t *testing.T) {
	public := setupProject(t)

	tests := []struct {
		name      string
		value     string
		directory string
		root      string
	}{
		{name: "bare", value: "build", directory: "build", root: filepath.Join(support.RelativePath, "build")},
		{name: "leading slash", value: "/build", directory: "build", root: filepath.Join(support.RelativePath, "build")},
		{name: "public prefix", value: "public/build", directory: "build", root: filepath.Join(public, "build")},
		{name: "slash and public prefix", value: "/public/build", directory: "build", root: filepath.Join(public, "build")},
		{name: "public directory itself", value: "public", directory: "", root: public},
		{name: "outside the public directory", value: "dist", directory: "dist", root: filepath.Join(support.RelativePath, "dist")},
		{name: "nested project path", value: "a/build", directory: "a/build", root: filepath.Join(support.RelativePath, "a", "build")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := mocksconfig.NewConfig(t)
			config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return(DefaultHotFile()).Once()
			config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return(test.value).Once()
			config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return(DefaultManifestFilename).Once()

			vite := NewVite(config)
			assert.Equal(t, test.directory, vite.BuildDirectory())
			assert.Equal(t, test.root, vite.buildPath())

			// The setter normalizes the same spellings as the constructor.
			vite.UseBuildDirectory(test.value)
			assert.Equal(t, test.directory, vite.BuildDirectory())
			assert.Equal(t, test.root, vite.buildPath())
		})
	}
}

func TestHotFileNormalized(t *testing.T) {
	public := setupProject(t)

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "public prefix", value: "public/hot", expected: filepath.Join(public, "hot")},
		{name: "slash and public prefix", value: "/public/hot", expected: filepath.Join(public, "hot")},
		{name: "bare", value: "hot", expected: filepath.Join(support.RelativePath, "hot")},
		{name: "leading slash", value: "/storage/hot", expected: filepath.Join(support.RelativePath, "storage", "hot")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := mocksconfig.NewConfig(t)
			config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return(test.value).Once()
			config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return(DefaultBuildDirectory()).Once()
			config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return(DefaultManifestFilename).Once()

			vite := NewVite(config)
			assert.Equal(t, test.expected, vite.HotFile())

			// The setter normalizes the same spellings as the constructor.
			vite.UseHotFile(test.value)
			assert.Equal(t, test.expected, vite.HotFile())
		})
	}
}

// TestPathsWithCustomPublicDirectory pins that both paths are read from the
// project root, while only how the build directory is reported follows the
// configured public directory.
func TestPathsWithCustomPublicDirectory(t *testing.T) {
	setupProject(t)
	public := support.Config.Paths.Public
	support.Config.Paths.Public = "static"
	t.Cleanup(func() { support.Config.Paths.Public = public })

	newVite := func(hotFile, buildDirectory string) *Vite {
		config := mocksconfig.NewConfig(t)
		config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return(hotFile).Once()
		config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return(buildDirectory).Once()
		config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return(DefaultManifestFilename).Once()

		return NewVite(config)
	}

	// The defaults follow the configured public directory.
	assert.Equal(t, "static/hot", DefaultHotFile())
	assert.Equal(t, "static/build", DefaultBuildDirectory())
	vite := newVite(DefaultHotFile(), DefaultBuildDirectory())
	assert.Equal(t, "build", vite.BuildDirectory())
	assert.Equal(t, filepath.Join(support.RelativePath, "static", "build"), vite.buildPath())

	// "public/hot" is a project-root path, not an alias for the configured directory.
	vite = newVite("/public/hot", "static/build")
	assert.Equal(t, filepath.Join(support.RelativePath, "public", "hot"), vite.HotFile())
	assert.Equal(t, "build", vite.BuildDirectory())
	assert.Equal(t, filepath.Join(support.RelativePath, "static", "build"), vite.buildPath())

	// A "public/" prefix is a directory name here, not an alias: it stays literal.
	vite = newVite("/static/hot", "public/build")
	assert.Equal(t, "public/build", vite.BuildDirectory())
	assert.Equal(t, filepath.Join(support.RelativePath, "public", "build"), vite.buildPath())

	// An empty public directory leaves both values untouched, never the filesystem root.
	support.Config.Paths.Public = ""
	assert.Equal(t, "hot", DefaultHotFile())
	assert.Equal(t, "build", DefaultBuildDirectory())
	vite = newVite("/public/hot", "public/build")
	assert.Equal(t, filepath.Join(support.RelativePath, "public", "hot"), vite.HotFile())
	assert.Equal(t, "public/build", vite.BuildDirectory())
	assert.Equal(t, filepath.Join(support.RelativePath, "public", "build"), vite.buildPath())
}

func TestUseBuildDirectory(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "dist", ".vite", "manifest.json"), `{"a.js": {"file": "a.js"}}`)
	vite := newTestVite(t, "")

	url, err := vite.UseBuildDirectory("/public/dist/").Asset("a.js")
	require.NoError(t, err)
	assert.Equal(t, "/dist/a.js", url)
}

func TestAssetBase(t *testing.T) {
	setupProject(t)
	vite := newTestVite(t, "")

	// The default "public/build" is served at /build because public is the web root.
	assert.Equal(t, "/build", vite.assetBase())

	public := support.Config.Paths.Public
	support.Config.Paths.Public = "static"
	t.Cleanup(func() { support.Config.Paths.Public = public })

	// The URL segment is fixed when the directory is set, like the path read from disk.
	assert.Equal(t, "build", vite.BuildDirectory())
	assert.Equal(t, "/build", vite.assetBase())

	vite.UseBuildDirectory("static/build")
	assert.Equal(t, "/build", vite.assetBase())
}

func TestFuncs(t *testing.T) {
	public := setupProject(t)
	writeFile(t, filepath.Join(public, "build", ".vite", "manifest.json"), sampleManifest)
	vite := newTestVite(t, "")

	tmpl, err := template.New("page").Funcs(Funcs(vite)).Parse(
		`<head>{{ vite "resources/css/site.css" "resources/js/app.js" }}{{ viteReactRefresh }}</head>` +
			`<img src="{{ viteAsset "resources/images/logo.svg" }}">`,
	)
	require.NoError(t, err)

	var out strings.Builder
	require.NoError(t, tmpl.Execute(&out, nil))
	assert.Equal(t,
		`<head><link rel="modulepreload" href="/build/assets/vendor-1f2e3d4c.js">`+
			`<link rel="modulepreload" href="/build/assets/shared-9c8b7a6f.js">`+
			`<link rel="stylesheet" href="/build/assets/site-0a0b0c0d.css">`+
			`<link rel="stylesheet" href="/build/assets/app-9a1b2c3d.css">`+
			`<link rel="stylesheet" href="/build/assets/vendor-aabbccdd.css">`+
			`<script type="module" src="/build/assets/app-4ed993c7.js"></script></head>`+
			`<img src="/build/assets/logo-11223344.svg">`,
		out.String(),
	)

	// A missing entry aborts rendering instead of emitting a broken page.
	tmpl, err = template.New("page").Funcs(Funcs(vite)).Parse(`{{ vite "resources/js/missing.js" }}`)
	require.NoError(t, err)
	out.Reset()
	err = tmpl.Execute(&out, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to locate file in vite manifest: resources/js/missing.js")
}

func TestIsCSS(t *testing.T) {
	for _, name := range []string{"app.css", "a/b.SCSS", "x.less", "x.sass", "x.styl", "x.stylus", "x.pcss", "x.postcss"} {
		assert.True(t, isCSS(name), name)
	}
	for _, name := range []string{"app.js", "app.ts", "app.css.js", "logo.svg", "noext"} {
		assert.False(t, isCSS(name), name)
	}
}
