package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/foundation"
)

type UpCommand struct {
	app     foundation.Application
	storage filesystem.Storage
}

func NewUpCommand(app foundation.Application) *UpCommand {
	return &UpCommand{app, app.MakeStorage()}
}

// Signature The name and signature of the console command.
func (r *UpCommand) Signature() string {
	return "up"
}

// Description The console command description.
func (r *UpCommand) Description() string {
	return "Bring the application out of maintenance mode"
}

// Extend The console command extend.
func (r *UpCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (r *UpCommand) Handle(ctx console.Context) error {
	path := "framework/maintenance.json"
	if ok := r.storage.Exists(path); ok {
		if err := r.storage.Delete(path); err != nil {
			return err
		}

		ctx.Success("The application is up and live now")

		return nil
	}

	ctx.Error("The application is not in maintenance mode")

	return nil
}
