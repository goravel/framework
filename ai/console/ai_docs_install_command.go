package console

import (
	"fmt"
	"os"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type AiDocsInstallCommand struct {
	manifestFetcher func(branch string) ([]ManifestEntry, error)
	fetcher         func(branch, path string) ([]byte, error)
	versionDetector func() (string, error)
}

func NewAiDocsInstallCommand() *AiDocsInstallCommand {
	return &AiDocsInstallCommand{
		manifestFetcher: fetchManifest,
		fetcher:         fetchRaw,
		versionDetector: detectGoravelVersion,
	}
}

func (r *AiDocsInstallCommand) Signature() string {
	return "ai:docs:install"
}

func (r *AiDocsInstallCommand) Description() string {
	return "Install AI documentation and skill files for Goravel (e.g., 'artisan ai:docs:install Auth Route')"
}

func (r *AiDocsInstallCommand) Extend() command.Extend {
	return command.Extend{
		Category: "ai",
		Flags: []command.Flag{
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

func (r *AiDocsInstallCommand) Handle(ctx console.Context) error {
	version, err := r.versionDetector()
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to detect version: %v", err))
		return nil
	}

	if !isSupportedVersion(version) {
		ctx.Error(fmt.Sprintf("AI docs are only available for Goravel v1.17 and above (got %s)", version))
		return nil
	}

	branch := resolveBranch(version)
	entries, err := r.manifestFetcher(branch)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if len(entries) == 0 {
		ctx.Error(fmt.Sprintf("No AI docs found for version %s. Check https://github.com/goravel/docs", version))
		return nil
	}

	if !ctx.OptionBool("force") {
		if _, statErr := os.Stat(versionFilePath); statErr == nil {
			if !ctx.Confirm("AI docs are already installed. Overwrite?") {
				ctx.Warning("Cancelled.")
				return nil
			}
		}
	}

	toInstall := r.determineFilesToInstall(ctx, entries)
	if len(toInstall) == 0 {
		return nil
	}

	downloaded, err := downloadFiles(branch, toInstall, r.fetcher)
	if err != nil {
		ctx.Error(err.Error())
		return nil
	}

	if err := saveFiles(version, downloaded); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Info(fmt.Sprintf("Installed %d file(s) for version %s.", len(downloaded), version))
	return nil
}

// determineFilesToInstall checks how the command was called to figure out what to download.
// Order of precedence: --all flag -> specific arguments (e.g., Auth) -> defaults.
func (r *AiDocsInstallCommand) determineFilesToInstall(ctx console.Context, entries []ManifestEntry) []ManifestEntry {
	if ctx.OptionBool("all") {
		return entries
	}

	// This allows an AI agent to run `artisan ai:docs:install Auth Route`
	facadeArgs := ctx.Arguments()
	if len(facadeArgs) > 0 {
		toInstall := entriesForFacades(entries, facadeArgs)
		if len(toInstall) == 0 {
			ctx.Error(fmt.Sprintf("No AI docs found for facade(s): %s", strings.Join(facadeArgs, ", ")))
		}
		return toInstall
	}

	return defaultEntries(entries)
}
