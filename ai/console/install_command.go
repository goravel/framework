package console

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type AgentsInstallCommand struct {
	treeFetcher func(branch string) ([]string, error)
	fetcher     func(branch, path string) ([]byte, error)
}

func NewAgentsInstallCommand() *AgentsInstallCommand {
	return &AgentsInstallCommand{
		treeFetcher: fetchFileTree,
		fetcher:     fetchRaw,
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
			&command.StringFlag{
				Name:  "file",
				Usage: "Install only one prompt file (e.g. route)",
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

	paths, branch, err := r.resolveFilePaths(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if len(paths) == 0 {
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

	filter := ctx.Option("file")
	pathsToInstall := filterPaths(paths, filter)
	if filter != "" && len(pathsToInstall) == 0 {
		ctx.Error(fmt.Sprintf("No file matching '%s' found in remote repository.", filter))

		return nil
	}

	type downloadResult struct {
		key     string
		content []byte
		err     error
	}

	ch := make(chan downloadResult, len(pathsToInstall))
	for _, key := range pathsToInstall {
		go func(k string) {
			content, fetchErr := r.fetcher(branch, k)
			ch <- downloadResult{key: k, content: content, err: fetchErr}
		}(key)
	}

	downloaded := make(map[string][]byte)
	for range pathsToInstall {
		res := <-ch
		if res.err != nil {
			ctx.Error(res.err.Error())
			return nil
		}
		if res.content == nil {
			ctx.Error(fmt.Sprintf("File not found upstream: %s", res.key))
			return nil
		}
		downloaded[res.key] = res.content
	}

	existing, _ := readVersionFile()
	local := VersionFile{Version: version, Files: make(map[string]string)}
	for k, v := range existing.Files {
		local.Files[k] = v
	}

	for key, content := range downloaded {
		if err := writeAgentFile(key, content); err != nil {
			ctx.Error(fmt.Sprintf("Failed to write %s: %v", key, err))
			return nil
		}
		local.Files[key] = sha256sum(content)
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

// resolveFilePaths fetches the file tree for the given branch, falling back to
// master if the version branch has no .ai/ files.
func (r *AgentsInstallCommand) resolveFilePaths(branch string) ([]string, string, error) {
	paths, err := r.treeFetcher(branch)
	if err != nil {
		return nil, branch, err
	}
	if len(paths) == 0 && branch != docsFallbackBranch {
		paths, err = r.treeFetcher(docsFallbackBranch)
		return paths, docsFallbackBranch, err
	}
	return paths, branch, nil
}
