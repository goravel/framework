package console

import (
	"os"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/file"
)

type DownCommand struct {
	app foundation.Application
}

func NewDownCommand(app foundation.Application) *DownCommand {
	return &DownCommand{app}
}

// Signature The name and signature of the console command.
func (r *DownCommand) Signature() string {
	return "up"
}

// Description The console command description.
func (r *DownCommand) Description() string {
	return "Bring the application out of maintenance mode"
}

// Extend The console command extend.
func (r *DownCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (r *DownCommand) Handle(ctx console.Context) error {
	path := r.app.StoragePath("framework/down")

	if ok := file.Exists(path); ok {
		ctx.Error("The application is in maintenance mode already!")

		return nil
	}

	os.Create(path)

	ctx.Info("The application is in maintenance mode now")

	return nil
}
