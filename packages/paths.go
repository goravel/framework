package packages

import (
	"path"
	"runtime/debug"
	"strings"

	"github.com/goravel/framework/contracts/packages"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/path/internals"
)

type Paths struct {
	main string
}

func NewPaths(main string) *Paths {
	return &Paths{main: main}
}

// Bootstrap returns the bootstrap package path, eg: goravel/bootstrap.
func (r *Paths) Bootstrap() packages.Path {
	return NewPath(support.Config.Paths.Bootstrap, r.main, false)
}

// Config returns the config package path, eg: goravel/config.
func (r *Paths) Config() packages.Path {
	return NewPath(support.Config.Paths.Config, r.main, false)
}

// Facades returns the facades package path, eg: goravel/app/facades.
func (r *Paths) Facades() packages.Path {
	return NewPath(support.Config.Paths.Facades, r.main, false)
}

// Main returns the main package path, eg: github.com/goravel/goravel.
func (r *Paths) Main() packages.Path {
	return NewPath(r.main, r.main, false)
}

// Module returns the module path of the package, eg: github.com/goravel/framework/auth.
func (r *Paths) Module() packages.Path {
	var p string
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasSuffix(info.Path, "setup") {
		p = path.Dir(info.Path)
	}

	return NewPath(p, r.main, true)
}

// Routes returns the routes package path, eg: goravel/routes.
func (r *Paths) Routes() packages.Path {
	return NewPath(support.Config.Paths.Routes, r.main, false)
}

// Tests returns the tests package path, eg: goravel/tests.
func (r *Paths) Tests() packages.Path {
	return NewPath(support.Config.Paths.Tests, r.main, false)
}

type Path struct {
	main     string
	path     string
	isModule bool
}

func NewPath(path, main string, isModule bool) *Path {
	return &Path{path: path, main: main, isModule: isModule}
}

// Package returns the sub-package name, or the main package name if no sub-package path is specified.
// For example, if r.path is "app/http/controllers", it returns "controllers".
// If r.path is empty, it returns the last component of r.main.
func (r *Path) Package() string {
	p := pkg(r.path)

	if p == "" {
		return pkg(r.main)
	}

	return p
}

// Import returns the sub-package import path, or the main package import path if no sub-package path is specified.
// For example, if r.path is "app/http/controllers" and r.main is "github.com/goravel/goravel",
// it returns "goravel/app/http/controllers". If r.path is empty, it returns "goravel".
// The path will be returned directly if it starts with "github.com/goravel/framework/", given it's a framework sub-package.
func (r *Path) Import() string {
	mainSlice := internals.ToSlice(r.main)
	mainImport := mainSlice[len(mainSlice)-1]

	if r.path != "" {
		if r.isModule {
			return r.path
		}

		pathSlice := internals.ToSlice(r.path)
		importSlice := append([]string{mainImport}, pathSlice...)

		return strings.Join(importSlice, "/")
	}

	return mainImport
}

// pkg extracts the last component of a file path string.
// For example, "app/http/controllers" returns "controllers".
func pkg(path string) string {
	s := internals.ToSlice(path)

	if len(s) == 0 {
		return ""
	}

	return s[len(s)-1]
}
