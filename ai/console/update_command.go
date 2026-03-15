package console

import (
	"fmt"
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type AgentsUpdateCommand struct {
	treeFetcher func(branch string) ([]string, error)
	fetcher     func(branch, path string) ([]byte, error)
}

func NewAgentsUpdateCommand() *AgentsUpdateCommand {
	return &AgentsUpdateCommand{
		treeFetcher: fetchFileTree,
		fetcher:     fetchRaw,
	}
}

func (r *AgentsUpdateCommand) Signature() string {
	return "agents:update"
}

func (r *AgentsUpdateCommand) Description() string {
	return "Update AI agent skill files to match the current Goravel version"
}

func (r *AgentsUpdateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "agents",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:               "force",
				Value:              false,
				Usage:              "Overwrite even if user modified locally",
				DisableDefaultText: true,
			},
			&command.StringFlag{
				Name:  "file",
				Usage: "Update only one file (e.g. route)",
			},
			&command.StringFlag{
				Name:  "version",
				Usage: "Force a specific version (e.g. v1.17)",
			},
		},
	}
}

func (r *AgentsUpdateCommand) Handle(ctx console.Context) error {
	if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
		ctx.Error("No .ai/.version found. Run agents:install first.")
		return nil
	}

	local, err := readVersionFile()
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to read .ai/.version: %v", err))
		return nil
	}

	version := ctx.Option("version")
	if version == "" {
		version, err = detectGoravelVersion()
		if err != nil {
			ctx.Error("Cannot detect Goravel version from go.mod. Use --version to specify it.")
			return nil
		}
	}
	if !isSupportedVersion(version) {
		ctx.Error(fmt.Sprintf("Agent files are only available for Goravel v1.17 and above (got %s).", version))
		return nil
	}

	branch := resolveBranch(version)
	paths, err := r.treeFetcher(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if len(paths) == 0 && branch != docsFallbackBranch {
		branch = docsFallbackBranch
		paths, err = r.treeFetcher(branch)
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}
	if len(paths) == 0 {
		ctx.Error(fmt.Sprintf("No agent files found for version %s. Check https://github.com/goravel/docs", version))
		return nil
	}

	filter := ctx.Option("file")
	force := ctx.OptionBool("force")
	pathsToCheck := filterPaths(paths, filter)

	var updated, skippedUserModified, conflicts, alreadyUpToDate int

	for _, key := range pathsToCheck {
		upstreamContent, fetchErr := r.fetcher(branch, key)
		if fetchErr != nil {
			ctx.Error(fetchErr.Error())
			return nil
		}
		if upstreamContent == nil {
			ctx.Warning(fmt.Sprintf("File not found upstream: %s", key))
			continue
		}
		upstreamSHA256 := sha256sum(upstreamContent)

		storedSHA256, exists := local.Files[key]
		if !exists {
			if writeErr := writeAgentFile(key, upstreamContent); writeErr != nil {
				ctx.Error(fmt.Sprintf("Failed to write %s: %v", key, writeErr))
				return nil
			}
			local.Files[key] = upstreamSHA256
			updated++
			continue
		}

		localPath := destPathFor(key)
		localContent, readErr := os.ReadFile(localPath)
		if readErr != nil {
			if writeErr := writeAgentFile(key, upstreamContent); writeErr != nil {
				ctx.Error(fmt.Sprintf("Failed to write %s: %v", key, writeErr))
				return nil
			}
			local.Files[key] = upstreamSHA256
			updated++
			continue
		}

		localCurrentSHA256 := sha256sum(localContent)
		localModified := localCurrentSHA256 != storedSHA256
		upstreamChanged := upstreamSHA256 != storedSHA256

		if !localModified && !upstreamChanged {
			alreadyUpToDate++
			continue
		}

		if localModified && !upstreamChanged {
			skippedUserModified++
			continue
		}

		if localModified && !force {
			ctx.Warning(fmt.Sprintf("Conflict: %s modified locally and changed upstream. Use --force to overwrite.", key))
			conflicts++
			continue
		}

		if writeErr := writeAgentFile(key, upstreamContent); writeErr != nil {
			ctx.Error(fmt.Sprintf("Failed to write %s: %v", key, writeErr))
			return nil
		}
		local.Files[key] = upstreamSHA256
		updated++
	}

	if err := writeVersionFile(local); err != nil {
		ctx.Error(fmt.Sprintf("Failed to write .version: %v", err))
		return nil
	}

	ctx.Info(fmt.Sprintf("%d updated, %d skipped (user modified), %d conflicts (use --force to overwrite), %d already up to date.", updated, skippedUserModified, conflicts, alreadyUpToDate))
	return nil
}
