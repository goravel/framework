package console

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/support/file"
)

type ViewMakeCommand struct {
}

// Signature The name and signature of the console command.
func (r *ViewMakeCommand) Signature() string {
	return "make:view"
}

// Description The console command description.
func (r *ViewMakeCommand) Description() string {
	return "Create a new view file"
}

// Extend The console command extend.
func (r *ViewMakeCommand) Extend() command.Extend {
	return command.Extend{
		Category: "make",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "path",
				Value: "resources/views",
				Usage: "The path where the view file should be created",
			},
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Create the view even if it already exists",
			},
		},
	}
}

// Handle Execute the console command.
func (r *ViewMakeCommand) Handle(ctx console.Context) error {
	viewName := ctx.Argument(0)
	if viewName == "" {
		ctx.Error("View name is required")
		return nil
	}

	viewPath := ctx.Option("path")
	if viewPath == "" {
		viewPath = "resources/views"
	}

	// Ensure the view name has the correct extension
	if !strings.HasSuffix(viewName, ".tmpl") {
		viewName = viewName + ".tmpl"
	}

	filePath := filepath.Join(viewPath, viewName)

	// Check if file already exists
	if file.Exists(filePath) && !ctx.OptionBool("force") {
		ctx.Error("View already exists. Use --force to overwrite.")
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if !file.Exists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			ctx.Error("Failed to create directory: " + err.Error())
			return nil
		}
	}

	// get the path name from the view path
	viewPathName := filePath
	if strings.Contains(filePath, "resources") {
		index := strings.Index(filePath, "resources")
		if index != -1 {
			viewPathName = filePath[index:]
		}
	}

	// Create the view file
	stub := r.getStub()
	content := r.populateStub(stub, viewName, viewPathName)

	if err := file.PutContent(filePath, content); err != nil {
		ctx.Error("Failed to create view: " + err.Error())
		return nil
	}

	ctx.Success("View created successfully: " + filePath)

	return nil
}

func (r *ViewMakeCommand) getStub() string {
	return Stubs{}.View()
}

// populateStub Populate the place-holders in the command stub.
func (r *ViewMakeCommand) populateStub(stub string, viewName string, viewPath string) string {
	viewPathDefinition := strings.ReplaceAll(viewPath, "resources/views/", "")

	stub = strings.ReplaceAll(stub, "DummyViewName", viewName)
	stub = strings.ReplaceAll(stub, "DummyPathName", viewPath)
	stub = strings.ReplaceAll(stub, "DummyPathDefinition", viewPathDefinition)

	return stub
}
