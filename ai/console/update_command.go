package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type AgentsUpdateCommand struct {
	manifestFetcher func(branch string) ([]ManifestEntry, error)
	fetcher         func(branch, path string) ([]byte, error)
}

func NewAgentsUpdateCommand() *AgentsUpdateCommand {
	return &AgentsUpdateCommand{
		manifestFetcher: fetchManifest,
		fetcher:         fetchRaw,
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
			&command.BoolFlag{
				Name:               "all",
				Aliases:            []string{"a"},
				Value:              false,
				Usage:              "Also install new facade files not yet installed",
				DisableDefaultText: true,
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
	entries, err := r.manifestFetcher(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}
	if len(entries) == 0 && branch != docsFallbackBranch {
		entries, err = r.manifestFetcher(docsFallbackBranch)
		if err != nil {
			ctx.Error(err.Error())
			return nil
		}
	}
	if len(entries) == 0 {
		ctx.Error(fmt.Sprintf("No agent files found for version %s. Check https://github.com/goravel/docs", version))
		return nil
	}

	facadeArgs := ctx.Arguments()
	force := ctx.OptionBool("force")

	var toProcess []ManifestEntry
	switch {
	case len(facadeArgs) > 0:
		toProcess = entriesForFacades(entries, facadeArgs)
		if len(toProcess) == 0 {
			ctx.Error(fmt.Sprintf("No agent files found for facade(s): %s", strings.Join(facadeArgs, ", ")))
			return nil
		}
	case ctx.OptionBool("all"):
		toProcess = entries
	default:
		toProcess = installedEntries(entries, local.Files)
	}

	var updated, skippedUserModified, conflicts, alreadyUpToDate int

	for _, entry := range toProcess {
		upstreamContent, fetchErr := r.fetcher(branch, entry.Path)
		if fetchErr != nil {
			ctx.Error(fetchErr.Error())
			return nil
		}
		if upstreamContent == nil {
			ctx.Warning(fmt.Sprintf("File not found upstream: %s", entry.Path))
			continue
		}
		upstreamSHA256 := sha256sum(upstreamContent)

		storedSHA256, exists := local.Files[entry.Path]
		if !exists {
			if writeErr := writeAgentFile(entry.Path, upstreamContent); writeErr != nil {
				ctx.Error(fmt.Sprintf("Failed to write %s: %v", entry.Path, writeErr))
				return nil
			}
			local.Files[entry.Path] = upstreamSHA256
			updated++
			continue
		}

		localPath := destPathFor(entry.Path)
		localContent, readErr := os.ReadFile(localPath)
		if readErr != nil {
			if writeErr := writeAgentFile(entry.Path, upstreamContent); writeErr != nil {
				ctx.Error(fmt.Sprintf("Failed to write %s: %v", entry.Path, writeErr))
				return nil
			}
			local.Files[entry.Path] = upstreamSHA256
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
			ctx.Warning(fmt.Sprintf("Conflict: %s modified locally and changed upstream. Use --force to overwrite.", entry.Path))
			conflicts++
			continue
		}

		if writeErr := writeAgentFile(entry.Path, upstreamContent); writeErr != nil {
			ctx.Error(fmt.Sprintf("Failed to write %s: %v", entry.Path, writeErr))
			return nil
		}
		local.Files[entry.Path] = upstreamSHA256
		updated++
	}

	if err := writeVersionFile(local); err != nil {
		ctx.Error(fmt.Sprintf("Failed to write .version: %v", err))
		return nil
	}

	ctx.Info(fmt.Sprintf("%d updated, %d skipped (user modified), %d conflicts (use --force to overwrite), %d already up to date.", updated, skippedUserModified, conflicts, alreadyUpToDate))
	return nil
}

// installedEntries returns only the entries whose paths are already tracked in the local .version file.
func installedEntries(entries []ManifestEntry, installedFiles map[string]string) []ManifestEntry {
	var out []ManifestEntry
	for _, e := range entries {
		if _, ok := installedFiles[e.Path]; ok {
			out = append(out, e)
		}
	}
	return out
}
