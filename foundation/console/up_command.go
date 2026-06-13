package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type UpCommand struct {
	maintenance *MaintenanceMode
}

func NewUpCommand(maintenance *MaintenanceMode) *UpCommand {
	return &UpCommand{maintenance}
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
	deleted, err := r.maintenance.Delete()
	if err != nil {
		return err
	}
	if !deleted {
		ctx.Error("The application is not in maintenance mode")
		return nil
	}

	ctx.Success("The application is up and live now")

	return nil
}
