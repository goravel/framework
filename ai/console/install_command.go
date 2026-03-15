package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type AgentsInstallCommand struct {
	manifestFetcher func(branch string) ([]ManifestEntry, error)
	fetcher         func(branch, path string) ([]byte, error)
}

func NewAgentsInstallCommand() *AgentsInstallCommand {
	return &AgentsInstallCommand{
		manifestFetcher: fetchManifest,
		fetcher:         fetchRaw,
	}
}

func (r *AgentsInstallCommand) Signature() string {
	return "agents:install"
}

func (r *AgentsInstallCommand) Description() string {
	return "Install AI agent skill files for the current Goravel version"
}

func (r *AgentsInstallCommand) Extend() command.Extend {
	return command.Extend{
		Category: "agents",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "version",
				Usage: "Override detected Goravel version (e.g. v1.17)",
			},
			&command.BoolFlag{
				Name:               "force",
				Value:              false,
				Usage:              "Skip confirmation and overwrite existing files",
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "all",
				Aliases:            []string{"a"},
				Value:              false,
				Usage:              "Install all available facade agent files",
				DisableDefaultText: true,
			},
		},
	}
}

func (r *AgentsInstallCommand) Handle(ctx console.Context) error {
	version, branch, err := r.resolveVersionAndBranch(ctx)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	entries, branch, err := r.resolveManifest(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if len(entries) == 0 {
		ctx.Error(fmt.Sprintf("No agent files found for version %s. Check https://github.com/goravel/docs", version))
		return nil
	}

	if !ctx.OptionBool("force") {
		if _, statErr := os.Stat(versionFilePath); statErr == nil {
			if !ctx.Confirm("Agent files are already installed. Overwrite?") {
				ctx.Warning("Cancelled.")
				return nil
			}
		}
	}

	facadeArgs := ctx.Arguments()
	var toInstall []ManifestEntry
	switch {
	case ctx.OptionBool("all"):
		toInstall = entries
	case len(facadeArgs) > 0:
		toInstall = entriesForFacades(entries, facadeArgs)
		if len(toInstall) == 0 {
			ctx.Error(fmt.Sprintf("No agent files found for facade(s): %s", strings.Join(facadeArgs, ", ")))
			return nil
		}
	default:
		toInstall = defaultEntries(entries)
	}

	type downloadResult struct {
		path    string
		content []byte
		err     error
	}

	ch := make(chan downloadResult, len(toInstall))
	for _, entry := range toInstall {
		go func(e ManifestEntry) {
			content, fetchErr := r.fetcher(branch, e.Path)
			ch <- downloadResult{path: e.Path, content: content, err: fetchErr}
		}(entry)
	}

	downloaded := make(map[string][]byte)
	for range toInstall {
		res := <-ch
		if res.err != nil {
			ctx.Error(res.err.Error())
			return nil
		}
		if res.content == nil {
			ctx.Error(fmt.Sprintf("File not found upstream: %s", res.path))
			return nil
		}
		downloaded[res.path] = res.content
	}

	existing, _ := readVersionFile()
	local := VersionFile{Version: version, Files: make(map[string]string)}
	for k, v := range existing.Files {
		local.Files[k] = v
	}

	for path, content := range downloaded {
		if err := writeAgentFile(path, content); err != nil {
			ctx.Error(fmt.Sprintf("Failed to write %s: %v", path, err))
			return nil
		}
		local.Files[path] = sha256sum(content)
	}

	if err := os.MkdirAll(".ai/skills", 0755); err != nil {
		ctx.Error(fmt.Sprintf("Failed to create .ai/skills: %v", err))
		return nil
	}

	if err := writeVersionFile(local); err != nil {
		ctx.Error(fmt.Sprintf("Failed to write .version: %v", err))
		return nil
	}

	ctx.Info(fmt.Sprintf("Installed %d file(s) for version %s.", len(downloaded), version))
	return nil
}

// resolveVersionAndBranch determines the framework version and docs branch.
// Precedence: --version flag → go.mod detection → interactive picker.
func (r *AgentsInstallCommand) resolveVersionAndBranch(ctx console.Context) (version, branch string, err error) {
	version = ctx.Option("version")
	if version != "" {
		if !isSupportedVersion(version) {
			return "", "", fmt.Errorf("agent files are only available for Goravel v1.17 and above (got %s)", version)
		}
		return version, resolveBranch(version), nil
	}

	version, err = detectGoravelVersion()
	if err == nil {
		if !isSupportedVersion(version) {
			return "", "", fmt.Errorf("agent files are only available for Goravel v1.17 and above (got %s)", version)
		}
		return version, resolveBranch(version), nil
	}

	available, fetchErr := fetchAvailableBranches()
	if fetchErr != nil || len(available) == 0 {
		return "", "", fmt.Errorf("cannot detect Goravel version from go.mod. Use --version to specify it")
	}

	choices := make([]console.Choice, len(available))
	for i, v := range available {
		choices[i] = console.Choice{Key: v, Value: v}
	}

	version, err = ctx.Choice("Select Goravel version to install agent files for", choices)
	if err != nil {
		return "", "", fmt.Errorf("version selection failed: %w", err)
	}

	return version, resolveBranch(version), nil
}

// resolveManifest fetches the manifest for the given branch, falling back to master
// if the version branch has no manifest.
func (r *AgentsInstallCommand) resolveManifest(branch string) ([]ManifestEntry, string, error) {
	entries, err := r.manifestFetcher(branch)
	if err != nil {
		return nil, branch, err
	}
	if len(entries) == 0 && branch != docsFallbackBranch {
		entries, err = r.manifestFetcher(docsFallbackBranch)
		return entries, docsFallbackBranch, err
	}
	return entries, branch, nil
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
