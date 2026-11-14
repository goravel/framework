// Package console holds the module Artisan commands. It never imports foundation.
package console

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/modular"
	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

const (
	statusCreated = "<fg=green;op=bold>CREATED</>"
	statusExists  = "<fg=yellow;op=bold>EXISTS</>"
	statusManual  = "<fg=blue;op=bold>MANUAL</>"

	statusPublished    = "<fg=green;op=bold>PUBLISHED</>"
	statusUnchanged    = "<fg=yellow;op=bold>UNCHANGED</>"
	statusFailed       = "<fg=red;op=bold>FAILED</>"
	statusSkippedLocal = "<fg=gray>SKIPPED (local)</>"
	statusSkippedNone  = "<fg=gray>SKIPPED (nothing to publish)</>"
)

// modulesDirectory is modular.path, also before config/modular.go exists.
func modulesDirectory(config config.Config) string {
	if config == nil {
		return modular.DefaultDirectory
	}

	return config.GetString("modular.path", modular.DefaultDirectory)
}

// facadesImport returns the app's facades package, or the framework's when the app has none.
func facadesImport() (importPath, pkg string) {
	if file.Exists(path.Facade()) {
		facades := packages.Paths().Facades()

		return facades.Import(), facades.Package()
	}

	return "github.com/goravel/framework/facades", "facades"
}

// putIfMissing writes content to file unless it exists and prints a row named label.
func putIfMissing(ctx console.Context, label, filePath, content string) error {
	if file.Exists(filePath) {
		ctx.TwoColumnDetail(label, statusExists)

		return nil
	}
	if err := file.PutContent(filePath, content); err != nil {
		return err
	}
	ctx.TwoColumnDetail(label, statusCreated)

	return nil
}
