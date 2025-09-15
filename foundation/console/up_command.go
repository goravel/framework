package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/file"
)

type UpCommand struct {
	app foundation.Application
}

func NewUpCommand(app foundation.Application) *UpCommand {
	return &UpCommand{app}
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
	path := r.app.StoragePath("framework/maintenance")
	if ok := file.Exists(path); ok {
		if err := file.Remove(path); err != nil {
			return err
		}

		ctx.Info("The application is up and live now")

		return nil
	}

	ctx.Error("The application is not in maintenance mode")

	return nil
}
