package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
)

type AiDocsUpdateCommand struct {
	manifestFetcher func(branch string) ([]ManifestEntry, error)
	fetcher         func(branch, path string) ([]byte, error)
	versionDetector func() (string, error)
}

func NewAiDocsUpdateCommand() *AiDocsUpdateCommand {
	return &AiDocsUpdateCommand{
		manifestFetcher: fetchManifest,
		fetcher:         fetchRaw,
		versionDetector: detectGoravelVersion,
	}
}

func (r *AiDocsUpdateCommand) Signature() string {
	return "ai:docs:update"
}

func (r *AiDocsUpdateCommand) Description() string {
	return "Update installed AI documentation files to match the current Goravel version"
}

func (r *AiDocsUpdateCommand) Extend() command.Extend {
	return command.Extend{
		Category: "ai",
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
		},
	}
}

func (r *AiDocsUpdateCommand) Handle(ctx console.Context) error {
	local, err := readVersionFile()
	if err != nil || local.Version == "" {
		ctx.Error("No .ai/.version found. Run 'artisan ai:docs:install' first.")
		return nil
	}

	branch := resolveBranch(local.Version)
	entries, err := r.manifestFetcher(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if len(entries) == 0 {
		ctx.Error(fmt.Sprintf("No AI docs found for version %s. Check https://github.com/goravel/docs", local.Version))
		return nil
	}

	toProcess := r.determineFilesToProcess(ctx, entries, local.Files)
	if len(toProcess) == 0 {
		return nil
	}

	var updated, skipped, conflicts, upToDate int
	force := ctx.OptionBool("force")

	for _, entry := range toProcess {
		upstreamContent, err := r.fetcher(branch, entry.Path)
		if err != nil || upstreamContent == nil {
			ctx.Warning(fmt.Sprintf("File not found upstream: %s", entry.Path))
			continue
		}

		upstreamSHA := sha256sum(upstreamContent)
		storedSHA, exists := local.Files[entry.Path]

		// New file being added via --all or specific facade args
		if !exists {
			if err := writeAgentFile(entry.Path, upstreamContent); err == nil {
				local.Files[entry.Path] = upstreamSHA
				updated++
			}
			continue
		}

		localContent, err := os.ReadFile(destPathFor(entry.Path))
		if err != nil {
			// File went missing locally, restore it
			if err := writeAgentFile(entry.Path, upstreamContent); err == nil {
				local.Files[entry.Path] = upstreamSHA
				updated++
			}
			continue
		}

		localCurrentSHA := sha256sum(localContent)
		localModified := localCurrentSHA != storedSHA
		upstreamChanged := upstreamSHA != storedSHA

		if !localModified && !upstreamChanged {
			upToDate++
			continue
		}
		if localModified && !upstreamChanged {
			skipped++
			continue
		}
		if localModified && !force {
			ctx.Warning(fmt.Sprintf("Conflict: %s modified locally and changed upstream. Use --force to overwrite.", entry.Path))
			conflicts++
			continue
		}

		if err := writeAgentFile(entry.Path, upstreamContent); err == nil {
			local.Files[entry.Path] = upstreamSHA
			updated++
		}
	}

	if err := writeVersionFile(local); err != nil {
		ctx.Error(fmt.Sprintf("Failed to write .version: %v", err))
		return nil
	}

	ctx.Info(fmt.Sprintf("%d updated, %d skipped (user modified), %d conflicts (use --force), %d up to date.", updated, skipped, conflicts, upToDate))
	return nil
}

func (r *AiDocsUpdateCommand) determineFilesToProcess(ctx console.Context, entries []ManifestEntry, installedFiles map[string]string) []ManifestEntry {
	facadeArgs := ctx.Arguments()

	// If the AI agent explicitly requests 'artisan ai:docs:update Auth', only update/install Auth
	if len(facadeArgs) > 0 {
		toProcess := entriesForFacades(entries, facadeArgs)
		if len(toProcess) == 0 {
			ctx.Error(errors.AiDocsFacadeNotFound.Args(strings.Join(facadeArgs, ", ")).Error())
		}
		return toProcess
	}

	if ctx.OptionBool("all") {
		return entries
	}

	return installedEntries(entries, installedFiles)
}
