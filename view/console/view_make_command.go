package console

import (
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/file"
)

const (
	DefaultViewPath      = "resources/views"
	DefaultViewExtension = ".tmpl"
)

type ViewMakeCommand struct {
	config config.Config
}

func NewViewMakeCommand(config config.Config) *ViewMakeCommand {
	return &ViewMakeCommand{
		config: config,
	}
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
				Value: DefaultViewPath,
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
		ctx.Error(errors.ConsoleEmptyFieldValue.Args("view name").Error())
		return nil
	}

	viewPath := ctx.Option("path")
	if viewPath == "" {
		viewPath = r.getViewPath()
	}

	// Get the view extension from configuration
	viewExtension := r.getViewExtension()

	// Ensure the view name has the correct extension
	if !strings.HasSuffix(viewName, viewExtension) {
		viewName = viewName + viewExtension
	}

	filePath := filepath.Join(viewPath, viewName)

	// Check if file already exists
	if file.Exists(filePath) && !ctx.OptionBool("force") {
		ctx.Error(errors.ConsoleFileAlreadyExists.Args(filePath).Error())
		return nil
	}

	// get the path name from the view path
	viewPathName := filePath
	if strings.Contains(filePath, "resources") {
		index := strings.Index(filePath, "resources")
		viewPathName = filePath[index:]
	}

	// Create the view file
	stub := r.getStub()
	content := r.populateStub(stub, viewName, viewPathName)

	if err := file.PutContent(filePath, content); err != nil {
		return err
	}

	ctx.Success("View created successfully: " + filePath)

	return nil
}

func (r *ViewMakeCommand) getViewPath() string {
	if r.config != nil {
		customViewPath := r.config.GetString("http.view.path", DefaultViewPath)
		if customViewPath != "" {
			return customViewPath
		}
	}

	return DefaultViewPath
}

func (r *ViewMakeCommand) getStub() string {
	if r.config != nil {
		customStub := r.config.GetString("http.view.stub", "")
		if customStub != "" {
			return customStub
		}
	}

	return Stubs{}.View()
}

// getViewExtension gets the view extension from configuration
func (r *ViewMakeCommand) getViewExtension() string {
	if r.config == nil {
		return DefaultViewExtension
	}

	extension := r.config.GetString("http.view.extension", DefaultViewExtension)
	if extension == "" {
		return DefaultViewExtension
	}

	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	return extension
}

// populateStub Populate the place-holders in the command stub.
func (r *ViewMakeCommand) populateStub(stub string, viewName string, viewPath string) string {
	viewPathDefinition := strings.ReplaceAll(viewPath, DefaultViewPath+"/", "")

	stub = strings.ReplaceAll(stub, "DummyViewName", viewName)
	stub = strings.ReplaceAll(stub, "DummyPathName", viewPath)
	stub = strings.ReplaceAll(stub, "DummyPathDefinition", viewPathDefinition)

	return stub
}
