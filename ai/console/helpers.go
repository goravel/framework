package console

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	versionFilePath    = ".ai/.version"
	docsFallbackBranch = "master"
)

// VersionFile is the local .ai/.version tracking file.
// It maps each relative path to its installed SHA256 content hash.
type VersionFile struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

// ManifestEntry describes a single AI doc file available upstream.
type ManifestEntry struct {
	Facade  string `json:"facade"`
	Path    string `json:"path"`
	Default bool   `json:"default"`
}

func detectGoravelVersion() (string, error) {
	return detectGoravelVersionFrom("go.mod")
}

func detectGoravelVersionFrom(gomodPath string) (string, error) {
	f, err := os.Open(gomodPath)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", gomodPath, err)
	}
	defer f.Close()

	re := regexp.MustCompile(`^\s*(?:require\s+)?github\.com/goravel/framework\s+v(\d+)\.(\d+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			return fmt.Sprintf("v%s.%s", m[1], m[2]), nil
		}
	}
	return "", fmt.Errorf("github.com/goravel/framework not found in %s", gomodPath)
}

func isSupportedVersion(version string) bool {
	if version == docsFallbackBranch || version == "latest" {
		return true
	}
	v := strings.TrimPrefix(version, "v")
	parts := strings.SplitN(v, ".", 2)
	if len(parts) < 2 {
		return false
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	return major > 1 || (major == 1 && minor >= 17)
}

func resolveBranch(version string) string {
	if version == "latest" {
		return docsFallbackBranch
	}
	return version
}

func fetchManifest(branch string) ([]ManifestEntry, error) {
	data, err := fetchRaw(branch, "manifest.json")
	if err != nil || data == nil {
		// Fallback to master if specific branch manifest doesn't exist
		if branch != docsFallbackBranch {
			return fetchManifest(docsFallbackBranch)
		}
		return nil, err
	}
	var entries []ManifestEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return entries, nil
}

func fetchRaw(branch, path string) ([]byte, error) {
	encodedBranch := strings.ReplaceAll(url.PathEscape(branch), "%2F", "/")
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/goravel/docs/%s/.ai/%s", encodedBranch, path)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", rawURL, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func downloadFiles(branch string, toInstall []ManifestEntry, fetcher func(string, string) ([]byte, error)) (map[string][]byte, error) {
	type result struct {
		path    string
		content []byte
		err     error
	}

	ch := make(chan result, len(toInstall))
	for _, entry := range toInstall {
		go func(e ManifestEntry) {
			content, err := fetcher(branch, e.Path)
			ch <- result{path: e.Path, content: content, err: err}
		}(entry)
	}

	downloaded := make(map[string][]byte)
	for range toInstall {
		res := <-ch
		if res.err != nil {
			return nil, res.err
		}
		if res.content == nil {
			return nil, fmt.Errorf("file not found upstream: %s", res.path)
		}
		downloaded[res.path] = res.content
	}
	return downloaded, nil
}

func saveFiles(version string, downloaded map[string][]byte) error {
	existing, _ := readVersionFile()
	local := VersionFile{Version: version, Files: make(map[string]string)}

	for k, v := range existing.Files {
		local.Files[k] = v
	}

	if err := os.MkdirAll(".ai/skills", 0755); err != nil {
		return err
	}

	for path, content := range downloaded {
		if err := writeAgentFile(path, content); err != nil {
			return err
		}
		local.Files[path] = sha256sum(content)
	}

	return writeVersionFile(local)
}

func readVersionFile() (VersionFile, error) {
	data, err := os.ReadFile(versionFilePath)
	if os.IsNotExist(err) {
		return VersionFile{Files: make(map[string]string)}, nil
	}
	if err != nil {
		return VersionFile{}, err
	}
	var v VersionFile
	if err := json.Unmarshal(data, &v); err != nil {
		return VersionFile{}, err
	}
	if v.Files == nil {
		v.Files = make(map[string]string)
	}
	return v, nil
}

func writeVersionFile(v VersionFile) error {
	if err := os.MkdirAll(filepath.Dir(versionFilePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(versionFilePath, data, 0644)
}

func destPathFor(key string) string {
	if key == "AGENTS.md" {
		return "AGENTS.md"
	}
	return filepath.Join(".ai", key)
}

func writeAgentFile(key string, content []byte) error {
	dest := destPathFor(key)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.WriteFile(dest, content, 0644)
}

func sha256sum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func entriesForFacades(entries []ManifestEntry, facades []string) []ManifestEntry {
	set := make(map[string]bool, len(facades))
	for _, f := range facades {
		set[f] = true
	}
	var out []ManifestEntry
	for _, e := range entries {
		if set[e.Facade] {
			out = append(out, e)
		}
	}
	return out
}

func defaultEntries(entries []ManifestEntry) []ManifestEntry {
	var out []ManifestEntry
	for _, e := range entries {
		if e.Default {
			out = append(out, e)
		}
	}
	return out
}

func installedEntries(entries []ManifestEntry, installedFiles map[string]string) []ManifestEntry {
	var out []ManifestEntry
	for _, e := range entries {
		if _, ok := installedFiles[e.Path]; ok {
			out = append(out, e)
		}
	}
	return out
}
