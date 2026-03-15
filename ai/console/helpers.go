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
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	versionFilePath    = ".ai/.version"
	docsFallbackBranch = "master"
)

// VersionFile is the local .ai/.version tracking file.
// It is never fetched from goravel/docs — it is created and maintained by these commands.
// files maps each relative path (e.g. "prompt/route.md") to the SHA256 of its content
// at the time it was installed or last updated. Used to detect both upstream changes
// and local user modifications during agents:update.
type VersionFile struct {
	Version string            `json:"version"`
	Files   map[string]string `json:"files"`
}

type githubBranch struct {
	Name string `json:"name"`
}

type gitTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type gitTreeResponse struct {
	Tree []gitTreeEntry `json:"tree"`
}

// isSupportedVersion reports whether a version string has agent file support.
// Agent files were introduced in Goravel v1.17. "master" and "latest" are always accepted.
func isSupportedVersion(version string) bool {
	if version == docsFallbackBranch || version == "latest" {
		return true
	}
	major, minor := parseVersionParts(version)
	if major > 1 {
		return true
	}
	return major == 1 && minor >= 17
}

// resolveBranch maps a framework version to its goravel/docs branch.
// "latest" is an alias for master. All other versions use their string as the branch name.
func resolveBranch(version string) string {
	if version == "latest" {
		return docsFallbackBranch
	}
	return version
}

// encodeBranchForURL percent-encodes characters that break URLs (e.g. #) while
// preserving forward slashes, which are valid in git branch names and GitHub URLs.
func encodeBranchForURL(branch string) string {
	encoded := url.PathEscape(branch)
	return strings.ReplaceAll(encoded, "%2F", "/")
}

// fetchFileTree lists all downloadable files under .ai/ in the goravel/docs repo
// for the given branch. Returns paths relative to .ai/ (e.g. "AGENTS.md", "prompt/route.md").
func fetchFileTree(branch string) ([]string, error) {
	encodedBranch := encodeBranchForURL(branch)
	apiURL := fmt.Sprintf("https://api.github.com/repos/goravel/docs/git/trees/%s?recursive=1", encodedBranch)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", apiURL, resp.StatusCode)
	}

	var tree gitTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("decode tree: %w", err)
	}

	var paths []string
	for _, entry := range tree.Tree {
		if entry.Type != "blob" {
			continue
		}
		if !strings.HasPrefix(entry.Path, ".ai/") {
			continue
		}
		rel := strings.TrimPrefix(entry.Path, ".ai/")
		if rel == "" {
			continue
		}
		paths = append(paths, rel)
	}

	return paths, nil
}

// fetchAvailableBranches returns versioned branches (v1.17+) from goravel/docs,
// sorted newest first. Used for the interactive version picker when go.mod detection fails.
func fetchAvailableBranches() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/goravel/docs/branches?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available versions: %w", err)
	}
	defer resp.Body.Close()

	var branches []githubBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, fmt.Errorf("failed to parse available versions: %w", err)
	}

	re := regexp.MustCompile(`^v\d+\.\d+$`)
	var versions []string
	for _, b := range branches {
		if re.MatchString(b.Name) && isSupportedVersion(b.Name) {
			versions = append(versions, b.Name)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		maj1, min1 := parseVersionParts(versions[i])
		maj2, min2 := parseVersionParts(versions[j])
		if maj1 != maj2 {
			return maj1 > maj2
		}
		return min1 > min2
	})

	return versions, nil
}

func parseVersionParts(v string) (int, int) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 2)
	if len(parts) < 2 {
		return 0, 0
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	return major, minor
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

	re := regexp.MustCompile(`github\.com/goravel/framework\s+v(\d+)\.(\d+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); m != nil {
			return fmt.Sprintf("v%s.%s", m[1], m[2]), nil
		}
	}
	return "", fmt.Errorf("github.com/goravel/framework not found in %s", gomodPath)
}

func fetchRaw(branch, path string) ([]byte, error) {
	encodedBranch := encodeBranchForURL(branch)
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/goravel/docs/%s/.ai/%s", encodedBranch, path)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL) //nolint:noctx
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

func sha256sum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
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

// filterPaths returns only paths whose base filename (without extension) matches filter.
// Returns all paths when filter is empty.
func filterPaths(paths []string, filter string) []string {
	if filter == "" {
		return paths
	}
	var filtered []string
	for _, p := range paths {
		base := filepath.Base(p)
		baseName := strings.TrimSuffix(base, filepath.Ext(base))
		if baseName == filter {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func writeAgentFile(key string, content []byte) error {
	dest := destPathFor(key)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	return os.WriteFile(dest, content, 0644)
}
